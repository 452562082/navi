package agent

import (
	"git.oschina.net/kuaishangtong/navi/registry"
	metrics "github.com/rcrowley/go-metrics"
	"testing"
)

func TestAgent_Serve(t *testing.T) {

	address := "localhost:9191"
	a, err := NewAgent(address)
	if err != nil {
		t.Fatal()
	}

	typ, err := a.ServiceType()
	if err != nil {
		t.Fatal(err)
	}

	r := &registry.ZooKeeperRegister{
		ServiceAddress:   typ + "@" + address,
		ZooKeeperServers: []string{"192.168.1.16:2181"},
		BasePath:         "/rpcx_test",
		Metrics:          metrics.NewRegistry(),
	}

	err = r.Start()
	if err != nil {
		t.Fatal(err)
	}

	a.Plugins.Add(r)

	a.Serve()
}
