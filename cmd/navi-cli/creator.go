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

func (c *Creator) createRootFolder(serviceRootPath string) {
	os.MkdirAll(serviceRootPath+"/gen", 0755)
}

func (c *Creator) createGrpcFolders() {
	os.MkdirAll(c.c.ServiceRootPathAbsolute()+"/gen/proto", 0755)
	os.MkdirAll(c.c.ServiceRootPathAbsolute()+"/grpcapi/component", 0755)
	os.MkdirAll(c.c.ServiceRootPathAbsolute()+"/grpcservice/impl", 0755)
}

func (c *Creator) createThriftFolders() {
	os.MkdirAll(c.c.ServiceRootPathAbsolute()+"/gen/thrift", 0755)
	os.MkdirAll(c.c.ServiceRootPathAbsolute()+"/thriftapi/component", 0755)
	os.MkdirAll(c.c.ServiceRootPathAbsolute()+"/thriftapi/engine", 0755)
	os.MkdirAll(c.c.ServiceRootPathAbsolute()+"/thriftservice/impl", 0755)
}

func (c *Creator) createServiceYaml(serviceRootPath, serviceName, configFileName string) {
	type serviceYamlValues struct {
		ServiceRoot string
		ServiceName string
	}
	if _, err := os.Stat(serviceRootPath + "/" + configFileName + ".yaml"); err == nil {
		return
	}
	writeFileWithTemplate(
		serviceRootPath+"/"+configFileName+".yaml",
		serviceYamlValues{ServiceRoot: serviceRootPath, ServiceName: serviceName},
		`config:
  environment: development
  service_root_path: {{.ServiceRoot}}
  turbo_log_path: log
  http_port: 8081
  grpc_service_name: {{.ServiceName}}
  grpc_service_host: 127.0.0.1
  grpc_service_port: 50051
  thrift_service_name: {{.ServiceName}}
  thrift_service_host: 127.0.0.1
  thrift_service_port: 50052
  zookeeper_servers_addr: 127.0.0.1:2181
  zookeeper_url_service_path: /navi/service
  zookeeper_http_service_path: /navi/httpservice
  zookeeper_rpc_service_path: /navi/rpcservice

urlmapping:
  - POST /hello SayHello
`)
}

func (c *Creator) createProto(serviceName string) {
	type protoValues struct {
		ServiceName string
	}
	nameLower := strings.ToLower(serviceName)
	writeFileWithTemplate(
		c.c.ServiceRootPathAbsolute()+"/"+nameLower+".proto",
		protoValues{ServiceName: serviceName},
		`syntax = "proto3";
package proto;

message SayHelloRequest {
    string yourName = 1;
}

message SayHelloResponse {
    string message = 1;
}

service {{.ServiceName}} {
    rpc sayHello (SayHelloRequest) returns (SayHelloResponse) {}
}
`,
	)
}

func (c *Creator) createThrift(serviceName string) {
	type thriftValues struct {
		ServiceName string
	}
	nameLower := strings.ToLower(serviceName)
	writeFileWithTemplate(
		c.c.ServiceRootPathAbsolute()+"/"+nameLower+".thrift",
		thriftValues{ServiceName: serviceName},
		`
/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

# Thrift Tutorial
# Mark Slee (mcslee@facebook.com)
#
# This file aims to teach you how to use Thrift, in a .thrift file. Neato. The
# first thing to notice is that .thrift files support standard shell comments.
# This lets you make your thrift file executable and include your Thrift build
# step on the top line. And you can place comments like this anywhere you like.
#
# Before running this file, you will need to have installed the thrift compiler
# into /usr/local/bin.

/**
 * The first thing to know about are types. The available types in Thrift are:
 *
 *  bool        Boolean, one byte
 *  i8 (byte)   Signed 8-bit integer
 *  i16         Signed 16-bit integer
 *  i32         Signed 32-bit integer
 *  i64         Signed 64-bit integer
 *  double      64-bit floating point value
 *  string      String
 *  binary      Blob (byte array)
 *  map<t1,t2>  Map from one type to another
 *  list<t1>    Ordered list of one type
 *  set<t1>     Set of unique elements of one type
 *
 * Did you also notice that Thrift supports C style comments?
 */

namespace go gen

# 这个结构体定义了服务调用者的请求信息
/*struct Request {
    # 传递的参数信息，使用格式进行表示
    1:required binary paramJSON;
    # 服务调用者请求的服务名，使用serviceName属性进行传递
    2:required string serviceName
}*/

# 这个结构体，定义了服务提供者的返回信息
struct Response {
    # RESCODE 是处理状态代码，是一个枚举类型。例如RESCODE._200表示处理成功
    1:required RESCODE responseCode;
    # 返回的处理结果，同样使用JSON格式进行描述
    2:required string responseJSON;
}

# 异常描述定义，当服务提供者处理过程出现异常时，向服务调用者返回
exception ServiceException {
    # EXCCODE 是异常代码，也是一个枚举类型。
    # 例如EXCCODE.PARAMNOTFOUND表示需要的请求参数没有找到
    1:required EXCCODE exceptionCode;
    # 异常的描述信息，使用字符串进行描述
    2:required string exceptionMess;
}

# 这个枚举结构，描述各种服务提供者的响应代码
enum RESCODE {
    SUCCESS = 200;
	FORBIDDEN = 403;
	NOTFOUND = 404;
    BADGATEWAY = 502;
}

# 这个枚举结构，描述各种服务提供者的异常种类
enum EXCCODE {
    PARAMNOTFOUND = 2001;
    SERVICENOTFOUND = 2002;
}

# 这是经过泛化后的Apache Thrift接口
service {{.ServiceName}} {
        string Ping(),

        string ServiceName(),

        string ServiceType(),

        Response SayHello(1:string yourName) throws (1:required ServiceException e)
}
`,
	)
}

func (c *Creator) generateGrpcServiceMain() {
	type serviceMainValues struct {
		PkgPath        string
		ConfigFilePath string
	}
	nameLower := strings.ToLower(c.c.GrpcServiceName())
	writeFileWithTemplate(
		c.c.ServiceRootPathAbsolute()+"/grpcservice/"+nameLower+".go",
		serviceMainValues{PkgPath: c.PkgPath, ConfigFilePath: c.c.ServiceRootPathAbsolute() + "/service.yaml"},
		`package main

import (
	"{{.PkgPath}}/grpcservice/impl"
	"git.oschina.net/kuaishangtong/navi/cmd/navi-cli"
	"os/signal"
	"os"
	"syscall"
	"fmt"
)

func main() {
	s := navicli.NewGrpcServer(nil, "{{.ConfigFilePath}}")
	s.StartGrpcService(impl.RegisterServer)

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)

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

func (c *Creator) generateThriftServiceMain() {
	type thriftServiceMainValues struct {
		PkgPath        string
		ConfigFilePath string
	}
	nameLower := strings.ToLower(c.c.ThriftServiceName())
	writeFileWithTemplate(
		c.c.ServiceRootPathAbsolute()+"/thriftservice/"+nameLower+".go",
		thriftServiceMainValues{PkgPath: c.PkgPath, ConfigFilePath: c.c.ServiceRootPathAbsolute() + "/service.yaml"},
		`package main

import (
	"{{.PkgPath}}/thriftservice/impl"
	"git.oschina.net/kuaishangtong/navi/cmd/navi-cli"
	"os/signal"
	"os"
	"syscall"
	"fmt"
)

func main() {
	s := navicli.NewThriftServer(nil, "{{.ConfigFilePath}}")
	s.StartThriftService(impl.TProcessor)

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)

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

func (c *Creator) generateGrpcServiceImpl() {
	type serviceImplValues struct {
		PkgPath     string
		ServiceName string
	}
	nameLower := strings.ToLower(c.c.GrpcServiceName())
	writeFileWithTemplate(
		c.c.ServiceRootPathAbsolute()+"/grpcservice/impl/"+nameLower+"impl.go",
		serviceImplValues{PkgPath: c.PkgPath, ServiceName: c.c.GrpcServiceName()},
		`package impl

import (
	"golang.org/x/net/context"
	"{{.PkgPath}}/gen/proto"
	"google.golang.org/grpc"
)

// RegisterServer registers a service struct to a server
func RegisterServer(s *grpc.Server) {
	proto.Register{{.ServiceName}}Server(s, &{{.ServiceName}}{})
}

// {{.ServiceName}} is the struct which implements generated interface
type {{.ServiceName}} struct {
}

// SayHello is an example entry point
func (s *{{.ServiceName}}) SayHello(ctx context.Context, req *proto.SayHelloRequest) (*proto.SayHelloResponse, error) {
	return &proto.SayHelloResponse{Message: "[grpc server]Hello, " + req.YourName}, nil
}
`,
	)
}

func (c *Creator) generateThriftServiceImpl() {
	type thriftServiceImplValues struct {
		PkgPath     string
		ServiceName string
	}
	nameLower := strings.ToLower(c.c.ThriftServiceName())
	writeFileWithTemplate(
		c.c.ServiceRootPathAbsolute()+"/thriftservice/impl/"+nameLower+"impl.go",
		thriftServiceImplValues{PkgPath: c.PkgPath, ServiceName: c.c.ThriftServiceName()},
		`package impl

import (
	"{{.PkgPath}}/gen/thrift/gen-go/gen"
	"git.apache.org/thrift.git/lib/go/thrift"
)

// TProcessor returns TProcessor
func TProcessor() thrift.TProcessor {
	return gen.New{{.ServiceName}}Processor({{.ServiceName}}{})
}

// {{.ServiceName}} is the struct which implements generated interface
type {{.ServiceName}} struct {
}

func (s {{.ServiceName}}) Ping() (str string, err error) {
	return "ping",nil
}

func (s {{.ServiceName}}) ServiceName() (str string, err error) {
	return "serviceName",nil
	return "serviceName",nil
}

func (s {{.ServiceName}}) ServiceType() (str string, err error) {
	return "ServiceType",nil
}

// SayHello is an example entry point
func (s {{.ServiceName}}) SayHello(yourName string) (r *gen.Response, err error) {
	return &gen.Response{ResponseCode: gen.RESCODE_SUCCESS, ResponseJSON: "{name: Hello, " + yourName + "}"}, nil
}
`,
	)
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
	err := engine.InitEngine(s.Config.ZookeeperRpcServicePath(), s.Config.ThriftServiceName(), s.Config.ZookeeperServersAddr(), 2, 15)
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
	host     string
	interval int
	*t.{{.ServiceName}}Client
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
					log.Error(err)
					continue
				} else {
					c.close()
					c.{{.ServiceName}}Client = newClient
					c.available = true
				}
			} else {
				c.available = true
			}
		}
	}
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

func (c *Engine) getConn() (*Conn, error) {
	h := c.selector.Select(context.Background(), "", "", nil)
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

func (c *Creator) generateServiceMain(rpcType string) {
	type rootMainValues struct {
		PkgPath        string
		ConfigFilePath string
	}
	if rpcType == "grpc" {
		writeFileWithTemplate(
			c.c.ServiceRootPathAbsolute()+"/main.go",
			rootMainValues{PkgPath: c.PkgPath, ConfigFilePath: c.c.ServiceRootPathAbsolute() + "/service.yaml"},
			rootMainGrpc,
		)
	} else if rpcType == "thrift" {
		writeFileWithTemplate(
			c.c.ServiceRootPathAbsolute()+"/main.go",
			rootMainValues{PkgPath: c.PkgPath, ConfigFilePath: c.c.ServiceRootPathAbsolute() + "/service.yaml"},
			rootMainThrift,
		)
	}
}

var rootMainGrpc string = `package main

import (
	"git.oschina.net/kuaishangtong/navi/cmd/navi-cli"
	"{{.PkgPath}}/gen"
	gcomponent "{{.PkgPath}}/grpcapi/component"
	gimpl "{{.PkgPath}}/grpcservice/impl"
	"os/signal"
	"os"
	"syscall"
	"fmt"
)

func main() {
	s := navicli.NewGrpcServer(&gcomponent.ServiceInitializer{}, "{{.ConfigFilePath}}")
	s.Start(gcomponent.GrpcClient, gen.GrpcSwitcher, gimpl.RegisterServer)

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-exit:
		fmt.Println("Service is stopping...")
	}

	s.Stop()
	fmt.Println("Service stopped")
}
`

var rootMainThrift string = `package main

import (
	"git.oschina.net/kuaishangtong/navi/cmd/navi-cli"
	"{{.PkgPath}}/gen"
	tcomponent "{{.PkgPath}}/thriftapi/component"
	"{{.PkgPath}}/thriftapi/engine"
	timpl "{{.PkgPath}}/thriftservice/impl"
	"os/signal"
	"os"
	"syscall"
	"fmt"
)

func main() {
	s := navicli.NewThriftServer(&tcomponent.ServiceInitializer{}, "{{.ConfigFilePath}}")
	err := engine.InitEngine(s.Config.ZookeeperRpcServicePath(), s.Config.ThriftServiceName(), s.Config.ZookeeperServersAddr(), 2, 15)
	if err != nil {
		log.Fatal(err)
	}
	s.Start(tcomponent.ThriftClient, gen.ThriftSwitcher, timpl.TProcessor, engine.XEngine)

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-exit:
		fmt.Println("Service is stopping...")
	}

	s.Stop()
	fmt.Println("Service stopped")
}

`
