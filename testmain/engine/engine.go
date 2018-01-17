package engine

import (
	"context"
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"git.oschina.net/kuaishangtong/common/utils/log"
	"git.oschina.net/kuaishangtong/navi/lb"
	"git.oschina.net/kuaishangtong/navi/registry"
	"git.oschina.net/kuaishangtong/navi/testmain/navi"
	"github.com/valyala/fastrand"
	"sync"
	"time"
)

var XEngine *Engine // 全局引擎

var gTimeout int
var gMaxConns int

type Conn struct {
	host     string
	interval int
	*navi.NaviServiceClient
	closed    bool
	available bool
}

func newConn(host string, interval int) (*Conn, error) {
	var err error
	conn := &Conn{
		host:      host,
		interval:  interval,
		closed:    false,
		available: true,
	}

	conn.NaviServiceClient, err = conn.connect()
	if err != nil {
		return nil, err
	}

	go conn.check()

	return conn, nil
}

func (c *Conn) check() {
	ticker := time.NewTicker(time.Duration(c.interval) * time.Second)
	defer ticker.Stop()

	for !c.closed {
		select {
		case <-ticker.C:
			_, err := c.Ping()
			if err != nil {
				log.Error(err)
				c.available = false

				newClient, err1 := c.connect()
				if err1 != nil {
					log.Error(err)
					continue
				} else {
					c.close()
					c.NaviServiceClient = newClient
					c.available = true
				}
			} else {
				c.available = true
			}
		}
	}
}

func (c *Conn) connect() (*navi.NaviServiceClient, error) {
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()

	transport, err := thrift.NewTSocketTimeout(c.host, time.Duration(gTimeout)*time.Second)
	if err != nil {
		return nil, err
	}

	useTransport, err := transportFactory.GetTransport(transport)
	if err != nil {
		transport.Close()
		return nil, err
	}

	client := navi.NewNaviServiceClientFactory(useTransport, protocolFactory)
	if err := transport.Open(); err != nil {
		transport.Close()
		useTransport.Close()
		return nil, err
	}

	return client, nil
}

func (c *Conn) Available() bool {
	return c.available
}

func (c *Conn) close() {
	if c.NaviServiceClient != nil {
		c.NaviServiceClient.InputProtocol.Transport().Close()
		c.NaviServiceClient.OutputProtocol.Transport().Close()
		c.NaviServiceClient.Transport.Close()
	}
}

func (c *Conn) Close() {
	c.closed = true
	c.close()
}

type ServerHost struct {
	host  string
	lock  *sync.RWMutex
	conns []*Conn
}

func newServerHost(host string, maxConns int) (*ServerHost, error) {
	serverHost := &ServerHost{
		host: host,
		lock: new(sync.RWMutex),
	}
	conns := make([]*Conn, maxConns, maxConns)
	for i := 0; i < maxConns; i++ {
		conn, err := newConn(host, 1)
		if err != nil {
			return nil, err
		}

		conns[i] = conn
	}

	serverHost.conns = conns
	return serverHost, nil
}

func (s *ServerHost) closeAllConns() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.conns != nil {
		for _, conn := range s.conns {
			conn.Close()
		}
		s.conns = s.conns[:0]
	}
}

func (s *ServerHost) getConn() *Conn {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.conns != nil && len(s.conns) > 0 {
		length := len(s.conns)
		i := int(fastrand.Uint32n(uint32(length)))

		loopCount := 0
		for !s.conns[i].Available() {
			loopCount++
			if loopCount == length {
				return nil
			}
			i = (i + 1) % length
		}

		return s.conns[i]
	}
	return nil
}

type Engine struct {
	servers   map[string]*ServerHost
	discovery registry.ServiceDiscovery
	selector  lb.Selector
}

func InitEngine(basePath string, servicePath string, zkhosts []string, timeout, maxConns int) (err error) {
	gTimeout = timeout
	gMaxConns = maxConns
	XEngine = &Engine{
		servers: make(map[string]*ServerHost),
	}

	XEngine.discovery, err = registry.NewZookeeperDiscovery(basePath, servicePath, zkhosts, nil)
	if err != nil {
		return
	}

	selecter := lb.NewSelector(lb.RoundRobin, nil)
	selecter.UpdateServer(XEngine.getServices())

	XEngine.selector = selecter

	return
}

func (c *Engine) serviceDiscovery() {
	ch := c.discovery.WatchService()

	for pairs := range ch {
		newservers := make(map[string]string)
		for _, p := range pairs {
			newservers[p.Key] = p.Value
		}

		c.selector.UpdateServer(newservers)

		var oldmap, newmap map[string]struct{} = make(map[string]struct{}), make(map[string]struct{})
		for k, _ := range c.servers {
			oldmap[k] = struct{}{}
		}

		for k, _ := range newservers {
			newmap[k] = struct{}{}
		}

		for k, _ := range newmap {
			if _, ok := oldmap[k]; ok {
				delete(oldmap, k)
				delete(newmap, k)
			}
		}

		for key, _ := range oldmap {
			if host, ok := c.servers[key]; ok {
				host.closeAllConns()
			}
			delete(c.servers, key)
		}

		for key, _ := range newmap {
			host, err := newServerHost(key, gMaxConns)
			if err != nil {
				log.Error(err)
				continue
			}
			c.servers[key] = host
		}
	}
}

func (c *Engine) getServices() map[string]string {
	kvpairs := c.discovery.GetServices()
	servers := make(map[string]string)
	for _, p := range kvpairs {
		servers[p.Key] = p.Value
	}
	return servers
}

func (c *Engine) GetConn() (*Conn, error) {
	host := c.selector.Select(context.Background(), "", "", nil)
	if host, ok := c.servers[host]; ok {
		conn := host.getConn()
		if conn == nil {
			return nil, fmt.Errorf("can not find available conn in %s", host)
		}
		return conn, nil
	}

	return nil, fmt.Errorf("can not find available conn")
}
