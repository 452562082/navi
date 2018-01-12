package cluster

import (
	"context"
	"git.oschina.net/kuaishangtong/navi/gateway/api"
	"git.oschina.net/kuaishangtong/navi/lb"
	"git.oschina.net/kuaishangtong/navi/registry"
	"github.com/docker/libkv/store"
)

type ServiceCluster struct {
	Name      string
	Api       *api.Api
	servers   map[string]string
	discovery registry.ServiceDiscovery
	selector  lb.Selector
	ch        chan []*registry.KVPair
}

func NewServiceCluster(name string) *ServiceCluster {
	return &ServiceCluster{
		Name:    name,
		servers: make(map[string]string, 10),
	}
}

func (sc *ServiceCluster) SetApi(api *api.Api) *ServiceCluster {
	sc.Api = api
	return sc
}

func (sc *ServiceCluster) SetSelector(s lb.Selector) *ServiceCluster {
	s.UpdateServer(sc.GetServices())
	sc.selector = s
	return sc
}

// 发现服务集群IP
func (sc *ServiceCluster) Discovery(basePath string, servicePath string, zkAddr []string, options *store.Config) error {
	var err error
	sc.discovery, err = registry.NewZookeeperDiscovery(basePath, servicePath, zkAddr, options)
	return err
}

func (sc *ServiceCluster) Commit() error {
	go sc.serviceDiscovery()
	return nil
}

func (sc *ServiceCluster) Select(servicePath, serviceMethod string) string {
	return sc.selector.Select(context.Background(), servicePath, serviceMethod, nil)
}

func (sc *ServiceCluster) GetServices() map[string]string {
	kvpairs := sc.discovery.GetServices()
	servers := make(map[string]string)
	for _, p := range kvpairs {
		servers[p.Key] = p.Value
	}
	return servers
}

func (sc *ServiceCluster) serviceDiscovery() {
	ch := sc.discovery.WatchService()
	if ch != nil {
		sc.ch = ch
	}

	for pairs := range ch {
		servers := make(map[string]string)
		for _, p := range pairs {
			servers[p.Key] = p.Value
		}

		sc.servers = servers

		if sc.selector != nil {
			sc.selector.UpdateServer(servers)
		}
	}
}

func (sc *ServiceCluster) Close() {
	sc.discovery.Close()
}
