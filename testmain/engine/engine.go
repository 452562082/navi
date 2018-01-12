package engine

import (
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"git.oschina.net/kuaishangtong/navi/testmain/navi"
	"time"
)

var XEngine *Engine // 全局引擎

var gTimeout int

type Engine struct {
	host string
	port string
}

func InitEngine(host, port string, timeout int) {
	gTimeout = timeout
	XEngine = &Engine{
		host: host,
		port: port,
	}
	return
}

func (c *Engine) shortConn() (*navi.NaviServiceClient, error) {

	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()

	transport, err := thrift.NewTSocketTimeout(fmt.Sprintf("%s:%s", c.host, c.port), time.Duration(gTimeout)*time.Second)
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
