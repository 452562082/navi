package conncenter

import (
	"context"
	"fmt"
	"kuaishangtong/common/utils/log"
	"kuaishangtong/navi/cmd/navi-cli"
	"kuaishangtong/navi/lb"
	"kuaishangtong/navi/registry"
	"sync"
	"time"
)

var connCenter *ConnCenter // 全局引擎

type RpcType int

var GRPC RpcType = 0
var THRIFT RpcType = 1

var gTimeout int
var gAllocSize int = 8

type ServerConnPool struct {
	lock *sync.RWMutex

	typ       RpcType
	free      []navicli.RPCConn
	nextAlloc int

	closed    bool
	available bool
	host      string
	interval  int
}

func newServerConnPool(host string, typ RpcType, interval int) (*ServerConnPool, error) {
	shpool := &ServerConnPool{
		typ:       typ,
		free:      make([]navicli.RPCConn, 0, gAllocSize),
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
	conns := make([]navicli.RPCConn, s.nextAlloc, s.nextAlloc)

	for i := 0; i < len(conns); i++ {
		var conn navicli.RPCConn
		var err error

		switch s.typ {
		case THRIFT:
			conn, err = newThriftConn(s)
			if err != nil {
				log.Errorf("ServerConnPool newThriftConn to %s err: %v", s.host, err)
			} else {
				s.SetAvailable(true)
			}
		case GRPC:
			conn, err = newGrpcConn(s)
			if err != nil {
				log.Errorf("ServerConnPool newGrpcConn to %s err: %v", s.host, err)
			} else {
				s.SetAvailable(true)
			}
		default:
			conn = nil
			log.Errorf("ServerConnPool [%s] new unknown rpc connect", s.host)
		}
		conns[i] = conn
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

func (s *ServerConnPool) getConn() (conn navicli.RPCConn) {
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

func (s *ServerConnPool) putConn(conn navicli.RPCConn) {
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

			s.lock.Lock()
			length := len(s.free)
			capab := cap(s.free)
			log.Infof("STAT ServerConnPool [%s]: count of conn %d", s.host, length)

			if capab <= gAllocSize || length <= gAllocSize {
				if length == 0 {
					s.grow()
				}

				s.lock.Unlock()
				continue
			}
			s.lock.Unlock()

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

	rpcTyp RpcType

	interval int
	retries  int

	closed bool

	lock *sync.RWMutex
}

func InitConnCenter(basePath string, servicePath string, zkhosts []string, typ RpcType, timeout, interval, connNum int, failMode lb.FailMode) (err error) {
	gTimeout = timeout
	gAllocSize = connNum

	connCenter = &ConnCenter{
		rpcTyp:      typ,
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
		pool, err := newServerConnPool(host, typ, interval)
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

func (cc *ConnCenter) Close() {
	cc.closed = true
	for _, connPool := range cc.serverPools {
		connPool.close()
	}
}

func (cc *ConnCenter) serviceDiscovery() {

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for !cc.closed {
		select {
		case pairs, ok := <-cc.discovery.WatchService():

			if !ok {
				continue
			}

			newservers := make(map[string]string)
			servers := make([]string, 0, len(newservers))
			for _, p := range pairs {
				newservers[p.Key] = p.Value
				servers = append(servers, p.Key)
			}

			cc.selector.UpdateServer(newservers)
			log.Infof("ConnCenter UpdateServer %v", servers)

			var oldmap, newmap map[string]struct{} = make(map[string]struct{}), make(map[string]struct{})
			for k, _ := range cc.serverPools {
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
				if pool, ok := cc.serverPools[host]; ok {
					pool.close()
				}
				delete(cc.serverPools, host)
			}

			// 创建新的server对应的连接池
			for host, _ := range newmap {
				pool, err := newServerConnPool(host, cc.rpcTyp, cc.interval)
				if err != nil {
					log.Error(err)
					continue
				}
				cc.serverPools[host] = pool
			}
		}
	}

}

func (cc *ConnCenter) getServices() map[string]string {
	kvpairs := cc.discovery.GetServices()
	servers := make(map[string]string)
	for _, p := range kvpairs {
		servers[p.Key] = p.Value
	}
	return servers
}

func (cc *ConnCenter) GetConn() (navicli.RPCConn, error) {
	return cc.getConn()
}

func (cc *ConnCenter) getConn() (navicli.RPCConn, error) {

	var host string
	var serverCount, index int = len(cc.serverPools), 0
	for {
		index++
		host = cc.selector.Select(context.Background(), "", "", host, nil)
		if host == "" {
			return nil, fmt.Errorf("can not find available serverConnPool")
		}

		if cc.serverPools[host].Available() {
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

	cc.lock.Lock()
	if pool, ok := cc.serverPools[host]; ok {
		conn := pool.getConn()
		if conn == nil {
			log.Errorf("can not find available conn in %s", host)
			cc.lock.Unlock()
			return nil, fmt.Errorf("can not find available conn in %s", host)
		}
		cc.lock.Unlock()
		return conn, nil
	}
	cc.lock.Unlock()

	log.Errorf("can not find available conn in %s", host)
	return nil, fmt.Errorf("can not find available conn in %s", host)
}

func (cc *ConnCenter) PutConn(conn navicli.RPCConn) error {
	return cc.putConn(conn)
}

func (cc *ConnCenter) putConn(conn navicli.RPCConn) error {
	cc.lock.Lock()
	if pool, ok := cc.serverPools[conn.HostName()]; ok {
		pool.putConn(conn)
	}
	cc.lock.Unlock()

	conn.Close()
	return fmt.Errorf("can not find serverConnPool %s", conn.HostName())
}

func (cc *ConnCenter) GetFailMode() interface{} {
	return cc.failMode
}

func (cc *ConnCenter) GetRetries() int {
	return cc.retries
}

func (cc *ConnCenter) SetServerConnPoolUnavailable(serverPool interface{}) {
	cc.lock.Lock()
	defer cc.lock.Unlock()

	serverPool.(*ServerConnPool).SetAvailable(false)
}
