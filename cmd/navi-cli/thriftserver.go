package navicli

import (
	"net/http"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"git.oschina.net/kuaishangtong/common/utils/log"
)

type ThriftServer struct {
	*Server
	tClient      *thriftClient
	thriftServer *thrift.TSimpleServer
}

func NewThriftServer(initializer Initializable, configFilePath string) *ThriftServer {
	if initializer == nil {
		initializer = &defaultInitializer{}
	}
	s := &ThriftServer{
		Server: &Server{
			Config:       NewConfig("thrift", configFilePath),
			Components:   new(Components),
			reloadConfig: make(chan bool),
			Initializer:  initializer,
		},
		tClient: new(thriftClient),
	}
	s.initChans()
	return s
}

type thriftClientCreator func(trans thrift.TTransport, f thrift.TProtocolFactory) interface{}

// Start starts both HTTP server and Thrift service
func (s *ThriftServer) Start(clientCreator thriftClientCreator, sw switcher,
	registerTProcessor func() thrift.TProcessor) {
	log.Infof("Starting %s...", s.Config.ThriftServiceName())
	s.Initializer.InitService(s)
	s.thriftServer = s.startThriftServiceInternal(registerTProcessor, false)
	time.Sleep(time.Second * 1)
	s.httpServer = s.startThriftHTTPServerInternal(clientCreator, sw)
	watchConfigReload(s)
}

// StartHTTPServer starts a HTTP server which sends requests via Thrift
func (s *ThriftServer) StartHTTPServer(clientCreator thriftClientCreator, sw switcher) {
	s.Initializer.InitService(s)
	s.httpServer = s.startThriftHTTPServerInternal(clientCreator, sw)
	watchConfigReload(s)
}

// StartThriftService starts a Thrift service
func (s *ThriftServer) StartThriftService(registerTProcessor func() thrift.TProcessor) {
	s.Initializer.InitService(s)
	s.thriftServer = s.startThriftServiceInternal(registerTProcessor, true)
}

func (s *ThriftServer) startThriftHTTPServerInternal(clientCreator thriftClientCreator, sw switcher) *http.Server {
	log.Info("Starting HTTP Server...")
	switcherFunc = sw
	s.tClient.init(s.Config.ThriftServiceHost()+":"+s.Config.ThriftServicePort(), clientCreator)
	return startHTTPServer(s)
}

func (s *ThriftServer) startThriftServiceInternal(registerTProcessor func() thrift.TProcessor, alone bool) *thrift.TSimpleServer {
	port := s.Config.ThriftServicePort()
	log.Infof("Starting Thrift Service at :%s...", port)
	transport, err := thrift.NewTServerSocket(":" + port)
	logPanicIf(err)
	server := thrift.NewTSimpleServer4(registerTProcessor(), transport,
		thrift.NewTTransportFactory(), thrift.NewTBinaryProtocolFactoryDefault())
	go server.Serve()
	log.Info("Thrift Service started")
	return server
}

// ThriftService returns a Thrift client instance,
// example: client := turbo.ThriftService().(proto.YourServiceClient)
func (s *ThriftServer) Service() interface{} {
	if s == nil || s.tClient == nil || s.tClient.thriftService == nil {
		log.Fatal("thrift connection not initiated!")
	}
	return s.tClient.thriftService
}

func (s *ThriftServer) ServerField() *Server { return s.Server }

func (s *ThriftServer) Stop() {
	stop(s, s.httpServer, nil, s.thriftServer)
}
