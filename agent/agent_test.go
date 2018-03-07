package agent

import (
	"kuaishangtong/navi/registry"
	metrics "github.com/rcrowley/go-metrics"
	"testing"
)

func TestAgent_Serve(t *testing.T) {

	address := "localhost:9191"
	typ := "rpc"
	a, err := NewAgent("mytest", address, "rpc")
	if err != nil {
		t.Fatal()
	}

	var basePath string
	switch typ {
	case "rpc":
		basePath = "/navi/rpcservice"
	case "http":
		basePath = "/navi/httpservice"
	}

	r := &registry.ZooKeeperRegister{
		ServiceAddress:   address,
		ZooKeeperServers: []string{"127.0.0.1:2181"},
		BasePath:         basePath,
		Metrics:          metrics.NewRegistry(),
	}

	err = r.Start()
	if err != nil {
		t.Fatal(err)
	}

	a.Plugins.Add(r)

	a.Serve()
}
