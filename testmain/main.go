package main

import (
	"flag"
	"fmt"
	//"git.apache.org/thrift.git/lib/go/thrift"
	"git.oschina.net/kuaishangtong/common/utils/log"
	"git.oschina.net/kuaishangtong/navi/testmain/navi"
	"os"
	"reflect"
	//"strings"
)

func Usage() {
	fmt.Fprint(os.Stderr, "Usage of ", os.Args[0], ":\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
}

func main() {
	//flag.Usage = Usage
	//addr := flag.String("addr", "localhost:9191", "Address to listen to")
	//flag.Parse()
	//
	//var protocolFactory thrift.TProtocolFactory
	//protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	//
	//var transportFactory thrift.TTransportFactory
	//transportFactory = thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	//
	//transport, err := thrift.NewTServerSocket(*addr)
	//if err != nil {
	//	log.Fatal(err)
	//}

	var service *navi.NaviService = new(navi.NaviService)

	numMethod := reflect.TypeOf(service).Elem().NumMethod()

	for i := 0; i < numMethod; i++ {
		funcName := reflect.TypeOf(service).Elem().Method(i).Name
		numParams := reflect.TypeOf(service).Elem().Method(i).Type.NumIn()

		params := ""
		for j := 0; j < numParams; j++ {
			params += fmt.Sprintf("arg%d %s, ", j+1, reflect.TypeOf(service).Elem().Method(i).Type.In(j).String())
		}
		if len(params) > 0 {
			params = params[:len(params)-2]
			params = "(" + params + ")"
		} else {
			params = "()"
		}

		numRets := reflect.TypeOf(service).Elem().Method(i).Type.NumOut()

		rets := ""
		for k := 0; k < numRets; k++ {
			rets += fmt.Sprintf("%s, ", reflect.TypeOf(service).Elem().Method(i).Type.Out(k))
		}
		if len(rets) > 0 {
			rets = rets[:len(rets)-2]
			rets = " (" + rets + ")"
		}

		method := funcName + params + rets
		log.Debug(method)
		//method := funcName + strings.Split(params, "func")[1]
		//log.Debug(method)
	}

	//handler := new(NaviService)
	//processor := navi.NewNaviServiceProcessor(handler)
	//server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)
	//log.Infof("navi-test start ...")
	//log.Fatal(server.Serve())
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
