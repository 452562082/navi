package agent

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"kuaishangtong/navi/agent/navi_grpc"
	"strings"
)

func (a *Agent) NewGrpcAgenter() (g *grpcAgenter, err error) {

	grcConn, err := grpc.Dial(a.address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	g = &grpcAgenter{
		gc:      grcConn,
		service: a.serverName,
		nc:      navi_grpc.NewNaviClient(grcConn),
	}

	return g, nil
}

type grpcAgenter struct {
	gc      *grpc.ClientConn
	service string
	nc      navi_grpc.NaviClient
}

func (p *grpcAgenter) Close() error {
	return p.gc.Close()
}

func (p *grpcAgenter) Ping() (r string, err error) {
	pingResp, err := p.nc.Ping(context.Background(), &navi_grpc.PingRequest{}, strings.ToLower(p.service))
	if err != nil {
		return "", err
	}
	return pingResp.Pong, nil
}

func (p *grpcAgenter) ServiceName() (r string, err error) {
	serviceNameResp, err := p.nc.ServiceName(context.Background(), &navi_grpc.ServiceNameRequest{}, strings.ToLower(p.service))
	if err != nil {
		return "", err
	}
	return serviceNameResp.ServiceName, nil
}

func (p *grpcAgenter) ServiceMode() (r string, err error) {
	serviceModeResp, err := p.nc.ServiceMode(context.Background(), &navi_grpc.ServiceModeRequest{}, strings.ToLower(p.service))
	if err != nil {
		return "", err
	}
	return serviceModeResp.ServiceMode, nil
}
