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
	c.createRootFolder(GOPATH() + "/src/" + c.PkgPath)
	c.createServiceYaml(GOPATH()+"/src/"+c.PkgPath, serviceName, "service")
	c.c = NewConfig(c.RpcType, GOPATH()+"/src/"+c.PkgPath+"/service.yaml")
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
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	
	"git.oschina.net/kuaishangtong/common/utils/log"
	"git.oschina.net/kuaishangtong/navi/cmd/navi-cli"
	"git.oschina.net/kuaishangtong/navi/registry"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	metrics "github.com/rcrowley/go-metrics"
	"{{.PkgPath}}/gen"
	"{{.PkgPath}}/thriftapi/component"
	"{{.PkgPath}}/thriftapi/engine"
	"git.oschina.net/kuaishangtong/navi/lb"
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
	s := navicli.NewThriftServer(&component.ServiceInitializer{}, "{{.ConfigFilePath}}")
	err := engine.InitEngine(s.Config.ZookeeperRpcServicePath(), s.Config.ThriftServiceName() + "/" + s.Config.ServiceVersionMode(), s.Config.ZookeeperServersAddr(), 2, 15, lb.Failover)
	if err != nil {
		log.Fatal(err)
	}
	s.StartHTTPServer(component.ThriftClient, gen.ThriftSwitcher, engine.XEngine)

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
		Mode:				s.Config.ServiceVersionMode(),
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
		path := strFirstToUpper(v[1])

		key := strings.Trim(s.Config.ZookeeperURLServicePath(),"/") + s.Config.ThriftServiceName() + "/" + s.Config.ServiceVersionMode() + path
		log.Infof("register url %s to registry in service %s", key, s.Config.ThriftServiceName())
		err = kv.Put(key, nil, nil)
		if err != nil {
			log.Fatal(err)
		}
	}

	select {
	case <-exit:
		log.Info("Service is stopping...")
	}

	s.Stop()
	log.Info("Service stopped")
}

func strFirstToUpper(str string) string {

	var upperStr string
	vv := []rune(str)
	for i := 0; i < len(vv); i++ {
		if i == 0 {
			if vv[i] >= 97 && vv[i] <= 122 {
				vv[i] -= 32
			}
			upperStr += string(vv[i]) // + string(vv[i+1])
		} else {
			upperStr += string(vv[i])
		}
	}
	return upperStr
}`,
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
	"git.oschina.net/kuaishangtong/common/utils/log"
	"git.oschina.net/kuaishangtong/navi/lb"
	"git.oschina.net/kuaishangtong/navi/registry"
	t "{{.PkgPath}}/gen/thrift/gen-go/gen"
	"github.com/valyala/fastrand"
	"sync"
	"time"
)

var XEngine *Engine // 全局引擎

var gTimeout int
var gMaxConns int

type Conn struct {
	serverHost    *ServerHost
	host     string
	interval int
	*t.{{.ServiceName}}Client
	closed    bool
	available bool
}

func newConn(serverHost *ServerHost, host string, interval int) (*Conn, error) {
	var err error
	conn := &Conn{
		serverHost: serverHost,
		host:      host,
		interval:  interval,
		closed:    false,
		available: true,
	}

	conn.{{.ServiceName}}Client, err = conn.connect()
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
					log.Error(err1)
					continue
				} else {
					c.close()
					c.{{.ServiceName}}Client = newClient
					c.available = true
					c.serverHost.available = true
				}
			} else {
				c.available = true
			}
		}
	}
}

func (c *Conn) GetServerHost() (*ServerHost) {
	return c.serverHost
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

func (c *Conn) close() {
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

type ServerHost struct {
	host  string
	lock  *sync.RWMutex
	conns []*Conn
	available bool
}

func newServerHost(host string, maxConns int) (*ServerHost, error) {
	serverHost := &ServerHost{
		host: host,
		lock: new(sync.RWMutex),
	}
	conns := make([]*Conn, maxConns, maxConns)
	for i := 0; i < maxConns; i++ {
		conn, err := newConn(serverHost, host, 1)
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
	failMode 	lb.FailMode
	lock  *sync.RWMutex

	// Retries retries to send
	retries int
}

func InitEngine(basePath string, servicePath string, zkhosts []string, timeout, maxConns int, failMode lb.FailMode) (err error) {
	gTimeout = timeout
	gMaxConns = maxConns
	XEngine = &Engine{
		servers: make(map[string]*ServerHost),
		failMode: failMode,
	}

	XEngine.discovery, err = registry.NewZookeeperDiscovery(basePath, servicePath, zkhosts, nil)
	if err != nil {
		return
	}

	log.Infof("Engine NewZookeeperDiscovery in /%s/%s", basePath, servicePath)

	selecter := lb.NewSelector(lb.RoundRobin, nil)

	initserver := XEngine.getServices()
	
	for key, _ := range initserver {
			host, err := newServerHost(key, gMaxConns)
			if err != nil {
				log.Error(err)
				return err
			}
			XEngine.servers[key] = host
			log.Infof("Engine add conn host in %s", key)
	}

	selecter.UpdateServer(initserver)

	XEngine.selector = selecter

	XEngine.lock = new(sync.RWMutex)
	XEngine.retries = 3

	return
}

func (c *Engine) Close() {
	for _, host := range c.servers {
		host.closeAllConns()
	}
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
		log.Debugf("Engine getServices in %s : %s", p.Key, p.Value)
	}
	return servers
}

func (c *Engine) GetConn() (interface{}, error) {
	return c.getConn()
}

func (c *Engine) GetFailMode() (interface{}) {
	return c.failMode
}

func (c *Engine) GetRetries() (int) {
	return c.retries
}

//func (c *Engine) ClearInvalidHost() {
//	c.lock.Lock()
//	defer c.lock.Unlock()
//	c.invalidHost = c.invalidHost[:0]
//}

func (c *Engine) SetServerHostUnavailable(serverHost interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	//c.invalidHost = append(c.invalidHost, conn.(*Conn).getHost())
	serverHost.(*ServerHost).available = false
}

func (c *Engine) getConn() (*Conn, error) {
	//c.lock.RLock()
	//for{
	//	h = c.selector.Select(context.Background(), "", "", nil)
		//isExist := false
		//for _, host := range c.invalidHost {
		//	if h == host {
		//		isExist = true
		//	}
		//}
		//
		//if !isExist {
		//	break
		//}
	//}
	//c.lock.RUnlock()
	var h string
	for{
		h = c.selector.Select(context.Background(), "", "", h, nil)
		if c.servers[h].available {
			break
		}
	}

	if host, ok := c.servers[h]; ok {
		conn := host.getConn()
		if conn == nil {
			return nil, fmt.Errorf("can not find available conn in %s", host)
		}
		return conn, nil
	}

	return nil, fmt.Errorf("can not find available conn in %s", h)
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
	"git.oschina.net/kuaishangtong/navi/cmd/navi-cli"
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
	"git.oschina.net/kuaishangtong/navi/cmd/navi-cli"
	"os/signal"
	"os"
	"syscall"
	"fmt"
	"net"
	"time"
	"strings"
	"git.oschina.net/kuaishangtong/common/utils/log"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"git.oschina.net/kuaishangtong/navi/registry"
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
	"git.oschina.net/kuaishangtong/navi/cmd/navi-cli"
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
