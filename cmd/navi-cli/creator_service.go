package navicli

import "strings"

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
	"kuaishangtong/navi/cmd/navi-cli"
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

func (c *Creator) generateThriftServiceImpl() {
	type thriftServiceImplValues struct {
		PkgPath     string
		ServiceName string
		ServiceMode string
	}
	nameLower := strings.ToLower(c.c.ThriftServiceName())
	writeFileWithTemplate(
		c.c.ServiceRootPathAbsolute()+"/thriftservice/impl/"+nameLower+"impl.go",
		thriftServiceImplValues{PkgPath: c.PkgPath, ServiceName: c.c.ThriftServiceName(), ServiceMode: c.c.ServiceVersionMode()},
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
	return "{{.ServiceName}}",nil
}

func (s {{.ServiceName}}) ServiceMode() (str string, err error) {
	return "{{.ServiceMode}}",nil
}

// SayHello is an example entry point
func (s {{.ServiceName}}) SayHello(yourName string) (r *gen.Response, err error) {
	return &gen.Response{ResponseCode: 200, ResponseJSON: "{\"name\": \"Hello, " + yourName + "\"}"}, nil
}

// SayHello is an example entry point
func (s {{.ServiceName}}) SaveWave(fileName string, wavFormat string, data []byte) (r *gen.Response, err error) {
	return &gen.Response{ResponseCode: 200, ResponseJSON: "{\"file_name\": \""+fileName+"\", \"wav_format\": \"" +wavFormat+"\", \"data\": \"" +string(data)+"\"}"}, nil
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
	"kuaishangtong/navi/cmd/navi-cli"
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

func (c *Creator) generateGrpcServiceImpl() {
	type serviceImplValues struct {
		PkgPath     string
		ServiceName string
		ServiceMode string
	}
	nameLower := strings.ToLower(c.c.GrpcServiceName())
	writeFileWithTemplate(
		c.c.ServiceRootPathAbsolute()+"/grpcservice/impl/"+nameLower+"impl.go",
		serviceImplValues{PkgPath: c.PkgPath, ServiceName: c.c.GrpcServiceName(), ServiceMode: c.c.ServiceVersionMode()},
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

//// SayHello is an example entry point
//func (s *{{.ServiceName}}) SayHello(ctx context.Context, req *proto.SayHelloRequest) (*proto.Response, error) {
//	return &proto.SayHelloResponse{Message: "[grpc server]Hello, " + req.YourName}, nil
//}

func (c *{{.ServiceName}}) Ping(ctx context.Context, req *proto.PingRequest) (*proto.Response, error) {
	return &proto.Response{ResponseCode: 200, ResponseJSON: "{\"message\": \"ping\"}"}, nil
}

func (c *{{.ServiceName}}) ServiceName(ctx context.Context, req *proto.ServiceNameRequest) (*proto.Response, error) {
	return &proto.Response{ResponseCode: 200, ResponseJSON: "{\"message\": \"{{.ServiceName}}\"}"}, nil
}

func (c *{{.ServiceName}}) ServiceMode(ctx context.Context, req *proto.ServiceModeRequest) (*proto.Response, error) {
	return &proto.Response{ResponseCode: 200, ResponseJSON: "{\"message\": \"{{.ServiceMode}}\"}"}, nil
}

func (c *{{.ServiceName}}) SayHello(ctx context.Context, req *proto.SayHelloRequest) (*proto.Response, error) {
	return &proto.Response{ResponseCode: 200, ResponseJSON: "{\"name\": \"Hello, " + req.YourName + "\"}"}, nil
}

func (c *{{.ServiceName}}) SaveWave(ctx context.Context, req *proto.SaveWaveRequest) (*proto.Response, error) {
	return &proto.Response{ResponseCode: 200, ResponseJSON: "{\"message\": \"savewave\"}"}, nil
}
`,
	)
}

func (c *Creator) generateServiceMain(rpcType string) {
	type rootMainValues struct {
		PkgPath        string
		ConfigFilePath string
		ServiceName    string
	}
	if rpcType == "grpc" {
		writeFileWithTemplate(
			c.c.ServiceRootPathAbsolute() + "/" + c.c.GrpcServiceName() + ".go",
			rootMainValues{PkgPath: c.PkgPath, ConfigFilePath: c.c.ServiceRootPathAbsolute() + "/service.yaml",ServiceName: c.c.GrpcServiceName()},
			rootMainGrpc,
		)
	} else if rpcType == "thrift" {
		writeFileWithTemplate(
			c.c.ServiceRootPathAbsolute()+ "/" + c.c.ThriftServiceName() + ".go",
			rootMainValues{PkgPath: c.PkgPath, ConfigFilePath: c.c.ServiceRootPathAbsolute() + "/service.yaml",ServiceName: c.c.ThriftServiceName()},
			rootMainThrift,
		)
	}
}

var rootMainGrpc string = `package main

import (
	"kuaishangtong/navi/cmd/navi-cli"
	"{{.PkgPath}}/gen"
	gcomponent "{{.PkgPath}}/grpcapi/component"
	"{{.PkgPath}}/grpcapi/engine"
	gimpl "{{.PkgPath}}/grpcservice/impl"
	"kuaishangtong/common/utils/log"
	"kuaishangtong/navi/lb"
	"os/signal"
	"os"
	"syscall"
	"fmt"
	"net"
	"flag"
	"strings"
	"time"
	"kuaishangtong/navi/registry"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	metrics "github.com/rcrowley/go-metrics"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jmetrics "github.com/uber/jaeger-lib/metrics"
)

var configFilePath = flag.String("path","{{.ConfigFilePath}}","set configFilePath")

func main() {
	flag.Parse()
	s := navicli.NewGrpcServer(&gcomponent.ServiceInitializer{}, *configFilePath)

	err := engine.InitConnCenter(s.Config.ZookeeperRpcServicePath(), "{{.ServiceName}}/" + s.Config.ServiceVersionMode(), s.Config.ZookeeperServersAddr(), 2, 1, 15, lb.Failover)
	if err != nil {
		log.Fatal(err)
	}

	// Recommended configuration for production.
	cfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort: s.Config.JaegerAddr(),
		},
	}
	jMetricsFactory := jmetrics.NullFactory

	closer, err := cfg.InitGlobalTracer(
		"{{.ServiceName}}",
		//jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		log.Fatalf("Could not initialize jaeger tracer: %s", err.Error())
		return
	}
	defer closer.Close()

	s.Start(gen.GrpcSwitcher, engine.XConnCenter, gimpl.RegisterServer)

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)

	serviceRegistry(s)

	// log 设置
	log.SetLogFuncCall(s.Config.ShowLines())
	log.SetColor(s.Config.Coloured())
	log.SetLevel(s.Config.Level())
	if s.Config.Enable() || s.Config.IsDocker() {
		log.SetLogFile(
			s.Config.FilePath(),
			s.Config.Level(),
			s.Config.Daily(),
			s.Config.Coloured(),
			s.Config.ShowLines(),
			s.Config.MaxDays())
	}
	log.Info("log setup is complete.")

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

func serviceRegistry(s *navicli.GrpcServer) {
	//服务注册
	var address string
	if s.Config.IsDocker() {
		address = s.Config.HTTPHost()
	}else {
		address, _ = getaddr()
		/*if err != nil {
			log.Fatal(err)
		}*/
	}

	r := &registry.ZooKeeperRegister {
		ServiceAddress: address,
		ZooKeeperServers:   s.Config.ZookeeperServersAddr(),
		BasePath:       	s.Config.ZookeeperHttpServicePath(),
		Mode:				s.Config.ServiceVersionMode(),
		Metrics:         	metrics.NewRegistry(),
		UpdateInterval:   	2 * time.Second,
	}

	err := r.Start()
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Register {{.ServiceName}} host to Registry")
	err = r.Register("{{.ServiceName}}", nil, "")
	if err != nil {
		log.Fatal(err)
	}

	kv, err := libkv.NewStore(store.ZK, r.ZooKeeperServers, nil)
	if err != nil {
		log.Fatal(err)
	}

	// 注册URL接口到zookeeper上，后续由admin后台手动管理，可删除该部分代码
	for _, v := range s.Config.UrlMappings() {
		path := strFirstToUpper(v[1])

		key := strings.Trim(s.Config.ZookeeperURLServicePath(),"/") + "/{{.ServiceName}}/" + s.Config.ServiceVersionMode() + path
		log.Infof("register url %s to registry in service {{.ServiceName}}", key)
		err = kv.Put(key, nil, nil)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func strFirstToUpper(str string) string {

	var upperStr string
	vv := []rune(str)
	for i := 0; i < len(vv); i++ {
		if i == 0 {
			if vv[i] >= 97 && vv[i] <= 122 {
				vv[i] -= 32
			}
			upperStr += string(vv[i])
		} else {
			upperStr += string(vv[i])
		}
	}
	return upperStr
}
`

var rootMainThrift string = `package main

import (
	"kuaishangtong/navi/cmd/navi-cli"
	"{{.PkgPath}}/gen"
	tcomponent "{{.PkgPath}}/thriftapi/component"
	"{{.PkgPath}}/thriftapi/engine"
	timpl "{{.PkgPath}}/thriftservice/impl"
	"kuaishangtong/common/utils/log"
	"kuaishangtong/navi/lb"
	"os/signal"
	"os"
	"syscall"
	"fmt"
	"net"
	"flag"
	"strings"
	"time"
	"kuaishangtong/navi/registry"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	metrics "github.com/rcrowley/go-metrics"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jmetrics "github.com/uber/jaeger-lib/metrics"
)

var configFilePath = flag.String("path","{{.ConfigFilePath}}","set configFilePath")

func main() {
	flag.Parse()
	s := navicli.NewThriftServer(&tcomponent.ServiceInitializer{}, *configFilePath)

	err := engine.InitConnCenter(s.Config.ZookeeperRpcServicePath(), "{{.ServiceName}}/" + s.Config.ServiceVersionMode(), s.Config.ZookeeperServersAddr(), 2, 1, 15, lb.Failover)
	if err != nil {
		log.Fatal(err)
	}

	// Recommended configuration for production.
	cfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort: s.Config.JaegerAddr(),
		},
	}
	jMetricsFactory := jmetrics.NullFactory

	closer, err := cfg.InitGlobalTracer(
		"{{.ServiceName}}",
		//jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		log.Fatalf("Could not initialize jaeger tracer: %s", err.Error())
		return
	}
	defer closer.Close()

	s.Start(gen.ThriftSwitcher, engine.XConnCenter, timpl.TProcessor)

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)

	serviceRegistry(s)

	// log 设置
	log.SetLogFuncCall(s.Config.ShowLines())
	log.SetColor(s.Config.Coloured())
	log.SetLevel(s.Config.Level())
	if s.Config.Enable() || s.Config.IsDocker() {
		log.SetLogFile(
			s.Config.FilePath(),
			s.Config.Level(),
			s.Config.Daily(),
			s.Config.Coloured(),
			s.Config.ShowLines(),
			s.Config.MaxDays())
	}
	log.Info("log setup is complete.")

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

func serviceRegistry(s *navicli.ThriftServer) {
	//服务注册
	var address string
	if s.Config.IsDocker() {
		address = s.Config.HTTPHost()
	}else {
		address, _ = getaddr()
		/*if err != nil {
			log.Fatal(err)
		}*/
	}

	r := &registry.ZooKeeperRegister {
		ServiceAddress: address,
		ZooKeeperServers:   s.Config.ZookeeperServersAddr(),
		BasePath:       	s.Config.ZookeeperHttpServicePath(),
		Mode:				s.Config.ServiceVersionMode(),
		Metrics:         	metrics.NewRegistry(),
		UpdateInterval:   	2 * time.Second,
	}

	err := r.Start()
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Register {{.ServiceName}} host to Registry")
	err = r.Register("{{.ServiceName}}", nil, "")
	if err != nil {
		log.Fatal(err)
	}

	kv, err := libkv.NewStore(store.ZK, r.ZooKeeperServers, nil)
	if err != nil {
		log.Fatal(err)
	}

	// 注册URL接口到zookeeper上，后续由admin后台手动管理，可删除该部分代码
	for _, v := range s.Config.UrlMappings() {
		path := strFirstToUpper(v[1])

		key := strings.Trim(s.Config.ZookeeperURLServicePath(),"/") + "/{{.ServiceName}}/" + s.Config.ServiceVersionMode() + path
		log.Infof("register url %s to registry in service {{.ServiceName}}", key)
		err = kv.Put(key, nil, nil)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func strFirstToUpper(str string) string {

	var upperStr string
	vv := []rune(str)
	for i := 0; i < len(vv); i++ {
		if i == 0 {
			if vv[i] >= 97 && vv[i] <= 122 {
				vv[i] -= 32
			}
			upperStr += string(vv[i])
		} else {
			upperStr += string(vv[i])
		}
	}
	return upperStr
}
`