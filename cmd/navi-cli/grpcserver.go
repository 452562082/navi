package navicli

import (
	"net"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"kuaishangtong/common/utils/log"
)

type GrpcServer struct {
	*Server
	gClient    *grpcClient
	grpcServer *grpc.Server
}

func NewGrpcServer(initializer Initializable, configFilePath string) *GrpcServer {
	if initializer == nil {
		initializer = &defaultInitializer{}
	}
	s := &GrpcServer{
		Server: &Server{
			Config:       NewConfig("grpc", configFilePath),
			Components:   new(Components),
			reloadConfig: make(chan bool),
			Initializer:  initializer,
		},
		gClient: new(grpcClient),
	}
	s.initChans()
	//initLogger(s.Config)
	return s
}

type grpcClientCreator func(conn *grpc.ClientConn) interface{}

// Start starts both HTTP server and GRPC service
func (s *GrpcServer) Start(clientCreator grpcClientCreator, sw switcher, registerServer func(s *grpc.Server)) {
	log.Info("Starting Turbo...")
	s.Initializer.InitService(s)
	s.grpcServer = s.startGrpcServiceInternal(registerServer, false)
	s.httpServer = s.startGrpcHTTPServerInternal(clientCreator, sw)
	watchConfigReload(s)
}

// StartHTTPServer starts a HTTP server which sends requests via grpc
func (s *GrpcServer) StartHTTPServer(clientCreator grpcClientCreator, sw switcher) {
	s.Initializer.InitService(s)
	s.httpServer = s.startGrpcHTTPServerInternal(clientCreator, sw)
	watchConfigReload(s)
}

// StartGrpcService starts a GRPC service
func (s *GrpcServer) StartGrpcService(registerServer func(s *grpc.Server)) {
	s.Initializer.InitService(s)
	s.grpcServer = s.startGrpcServiceInternal(registerServer, true)
}

func (s *GrpcServer) startGrpcHTTPServerInternal(clientCreator grpcClientCreator, sw switcher) *http.Server {
	log.Info("Starting HTTP Server...")
	switcherFunc = sw
	//TODO register multi gClients
	s.gClient.init(s.Config.GrpcServiceHost()+":"+s.Config.GrpcServicePort(), clientCreator)
	return startHTTPServer(s)
}

func (s *GrpcServer) startGrpcServiceInternal(registerServer func(s *grpc.Server), alone bool) *grpc.Server {
	log.Info("Starting GRPC Service...")
	lis, err := net.Listen("tcp", ":"+s.Config.GrpcServicePort())
	logPanicIf(err)
	grpcServer := grpc.NewServer()
	registerServer(grpcServer)
	reflection.Register(grpcServer)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Errorf("GRPC Service failed to serve: %v", err)
		}
	}()
	log.Info("GRPC Service started")
	return grpcServer
}

// GrpcService returns a grpc client instance,
// example: client := turbo.GrpcService().(proto.YourServiceClient)
func (s *GrpcServer) Service() interface{} {
	if s == nil || s.gClient == nil || s.gClient.grpcService == nil {
		log.Fatal("grpc connection not initiated!")
	}
	return s.gClient.grpcService
}
func (s *GrpcServer) ServerField() *Server { return s.Server }

func (s *GrpcServer) Stop() {
	log.Info("Stop() invoked, Service is stopping...")
	stop(s, s.httpServer, s.grpcServer, nil)
}
