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

type ThriftConn struct {
	scpool   *ServerConnPool
	host     string
	interval int
	*navi.NaviServiceClient
	closed    bool
	available bool

	reConnFlag chan struct{}
}

func newThriftConn(pool *ServerConnPool) (*ThriftConn, error) {
	var err error
	conn := &ThriftConn{
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

func (tc *ThriftConn) check() {

	for !tc.closed {
		select {
		case <-tc.reConnFlag:
			tc.close()

			newClient, err := tc.connect()
			if err != nil {
				log.Errorf("reconnect to %s err: %v", tc.host, err)

				reconnectOK := false
				for !reconnectOK && !tc.closed {
					ticker := time.NewTicker(1 * time.Second)

					select {
					case <-ticker.C:
						newClient, err := tc.connect()
						if err != nil {
							log.Errorf("reconnect to %s err: %v", tc.host, err)
						} else {
							tc.NaviServiceClient = newClient
							tc.available = true
							reconnectOK = true
						}
					}
					ticker.Stop()
				}
			} else {
				tc.NaviServiceClient = newClient
				tc.available = true
			}
		}
	}
}

func (tc *ThriftConn) GetServerConnPool() *ServerConnPool {
	return tc.scpool
}

func (tc *ThriftConn) connect() (*navi.NaviServiceClient, error) {
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()

	transport, err := thrift.NewTSocketTimeout(tc.host, time.Duration(gTimeout)*time.Second)
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

func (tc *ThriftConn) Available() bool {
	return tc.available
}

// 当外部调用rpc接口报错err，一般为rpc连接出问题，需要重新连接
func (tc *ThriftConn) Reconnect() {
	tc.available = false
	tc.reConnFlag <- struct{}{}
}

// 关闭Conn下的thrift rpc连接
func (tc *ThriftConn) close() {
	tc.available = false
	if tc.NaviServiceClient != nil {
		tc.NaviServiceClient.InputProtocol.Transport().Close()
		tc.NaviServiceClient.OutputProtocol.Transport().Close()
		tc.NaviServiceClient.Transport.Close()
	}
}

// 关闭Conn连接
func (tc *ThriftConn) Close() {
	tc.closed = true
	tc.close()
	//close(tc.reConnFlag)
}

type ServerConnPool struct {
	lock      *sync.RWMutex
	free      []*ThriftConn
	nextAlloc int

	closed    bool
	available bool
	host      string
	interval  int
}

func newServerConnPool(host string, interval int) (*ServerConnPool, error) {
	shpool := &ServerConnPool{
		free:      make([]*ThriftConn, 0, gAllocSize),
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
	conns := make([]*ThriftConn, s.nextAlloc, s.nextAlloc)

	for i := 0; i < len(conns); i++ {
		conn, err := newThriftConn(s)
		if err != nil {
			log.Errorf("ServerConnPool newThriftConn to %s err: %v", s.host, err)
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

func (s *ServerConnPool) getConn() (conn *ThriftConn) {
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

func (s *ServerConnPool) putConn(conn *ThriftConn) {
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

func (tc *ConnCenter) Close() {
	tc.closed = true
	for _, connPool := range tc.serverPools {
		connPool.close()
	}
}

func (tc *ConnCenter) serviceDiscovery() {

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for !tc.closed {
		select {
		case pairs, ok := <-tc.discovery.WatchService():

			if !ok {
				continue
			}

			newservers := make(map[string]string)
			servers := make([]string, 0, len(newservers))
			for _, p := range pairs {
				newservers[p.Key] = p.Value
				servers = append(servers, p.Key)
			}

			tc.selector.UpdateServer(newservers)
			log.Infof("ConnCenter UpdateServer %v", servers)

			var oldmap, newmap map[string]struct{} = make(map[string]struct{}), make(map[string]struct{})
			for k, _ := range tc.serverPools {
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
				if pool, ok := tc.serverPools[host]; ok {
					pool.close()
				}
				delete(tc.serverPools, host)
			}

			// 创建新的server对应的连接池
			for host, _ := range newmap {
				pool, err := newServerConnPool(host, tc.interval)
				if err != nil {
					log.Error(err)
					continue
				}
				tc.serverPools[host] = pool
			}
		}
	}

}

func (tc *ConnCenter) getServices() map[string]string {
	kvpairs := tc.discovery.GetServices()
	servers := make(map[string]string)
	for _, p := range kvpairs {
		servers[p.Key] = p.Value
	}
	return servers
}

func (tc *ConnCenter) GetConn() (interface{}, error) {
	return tc.getConn()
}

func (tc *ConnCenter) getConn() (*ThriftConn, error) {

	var host string
	var serverCount, index int = len(tc.serverPools), 0
	for {
		index++
		host = tc.selector.Select(context.Background(), "", "", host, nil)
		if host == "" {
			return nil, fmt.Errorf("can not find available serverConnPool")
		}

		if tc.serverPools[host].Available() {
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

	tc.lock.Lock()
	if pool, ok := tc.serverPools[host]; ok {
		conn := pool.getConn()
		if conn == nil {
			log.Errorf("can not find available conn in %s", host)
			tc.lock.Unlock()
			return nil, fmt.Errorf("can not find available conn in %s", host)
		}
		tc.lock.Unlock()
		return conn, nil
	}
	tc.lock.Unlock()

	log.Errorf("can not find available conn in %s", host)
	return nil, fmt.Errorf("can not find available conn in %s", host)
}

func (tc *ConnCenter) PutConn(conn interface{}) error {
	return tc.putConn(conn.(*ThriftConn))
}

func (tc *ConnCenter) putConn(conn *ThriftConn) error {

	tc.lock.Lock()
	if pool, ok := tc.serverPools[conn.host]; ok {
		pool.putConn(conn)
	}
	tc.lock.Unlock()

	//if conn.scpool != nil {
	//	conn.scpool.close()
	//}
	conn.Close()
	return fmt.Errorf("can not find serverConnPool %s", conn.host)
}

func (tc *ConnCenter) GetFailMode() interface{} {
	return tc.failMode
}

func (tc *ConnCenter) GetRetries() int {
	return tc.retries
}

func (tc *ConnCenter) SetServerConnPoolUnavailable(serverPool interface{}) {
	tc.lock.Lock()
	defer tc.lock.Unlock()

	serverPool.(*ServerConnPool).SetAvailable(false)
}
