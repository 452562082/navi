/*
 * Copyright © 2017 Xiao Zhang <zzxx513@gmail.com>.
 * Use of this source code is governed by an MIT-style
 * license that can be found in the LICENSE file.
 */
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
  zookeeper_service_base_path: /navi-test/servicebase
  zookeeper_service_list_path: /navi-test/servicelist

urlmapping:
  - GET /hello SayHello
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
    1:required  RESCODE responseCode;
    # 返回的处理结果，同样使用JSON格式进行描述
    2:required  binary responseJSON;
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
    _200=200;
    _500=500;
    _400=400;
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
}

func (s {{.ServiceName}}) ServiceType() (str string, err error) {
	return "ServiceType",nil
}

// SayHello is an example entry point
func (s {{.ServiceName}}) SayHello(yourName string) (r *gen.Response, err error) {
	return &gen.Response{ResponseCode: gen.RESCODE__200, ResponseJSON: []byte("{name: [thrift server]Hello, " + yourName + "}")}, nil
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
)

func main() {
	s := navicli.NewGrpcServer(&component.ServiceInitializer{}, "{{.ConfigFilePath}}")
	s.StartHTTPServer(component.GrpcClient, gen.GrpcSwitcher)

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
	"git.oschina.net/kuaishangtong/navi/cmd/navi-cli"
	"{{.PkgPath}}/gen"
	"{{.PkgPath}}/thriftapi/component"
	"os/signal"
	"os"
	"syscall"
	"fmt"
)

func main() {
	s := navicli.NewThriftServer(&component.ServiceInitializer{}, "{{.ConfigFilePath}}")
	s.StartHTTPServer(component.ThriftClient, gen.ThriftSwitcher)

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
	//tcomponent "{{.PkgPath}}/thriftapi/component"
	//timpl "{{.PkgPath}}/thriftservice/impl"
	"os/signal"
	"os"
	"syscall"
	"fmt"
)

func main() {
	s := navicli.NewGrpcServer(&gcomponent.ServiceInitializer{}, "{{.ConfigFilePath}}")
	s.Start(gcomponent.GrpcClient, gen.GrpcSwitcher, gimpl.RegisterServer)

	//s := navicli.NewThriftServer(&tcomponent.ServiceInitializer{}, "{{.ConfigFilePath}}")
	//s.Start(tcomponent.ThriftClient, gen.ThriftSwitcher, timpl.TProcessor)

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
	//gcomponent "{{.PkgPath}}/grpcapi/component"
	//gimpl "{{.PkgPath}}/grpcservice/impl"
	tcomponent "{{.PkgPath}}/thriftapi/component"
	timpl "{{.PkgPath}}/thriftservice/impl"
	"os/signal"
	"os"
	"syscall"
	"fmt"
	"net"
	"strings"
	"time"
	metrics "github.com/rcrowley/go-metrics"
	"git.oschina.net/kuaishangtong/navi/registry"
	"git.oschina.net/kuaishangtong/common/utils/log"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv"
)

func main() {
	//s := navicli.NewGrpcServer(&gcomponent.ServiceInitializer{}, "{{.ConfigFilePath}}")
	//s.Start(gcomponent.GrpcClient, gen.GrpcSwitcher, gimpl.RegisterServer)

	s := navicli.NewThriftServer(&tcomponent.ServiceInitializer{}, "{{.ConfigFilePath}}")
	s.Start(tcomponent.ThriftClient, gen.ThriftSwitcher, timpl.TProcessor)

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)

	//服务注册
	address, err := getaddr()
	if err != nil {
		log.Fatal(err)
	}

	port := fmt.Sprintf("%d", s.Config.HTTPPort())

	r := &registry.ZooKeeperRegister{
		ServiceAddress: address+":"+port,

		ZooKeeperServers:    s.Config.ZookeeperServersAddr(),


		BasePath:       s.Config.ZookeeperServiceBasePath(),
		Metrics:          metrics.NewRegistry(),
		UpdateInterval:   2 * time.Second,
	}

	err = r.Start()
	if err != nil {
		log.Fatal(err)
	}


	kv, err := libkv.NewStore(store.ZK, r.ZooKeeperServers, nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range s.Config.UrlMappings() {
		path := v[1]

		err = kv.Put(s.Config.ZookeeperServiceBasePath() + path, nil, nil)
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

func getaddr() (string,error) {
	conn, err := net.Dial("udp", "www.google.com.hk:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	return strings.Split(conn.LocalAddr().String(),":")[0], nil
}
`
