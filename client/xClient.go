package client

import (
	"context"
	"errors"
	"sync"
	"strings"
	"fmt"
	"github.com/smallnest/rpcx/share"
)

// XClient is an interface that used by client with service discovery and service governance.
// One XClient is used only for one service. You should create multiple XClient for multiple services.
type XClient interface {
	Call(ctx context.Context, serviceMethod string, args interface{}, reply interface{}) error
	Close() error
}

type xClient struct {
	failMode 		FailMode
	selectMode		SelectMode
	servers 		map[string]string
	cachedClient	map[string]RPCClient
	servicePath 	string
	serviceMethod	string
	option			Option

	mu				sync.RWMutex
	selector 		Selector

	isShutdown bool
}

// NewXClient creates a XClient that supports service discovery and service governance
func NewXClient(servicePath string, failMode FailMode, selectMode SelectMode, option Option) XClient {
	client := &xClient{
		failMode:		failMode,
		selectMode:		selectMode,
		servicePath:	servicePath,
		cachedClient:	make(map[string]RPCClient),
		option:			option,
	}

	servers := make(map[string]string)
	//TODO
	//服务发现
	//pairs := discovery.GetServices()
	//for _, p := range pairs {
	//	servers[p.Key] = p.Value
	//}
	client.servers = servers
	if selectMode != Closest {
		client.selector = newSelector(selectMode, servers)
	}

	//TODO
	//服务状态监控

	return client
}

type ServiceError string

func (e ServiceError) Error() string {
	return string(e)
}

// Go invokes the function asynchronously. It returns the Call structure representing the invocation. The done channel will signal when the call is complete by returning the same Call object. If done is nil, Go will allocate a new channel. If non-nil, done must be buffered or Go will deliberately crash.
// It does not use FailMode.
func (c *xClient) Go(ctx context.Context, serviceMethod string, args interface{}, reply interface{}, done chan *Call) (*Call, error) {
	if c.isShutdown {
		return nil, errors.New("xClient is shutdown")
	}

	//if c.auth != "" {
	//	metadata := ctx.Value(share.ReqMetaDataKey)
	//	if metadata == nil {
	//		return nil, errors.New("must set ReqMetaDataKey in context")
	//	}
	//	m := metadata.(map[string]string)
	//	m[share.AuthKey] = c.auth
	//}

	_, client, err := c.selectClient(ctx, c.servicePath, serviceMethod, args)
	if err != nil {
		return nil, err
	}
	return client.Go(ctx, c.servicePath, serviceMethod, args, reply, done), nil
}

// Call invokes the named function, waits for it to complete, and returns its error status.
// It handles errors base on FailMode.
func (c *xClient) Call(ctx context.Context, serviceMethod string, args interface{}, reply interface{}) error {
	//if c.isShutdown {
	//	return ErrXClientShutDown
	//}

	var err error
	k, client, err := c.selectClient(ctx, c.servicePath, serviceMethod, args)
	if err != nil {
		if c.failMode == Failfast {
			return err
		}

		if _, ok := err.(ServiceError); ok {
			return err
		}
	}

	//容灾处理
	switch c.failMode {
	case Failtry:
		retries := c.option.Retries
		for retries > 0 {
			retries--
			err := client.Call(ctx, c.servicePath, serviceMethod, args, reply)
			if err == nil {
				return nil
			}
			if _, ok := err.(ServiceError); ok {
				return err
			}

			c.removeClient(k, client)
			client, _ = c.getCachedClient(k)
		}
		return err
	case Failover:
		retries := c.option.Retries
		for retries > 0 {
			retries--
			err = client.Call(ctx, c.servicePath, serviceMethod, args, reply)
			if err == nil {
				return nil
			}
			if _, ok := err.(ServiceError); ok {
				return err
			}

			c.removeClient(k, client)
			//select another server
			k, client, _ = c.selectClient(ctx, c.servicePath, serviceMethod, args)
		}

		return err

	default: //Failfast
		err = client.Call(ctx, c.servicePath, serviceMethod, args, reply)
		if err != nil {
			if _, ok := err.(ServiceError); !ok {
				c.removeClient(k, client)
			}
		}

		return err
	}
}

func (c *xClient) selectClient(ctx context.Context, servicePath, serviceMethod string, args interface{}) (string, RPCClient, error) {
	k := c.selector.Select(ctx, servicePath, serviceMethod, args)
	if k == "" {
		return "", nil, errors.New("can not found any server")
	}

	client, err := c.getCachedClient(k)
	return k, client, err
}

func (c *xClient) getCachedClient(k string) (RPCClient, error) {
	c.mu.RLock()
	client := c.cachedClient[k]
	//if client != nil {
	//
	//}
	c.mu.RUnlock()

	//double check
	c.mu.Lock()
	client = c.cachedClient[k]
	if client == nil {
		network, addr := splitNetworkAndAddress(k)
		//TODO
		//创建连接
		//client = &Client{
		//	option:		c.option,
		//}

		c.cachedClient[k] = client
	}
	c.mu.Unlock()

	return client, nil
}

func splitNetworkAndAddress(server string) (string, string) {
	ss := strings.SplitN(server, "@", 2)
	if len(ss) == 1{
		return "tcp",server
	}

	return ss[0], ss[1]
}

func (c *xClient) removeClient(k string, client RPCClient) {
	c.mu.Lock()
	cl := c.cachedClient[k]
	if cl == client {
		delete(c.cachedClient, k)
	}
	c.mu.Unlock()

	if client != nil {
		client.Close()
	}
}

// Close closes this client and its underlying connnections to services.
func (c *xClient) Close() error {
	//c.isShutdown = true

	var errs []error
	c.mu.Lock()
	for k, v := range c.cachedClient {
		e := v.Close()
		if e != nil {
			errs = append(errs, e)
		}

		delete(c.cachedClient, k)
	}
	c.mu.Unlock()

	//go func() {
	//	defer func() {
	//		if r := recover(); r != nil {
	//
	//		}
	//	}()

		//c.discovery.RemoveWarcher(c.ch)
		//close(c.ch)
	//}()

	if len(errs) > 0 {
		return fmt.Errorf("%v", errs)
	}
	return nil
}