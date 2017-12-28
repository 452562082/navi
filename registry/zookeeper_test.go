package registry

import (
	"testing"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/smallnest/rpcx/server"
)

import "context"

type Args struct {
	A int
	B int
}

type Reply struct {
	C int
}

type Arith int

func (t *Arith) Mul(ctx context.Context, args *Args, reply *Reply) error {
	reply.C = args.A * args.B
	return nil
}

func TestZookeeperRegistry(t *testing.T) {
	s := server.NewServer()

	r := &ZooKeeperRegister{
		ServiceAddress:   "tcp@127.0.0.1:8972",
		ZooKeeperServers: []string{"192.168.1.16:2181"},
		BasePath:         "/rpcx_test",
		Metrics:          metrics.NewRegistry(),
		UpdateInterval:   time.Minute,
	}
	err := r.Start()
	if err != nil {
		t.Fatal(err)
	}
	s.Plugins.Add(r)

	s.RegisterName("Arith", new(Arith), "")
	go s.Serve("tcp", "127.0.0.1:8972")
	defer s.Close()

	if len(r.Services) != 1 {
		t.Fatal("failed to register services in zookeeper")
	}

	select {

	}

}
