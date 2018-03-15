package navicli

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// TODO support plugin for e.g. customize folder structure

// Creator creates new projects
type Creator struct {
	RpcType string
	PkgPath string
	c       *Config
}

// CreateProject creates a whole new project!
func (c *Creator) CreateProject(serviceName string, force bool) {
	if !force {
		c.validateServiceRootPath(nil)
	}
	rootPath := GOPATH() + "/src/"
	c.createRootFolder(rootPath + c.PkgPath)
	c.createServiceYaml(rootPath+c.PkgPath, serviceName, "service")
	c.c = NewConfig(c.RpcType, rootPath+c.PkgPath+"/service.yaml")
	if c.RpcType == "grpc" {
		c.createGrpcProject(serviceName)
	} else if c.RpcType == "thrift" {
		c.createThriftProject(serviceName)
	}
}

func (c *Creator) createThriftProject(serviceName string) {
	c.createThriftFolders()
	c.createThrift(serviceName)
	c.generateThriftServiceMain()
	c.generateThriftServiceImpl()
	c.generateThriftHTTPMain()
	c.generateThriftHTTPComponent()
	c.generateThriftConnPool()
	c.generateServiceMain("thrift")

	g := Generator{
		RpcType:        c.RpcType,
		PkgPath:        c.PkgPath,
		ConfigFileName: "service",
	}
	g.c = NewConfig(g.RpcType, c.c.ServiceRootPathAbsolute()+"/"+g.ConfigFileName+".yaml")
	g.Options = " -I " + c.c.ServiceRootPathAbsolute() + " "
	g.GenerateThriftStub()
	g.GenerateBuildThriftParameters()
	g.c.loadFieldMapping()
	g.GenerateThriftSwitcher()

}

func (c *Creator) generateThriftHTTPMain() {
	type HTTPMainValues struct {
		ServiceName    string
		PkgPath        string
		ConfigFilePath string
	}
	nameLower := strings.ToLower(c.c.ThriftServiceName())
	writeFileWithTemplate(
		c.c.ServiceRootPathAbsolute()+"/thriftapi/"+nameLower+"api.go",
		HTTPMainValues{
			ServiceName:    c.c.ThriftServiceName(),
			PkgPath:        c.PkgPath,
			ConfigFilePath: c.c.ServiceRootPathAbsolute() + "/service.yaml"},
		`package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	
	"kuaishangtong/common/utils/log"
	"kuaishangtong/navi/cmd/navi-cli"
	"{{.PkgPath}}/gen"
	"{{.PkgPath}}/thriftapi/component"
	"{{.PkgPath}}/thriftapi/engine"
	"kuaishangtong/navi/lb"
)

func main() {
	s := navicli.NewThriftServer(&component.ServiceInitializer{}, "{{.ConfigFilePath}}")
	//err := engine.InitEngine(s.Config.ZookeeperRpcServicePath(), s.Config.ThriftServiceName() + "/" + s.Config.ServiceVersionMode(), s.Config.ZookeeperServersAddr(), 2, 15, lb.Failover)
	err := engine.InitConnCenter(s.Config.ZookeeperRpcServicePath(), s.Config.ThriftServiceName() + "/" + s.Config.ServiceVersionMode(), s.Config.ZookeeperServersAddr(), 2, 1, 15, lb.Failover)
	if err != nil {
		log.Fatal(err)
	}

	s.StartHTTPServer(component.ThriftClient, gen.ThriftSwitcher, engine.XConnCenter)

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-exit:
		log.Info("Service is stopping...")
	}

	s.Stop()
	log.Info("Service stopped")
}
`,
	)
}

func (c *Creator) generateThriftConnPool() {
	type EngineMainValues struct {
		ServiceName string
		PkgPath     string
	}

	writeFileWithTemplate(
		c.c.ServiceRootPathAbsolute()+"/thriftapi/engine/engine.go",
		EngineMainValues{
			ServiceName: c.c.ThriftServiceName(),
			PkgPath:     c.PkgPath,
		},
		`package engine

import (
	"context"
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"kuaishangtong/common/utils/log"
	"kuaishangtong/navi/lb"
	"kuaishangtong/navi/registry"
	t "{{.PkgPath}}/gen/thrift/gen-go/gen"
	//"github.com/valyala/fastrand"
	"sync"
	"time"
)

//var XEngine *Engine // 全局引擎
var XConnCenter *ConnCenter // 全局引擎

var gTimeout int
//var gMaxConns int
var gAllocSize int = 8

type Conn struct {
	//serverHost    *ServerHost
	scpool		*ServerConnPool
	host     	string
	interval 	int
	*t.{{.ServiceName}}Client
	closed    	bool
	available 	bool

	reConnFlag chan struct{}
}

//func newConn(serverHost *ServerHost, host string, interval int) (*Conn, error) {
//	var err error
//	conn := &Conn{
//		serverHost: serverHost,
//		host:      host,
//		interval:  interval,
//		closed:    false,
//		available: true,
//	}
//
//	conn.{{.ServiceName}}Client, err = conn.connect()
//	if err != nil {
//		return nil, err
//	}
//
//	go conn.check()
//	return conn, nil
//}

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

	conn.{{.ServiceName}}Client, err = conn.connect()
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
							c.{{.ServiceName}}Client = newClient
							c.available = true
							reconnectOK = true
						}
					}
					ticker.Stop()
				}
			} else {
				c.{{.ServiceName}}Client = newClient
				c.available = true
			}
		}
	}
}

func (c *Conn) GetServerConnPool() (*ServerConnPool) {
	return c.scpool
}

func (c *Conn) connect() (*t.{{.ServiceName}}Client, error) {
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

	client := t.New{{.ServiceName}}ClientFactory(useTransport, protocolFactory)
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

//func (c *Conn) close() {
//	if c.{{.ServiceName}}Client != nil {
//		c.{{.ServiceName}}Client.InputProtocol.Transport().Close()
//		c.{{.ServiceName}}Client.OutputProtocol.Transport().Close()
//		c.{{.ServiceName}}Client.Transport.Close()
//	}
//}

func (c *Conn) Reconnect() {
	c.available = false
	c.reConnFlag <- struct{}{}
}

func (c *Conn) close() {
	c.available = false
	if c.{{.ServiceName}}Client != nil {
		c.{{.ServiceName}}Client.InputProtocol.Transport().Close()
		c.{{.ServiceName}}Client.OutputProtocol.Transport().Close()
		c.{{.ServiceName}}Client.Transport.Close()
	}
}

func (c *Conn) Close() {
	c.closed = true
	c.close()
}

//type ServerHost struct {
//	host  string
//	lock  *sync.RWMutex
//	conns []*Conn
//	available bool
//}

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
		free:		make([]*Conn, 0, gAllocSize),
		nextAlloc: gAllocSize,
		lock:		&sync.RWMutex{},
		closed:	false,
		available:	true,
		host:		host,
		interval:	interval,
	}

	go shpool.stat()

	return shpool, nil
}

//func newServerHost(host string, maxConns int) (*ServerHost, error) {
//	serverHost := &ServerHost{
//		host: host,
//		lock: new(sync.RWMutex),
//		available: true,
//	}
//
//	conns := make([]*Conn, maxConns, maxConns)
//	for i := 0; i < maxConns; i++ {
//		conn, err := newConn(serverHost, host, 1)
//		if err != nil {
//			return nil, err
//		}
//
//		conns[i] = conn
//	}
//
//	serverHost.conns = conns
//	return serverHost, nil
//}

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

//func (s *ServerHost) getConn() *Conn {
//	s.lock.RLock()
//	defer s.lock.RUnlock()
//
//	if s.conns != nil && len(s.conns) > 0 {
//		length := len(s.conns)
//		i := int(fastrand.Uint32n(uint32(length)))
//
//		loopCount := 0
//		for !s.conns[i].Available() {
//			loopCount++
//			if loopCount == length {
//				return nil
//			}
//			i = (i + 1) % length
//		}
//
//		return s.conns[i]
//	}
//	return nil
//}

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

//type Engine struct {
//	servers   map[string]*ServerHost
//	discovery registry.ServiceDiscovery
//	selector  lb.Selector
//	failMode 	lb.FailMode
//	lock  *sync.RWMutex
//
//	// Retries retries to send
//	retries int
//}

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

//func InitEngine(basePath string, servicePath string, zkhosts []string, timeout, maxConns int, failMode lb.FailMode) (err error) {
//	gTimeout = timeout
//	gMaxConns = maxConns
//	XEngine = &Engine{
//		servers: make(map[string]*ServerHost),
//		failMode: failMode,
//	}
//
//	XEngine.discovery, err = registry.NewZookeeperDiscovery(basePath, servicePath, zkhosts, nil)
//	if err != nil {
//		return
//	}
//
//	log.Infof("Engine NewZookeeperDiscovery in %s/%s", basePath, servicePath)
//
//	selecter := lb.NewSelector(lb.RoundRobin, nil)
//
//	initserver := XEngine.getServices()
//
//	for key, _ := range initserver {
//			host, err := newServerHost(key, gMaxConns)
//			if err != nil {
//				log.Error(err)
//				return err
//			}
//			XEngine.servers[key] = host
//			log.Infof("Engine add conn host in %s", key)
//	}
//
//	selecter.UpdateServer(initserver)
//
//	XEngine.selector = selecter
//
//	XEngine.lock = new(sync.RWMutex)
//	XEngine.retries = 3
//
//	return
//}

func InitConnCenter(basePath string, servicePath string, zkhosts []string, timeout, interval, connNum int, failMode lb.FailMode) (err error) {
	gTimeout = timeout
	gAllocSize = connNum

	XConnCenter = &ConnCenter{
		serverPools: make(map[string]*ServerConnPool),
		failMode:    failMode,
		interval:    interval,
		lock:        &sync.RWMutex{},
		retries:     3,
		closed:      false,
	}

	XConnCenter.discovery, err = registry.NewZookeeperDiscovery(basePath, servicePath, zkhosts, nil)
	if err != nil {
		return
	}

	log.Infof("ConnCenter NewZookeeperDiscovery in %s/%s", basePath, servicePath)

	initserver := XConnCenter.getServices()
	selecter := lb.NewSelector(lb.RoundRobin, initserver)

	for host, _ := range initserver {
		pool, err := newServerConnPool(host, interval)
		if err != nil {
			log.Error(err)
			return err
		}
		XConnCenter.serverPools[host] = pool
		log.Infof("ConnCenter add conn pool into %s", host)
	}

	selecter.UpdateServer(initserver)
	XConnCenter.selector = selecter

	go XConnCenter.serviceDiscovery()

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

func (c *ConnCenter) GetConn() (interface{}, error) {
	return c.getConn()
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
			return nil, fmt.Errorf("conn is nil in %s", host)
		}
		c.lock.Unlock()
		return conn, nil
	}
	c.lock.Unlock()

	log.Errorf("can not find available conn in %s", host)
	return nil, fmt.Errorf("can not find available conn in %s", host)
}

func (c *ConnCenter) PutConn(conn interface{}) error {
	return XConnCenter.putConn(conn.(*Conn))
}

func (c *ConnCenter) putConn(conn *Conn) error {

	c.lock.Lock()
	if pool, ok := c.serverPools[conn.host]; ok {
		pool.putConn(conn)
		c.lock.Unlock()
		return nil
	}
	c.lock.Unlock()

	//if conn.scpool != nil {
	//	conn.scpool.close()
	//}
	conn.Close()
	return fmt.Errorf("can not find serverConnPool %s", conn.host)
}

func (c *ConnCenter) GetFailMode() (interface{}) {
	return c.failMode
}

func (c *ConnCenter) GetRetries() (int) {
	return c.retries
}

func (c *ConnCenter) SetServerConnPoolUnavailable(serverConnPool interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	
	serverConnPool.(*ServerConnPool).SetAvailable(false)
}
`,
	)
}

func (c *Creator) generateThriftHTTPComponent() {
	type thriftHTTPComponentValues struct {
		ServiceName string
		PkgPath     string
	}
	writeFileWithTemplate(
		c.c.ServiceRootPathAbsolute()+"/thriftapi/component/components.go",
		thriftHTTPComponentValues{ServiceName: c.c.GrpcServiceName(), PkgPath: c.PkgPath},
		`package component

import (
	t "{{.PkgPath}}/gen/thrift/gen-go/gen"
	"git.apache.org/thrift.git/lib/go/thrift"
	"kuaishangtong/navi/cmd/navi-cli"
)

// ThriftClient returns a thrift client
func ThriftClient(trans thrift.TTransport, f thrift.TProtocolFactory) interface{} {
	return t.New{{.ServiceName}}ClientFactory(trans, f)
}

type ServiceInitializer struct {
}

// InitService is run before the service is started, do initializing staffs for your service here.
// For example, init navicli components, such as interceptors, pre/postprocessors, errorHandlers, etc.
func (i *ServiceInitializer) InitService(s navicli.Servable) error {
	// TODO
	return nil
}

// StopService is run after both grpc server and http server are stopped,
// do your cleaning up work here.
func (i *ServiceInitializer) StopService(s navicli.Servable) {
	// TODO
}
`,
	)
}

func (c *Creator) createGrpcProject(serviceName string) {
	c.createGrpcFolders()
	c.createProto(serviceName)
	c.generateGrpcServiceMain()
	c.generateGrpcServiceImpl()
	c.generateGrpcHTTPMain()
	c.generateGrpcHTTPComponent()
	c.generateServiceMain("grpc")

	g := Generator{
		RpcType:        c.RpcType,
		PkgPath:        c.PkgPath,
		ConfigFileName: "service",
	}
	g.c = NewConfig(g.RpcType, c.c.ServiceRootPathAbsolute()+"/"+g.ConfigFileName+".yaml")
	g.Options = " -I " + c.c.ServiceRootPathAbsolute() + " " + c.c.ServiceRootPathAbsolute() + "/" + strings.ToLower(serviceName) + ".proto "
	g.GenerateProtobufStub()
	g.c.loadFieldMapping()
	g.GenerateGrpcSwitcher()
}

func (c *Creator) createThriftFolders() {
	os.MkdirAll(c.c.ServiceRootPathAbsolute()+"/gen/thrift", 0755)
	os.MkdirAll(c.c.ServiceRootPathAbsolute()+"/thriftapi/component", 0755)
	os.MkdirAll(c.c.ServiceRootPathAbsolute()+"/thriftapi/engine", 0755)
	os.MkdirAll(c.c.ServiceRootPathAbsolute()+"/thriftservice/impl", 0755)
}

func (c *Creator) createGrpcFolders() {
	os.MkdirAll(c.c.ServiceRootPathAbsolute()+"/gen/proto", 0755)
	os.MkdirAll(c.c.ServiceRootPathAbsolute()+"/grpcapi/component", 0755)
	os.MkdirAll(c.c.ServiceRootPathAbsolute()+"/grpcservice/impl", 0755)
}

func (c *Creator) generateGrpcHTTPMain() {
	type HTTPMainValues struct {
		ServiceName    string
		PkgPath        string
		ConfigFilePath string
	}
	nameLower := strings.ToLower(c.c.GrpcServiceName())
	writeFileWithTemplate(
		c.c.ServiceRootPathAbsolute()+"/grpcapi/"+nameLower+"api.go",
		HTTPMainValues{
			ServiceName:    c.c.GrpcServiceName(),
			PkgPath:        c.PkgPath,
			ConfigFilePath: c.c.ServiceRootPathAbsolute() + "/service.yaml"},
		`package main

import (
	"{{.PkgPath}}/gen"
	"{{.PkgPath}}/grpcapi/component"
	"kuaishangtong/navi/cmd/navi-cli"
	"os/signal"
	"os"
	"syscall"
	"fmt"
	"net"
	"time"
	"strings"
	"kuaishangtong/common/utils/log"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"kuaishangtong/navi/registry"
	metrics "github.com/rcrowley/go-metrics"
)

func getaddr() (string,error) {
	conn, err := net.Dial("udp", "www.google.com.hk:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	return strings.Split(conn.LocalAddr().String(),":")[0], nil
}

func main() {
	s := navicli.NewGrpcServer(&component.ServiceInitializer{}, "{{.ConfigFilePath}}")
	s.StartHTTPServer(component.GrpcClient, gen.GrpcSwitcher)

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)

	//服务注册
	address, err := getaddr()
	if err != nil {
		log.Fatal(err)
	}

	port := fmt.Sprintf("%d", s.Config.HTTPPort())

	r := &registry.ZooKeeperRegister {
		ServiceAddress: address+":"+port,
		ZooKeeperServers:   s.Config.ZookeeperServersAddr(),
		BasePath:       	s.Config.ZookeeperHttpServicePath(),
		Metrics:         	metrics.NewRegistry(),
		UpdateInterval:   	2 * time.Second,
	}

	err = r.Start()
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Register %s host to Registry", s.Config.ThriftServiceName())
	err = r.Register(s.Config.ThriftServiceName(), nil, "")
	if err != nil {
		log.Fatal(err)
	}

	kv, err := libkv.NewStore(store.ZK, r.ZooKeeperServers, nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range s.Config.UrlMappings() {
		path := v[1]

		key := strings.Trim(s.Config.ZookeeperURLServicePath(),"/") + "/" + s.Config.ThriftServiceName() + path
		log.Infof("register url %s to registry in service %s", key, s.Config.ThriftServiceName())
		err = kv.Put(key, nil, nil)
		if err != nil {
			log.Fatal(err)
		}
	}


	select {
	case <-exit:
		fmt.Println("Service is stopping...")
	}

	s.Stop()
	fmt.Println("Service stopped")
}
`,
	)
}

func (c *Creator) generateGrpcHTTPComponent() {
	type HTTPComponentValues struct {
		ServiceName string
		PkgPath     string
	}
	writeFileWithTemplate(
		c.c.ServiceRootPathAbsolute()+"/grpcapi/component/components.go",
		HTTPComponentValues{ServiceName: c.c.GrpcServiceName(), PkgPath: c.PkgPath},
		`package component

import (
	"{{.PkgPath}}/gen/proto"
	"google.golang.org/grpc"
	"kuaishangtong/navi/cmd/navi-cli"
)

// GrpcClient returns a grpc client
func GrpcClient(conn *grpc.ClientConn) interface{} {
	return proto.New{{.ServiceName}}Client(conn)
}

type ServiceInitializer struct {
}

// InitService is run before the service is started, do initializing staffs for your service here.
// For example, init navicli components, such as interceptors, pre/postprocessors, errorHandlers, etc.
func (i *ServiceInitializer) InitService(s navicli.Servable) error {
	// TODO
	return nil
}

// StopService is run after both grpc server and http server are stopped,
// do your cleaning up work here.
func (i *ServiceInitializer) StopService(s navicli.Servable) {
	// TODO
}
`,
	)
}

func (c *Creator) validateServiceRootPath(in io.Reader) {
	if in == nil {
		in = os.Stdin
	}
	if len(strings.TrimSpace(c.PkgPath)) == 0 {
		panic("pkgPath is blank")
	}
	p := GOPATH() + "/src/" + c.PkgPath
	_, err := os.Stat(p)
	if os.IsNotExist(err) {
		return
	}
	fmt.Print("Path '" + p + "' already exist!\n" +
		"Do you want to remove this directory before creating a new project? (type 'y' to remove):")
	var input string
	fmt.Fscan(in, &input)
	if input != "y" {
		return
	}
	fmt.Print("All files in that directory will be lost, are you sure? (type 'y' to continue):")
	fmt.Fscan(in, &input)
	if input != "y" {
		panic("aborted")
	}
	os.RemoveAll(p)
}

func (c *Creator) createRootFolder(serviceRootPath string) {
	os.MkdirAll(serviceRootPath+"/gen", 0755)
}
