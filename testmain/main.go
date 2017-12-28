package main

import (
	"flag"
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"git.oschina.net/kuaishangtong/common/utils/log"
	"git.oschina.net/kuaishangtong/navi/testmain/navitest"
	"os"
)

func Usage() {
	fmt.Fprint(os.Stderr, "Usage of ", os.Args[0], ":\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
}

func main() {
	flag.Usage = Usage
	addr := flag.String("addr", "localhost:9191", "Address to listen to")
	flag.Parse()

	var protocolFactory thrift.TProtocolFactory
	protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()

	var transportFactory thrift.TTransportFactory
	transportFactory = thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())

	transport, err := thrift.NewTServerSocket(*addr)
	if err != nil {
		log.Fatal(err)
	}

	handler := new(NaviService)
	processor := navi.NewNaviServiceProcessor(handler)
	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)
	log.Infof("navi-test start ...")
	log.Fatal(server.Serve())
}

type NaviService struct {
	host string
}

func (ks *NaviService) Ping() (r string, err error) {
	return "pong", nil
}

func (ks *NaviService) ServiceName() (r string, err error) {
	return "navi-test", nil
}

func (ks *NaviService) ServiceType() (r string, err error) {
	return "rpc", nil
}
