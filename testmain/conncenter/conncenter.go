package conncenter

import (
	"context"
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"kuaishangtong/common/utils/log"
	"kuaishangtong/navi/lb"
	"kuaishangtong/navi/registry"
	"kuaishangtong/navi/testmain/navi"
	"sync"
	"time"
)

var connCenter *ConnCenter // 全局引擎

var gTimeout int
var gAllocSize int = 8

type Conn struct {
	scpool   *ServerConnPool
	host     string
	interval int
	*navi.NaviServiceClient
	closed    bool
	available bool

	reConnFlag chan struct{}
}

func newConn(pool *ServerConnPool) (*Conn, error) {
	var err error
	conn := &Conn{
		scpool:     pool,
		host:       pool.host,
		interval:   pool.interval,
		closed:     false,
		available:  true,
		reConnFlag: make(chan struct{}),
	}

	conn.NaviServiceClient, err = conn.connect()
	if err != nil {
		return nil, err
	}

	go conn.check()
	return conn, nil
}

func (c *Conn) check() {

	for !c.closed {
		select {
		case <-c.reConnFlag:
			c.close()

			newClient, err := c.connect()
			if err != nil {
				log.Errorf("reconnect to %s err: %v", c.host, err)

				reconnectOK := false
				for !reconnectOK && !c.closed {
					ticker := time.NewTicker(1 * time.Second)

					select {
					case <-ticker.C:
						newClient, err := c.connect()
						if err != nil {
							log.Errorf("reconnect to %s err: %v", c.host, err)
						} else {
							c.NaviServiceClient = newClient
							c.available = true
							reconnectOK = true
						}
					}
					ticker.Stop()
				}
			} else {
				c.NaviServiceClient = newClient
				c.available = true
			}
		}
	}
}

func (c *Conn) GetServerConnPool() *ServerConnPool {
	return c.scpool
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

// 当外部调用rpc接口报错err，一般为rpc连接出问题，需要重新连接
func (c *Conn) Reconnect() {
	c.available = false
	c.reConnFlag <- struct{}{}
}

// 关闭Conn下的thrift rpc连接
func (c *Conn) close() {
	c.available = false
	if c.NaviServiceClient != nil {
		c.NaviServiceClient.InputProtocol.Transport().Close()
		c.NaviServiceClient.OutputProtocol.Transport().Close()
		c.NaviServiceClient.Transport.Close()
	}
}

// 关闭Conn连接
func (c *Conn) Close() {
	c.closed = true
	c.close()
	//close(c.reConnFlag)
}

type ServerConnPool struct {
	lock      *sync.RWMutex
	free      []*Conn
	nextAlloc int

	closed    bool
	available bool
	host      string
	interval  int
}

func newServerConnPool(host string, interval int) (*ServerConnPool, error) {
	shpool := &ServerConnPool{
		free:      make([]*Conn, 0, gAllocSize),
		nextAlloc: gAllocSize,
		lock:      &sync.RWMutex{},
		closed:    false,
		available: true,
		host:      host,
		interval:  interval,
	}

	go shpool.stat()

	return shpool, nil
}

func (s *ServerConnPool) Available() bool {
	return s.available
}

func (s *ServerConnPool) SetAvailable(flag bool) {
	s.available = flag
}

func (s *ServerConnPool) grow() error {
	conns := make([]*Conn, s.nextAlloc, s.nextAlloc)

	for i := 0; i < len(conns); i++ {
		conn, err := newConn(s)
		if err != nil {
			log.Errorf("ServerConnPool newConn to %s err: %v", s.host, err)
		}
		conns[i] = conn
		//log.Infof("new conn to %s", s.host)
	}

	for _, conn := range conns {
		s.free = append(s.free, conn)
	}

	s.nextAlloc *= 2
	return nil
}

func (s *ServerConnPool) close() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.closed = true
	s.closeAllConns()
}

func (s *ServerConnPool) closeAllConns() {
	//s.lock.Lock()
	//defer s.lock.Unlock()

	if s.free != nil {
		for _, conn := range s.free {
			conn.Close()
		}
		s.free = s.free[:0]
	}
}

func (s *ServerConnPool) getConn() (conn *Conn) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.free) == 0 {
		s.grow()
	}

	connOK := false

	index := len(s.free) - 1

	for !connOK && index >= 0 {

		conn, s.free = s.free[index], s.free[:index]
		if conn != nil && conn.Available() {
			connOK = true
		} else {
			if conn != nil {
				conn.Close()
			}
			index = len(s.free) - 1
		}
	}

	if !connOK || index < 0 {
		return nil
	}

	return conn
}

func (s *ServerConnPool) putConn(conn *Conn) {
	// 回收的连接不可用，直接关闭并丢弃
	if !conn.Available() {
		conn.Close()
		return
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.free = append(s.free, conn)
}

func (s *ServerConnPool) reset(size int) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.nextAlloc = size

	for i := size - 1; i < len(s.free); i++ {
		if conn := s.free[i]; conn != nil {
			conn.Close()
		}
	}

	s.free = s.free[:size]
	return nil
}

func (s *ServerConnPool) stat() {

	ticker10 := time.NewTicker(10 * time.Second)
	defer ticker10.Stop()

	var count int

	for !s.closed {
		select {

		case <-ticker10.C:

			length := len(s.free)
			capab := cap(s.free)
			log.Infof("STAT ServerConnPool [%s]: count of conn %d", s.host, length)

			if capab <= gAllocSize {
				continue
			}

			if length > gAllocSize*2 {
				count++
			} else {
				count = 0
			}

			// 5分钟， 空闲的连接太多，连接池重置，连接数量减半
			if count >= 30 {
				s.reset(length / 2)
				count = 0
			}
		}
	}
}

type ConnCenter struct {
	serverPools map[string]*ServerConnPool
	discovery   registry.ServiceDiscovery
	selector    lb.Selector
	failMode    lb.FailMode

	interval int
	retries  int

	closed bool

	lock *sync.RWMutex
}

func InitConnCenter(basePath string, servicePath string, zkhosts []string, timeout, interval, connNum int, failMode lb.FailMode) (err error) {
	gTimeout = timeout
	gAllocSize = connNum

	connCenter = &ConnCenter{
		serverPools: make(map[string]*ServerConnPool),
		failMode:    failMode,
		interval:    interval,
		lock:        &sync.RWMutex{},
		retries:     3,
		closed:      false,
	}

	connCenter.discovery, err = registry.NewZookeeperDiscovery(basePath, servicePath, zkhosts, nil)
	if err != nil {
		return
	}

	log.Infof("ConnCenter NewZookeeperDiscovery in %s/%s", basePath, servicePath)

	initserver := connCenter.getServices()
	selecter := lb.NewSelector(lb.RoundRobin, initserver)

	for host, _ := range initserver {
		pool, err := newServerConnPool(host, interval)
		if err != nil {
			log.Error(err)
			return err
		}
		connCenter.serverPools[host] = pool
		log.Infof("ConnCenter add conn pool into %s", host)
	}

	selecter.UpdateServer(initserver)
	connCenter.selector = selecter

	go connCenter.serviceDiscovery()

	return
}

func (c *ConnCenter) Close() {
	c.closed = true
	for _, connPool := range c.serverPools {
		connPool.close()
	}
}

func (c *ConnCenter) serviceDiscovery() {

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for !c.closed {
		select {
		case pairs, ok := <-c.discovery.WatchService():

			if !ok {
				continue
			}

			newservers := make(map[string]string)
			servers := make([]string, 0, len(newservers))
			for _, p := range pairs {
				newservers[p.Key] = p.Value
				servers = append(servers, p.Key)
			}

			c.selector.UpdateServer(newservers)
			log.Infof("ConnCenter UpdateServer %v", servers)

			var oldmap, newmap map[string]struct{} = make(map[string]struct{}), make(map[string]struct{})
			for k, _ := range c.serverPools {
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

			// 关闭不可用的server对应的连接池
			for host, _ := range oldmap {
				if pool, ok := c.serverPools[host]; ok {
					pool.close()
				}
				delete(c.serverPools, host)
			}

			// 创建新的server对应的连接池
			for host, _ := range newmap {
				pool, err := newServerConnPool(host, c.interval)
				if err != nil {
					log.Error(err)
					continue
				}
				c.serverPools[host] = pool
			}
		}
	}

}

func (c *ConnCenter) getServices() map[string]string {
	kvpairs := c.discovery.GetServices()
	servers := make(map[string]string)
	for _, p := range kvpairs {
		servers[p.Key] = p.Value
	}
	return servers
}

func GetConn() (interface{}, error) {
	return connCenter.getConn()
}

func (c *ConnCenter) getConn() (*Conn, error) {

	var host string
	var serverCount, index int = len(c.serverPools), 0
	for {
		index++
		host = c.selector.Select(context.Background(), "", "", host, nil)
		if host == "" {
			return nil, fmt.Errorf("can not find available serverConnPool")
		}

		if c.serverPools[host].Available() {
			break
		}

		if serverCount == index {
			host = ""
			break
		}
	}

	if host == "" {
		log.Errorf("can not find available serverConnPool")
		return nil, fmt.Errorf("can not find available serverConnPool")
	}

	c.lock.Lock()
	if pool, ok := c.serverPools[host]; ok {
		conn := pool.getConn()
		if conn == nil {
			log.Errorf("can not find available conn in %s", host)
			c.lock.Unlock()
			return nil, fmt.Errorf("can not find available conn in %s", host)
		}
		c.lock.Unlock()
		return conn, nil
	}
	c.lock.Unlock()

	log.Errorf("can not find available conn in %s", host)
	return nil, fmt.Errorf("can not find available conn in %s", host)
}

func PutConn(conn interface{}) error {
	return connCenter.putConn(conn.(*Conn))
}

func (c *ConnCenter) putConn(conn *Conn) error {

	c.lock.Lock()
	if pool, ok := c.serverPools[conn.host]; ok {
		pool.putConn(conn)
	}
	c.lock.Unlock()

	//if conn.scpool != nil {
	//	conn.scpool.close()
	//}
	conn.Close()
	return fmt.Errorf("can not find serverConnPool %s", conn.host)
}

func (c *ConnCenter) GetFailMode() interface{} {
	return c.failMode
}

func (c *ConnCenter) GetRetries() int {
	return c.retries
}

func (c *ConnCenter) SetServerConnPoolUnavailable(serverPool interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	serverPool.(*ServerConnPool).SetAvailable(false)
}
