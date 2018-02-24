package api

import (
	"context"
	"git.oschina.net/kuaishangtong/navi/gateway/constants"
	"git.oschina.net/kuaishangtong/navi/lb"
	"git.oschina.net/kuaishangtong/navi/registry"
	"github.com/docker/libkv/store"
	"strings"
)

type ServiceCluster struct {
	Name string
	Api  *Api

	prodServerIps map[string]string
	devServerIps  map[string]string

	prodIpDiscovery registry.ServiceDiscovery
	devIpDiscovery  registry.ServiceDiscovery

	prodSelector lb.Selector
	devSelector  lb.Selector

	prod_ch chan []*registry.KVPair
	dev_ch  chan []*registry.KVPair
}

func NewServiceCluster(name string) *ServiceCluster {
	return &ServiceCluster{
		Name:          name,
		prodServerIps: make(map[string]string, 10),
		devServerIps:  make(map[string]string, 10),
	}
}

func (sc *ServiceCluster) SetApi(api *Api) *ServiceCluster {
	sc.Api = api
	return sc
}

func (sc *ServiceCluster) SetProdSelector(s lb.Selector) *ServiceCluster {
	s.UpdateServer(sc.GetProdServices())
	sc.prodSelector = s
	return sc
}

func (sc *ServiceCluster) SetDevSelector(s lb.Selector) *ServiceCluster {
	s.UpdateServer(sc.GetDevServices())
	sc.devSelector = s
	return sc
}

// 发现服务集群IP
func (sc *ServiceCluster) Discovery(basePath string, servicePath string, zkAddr []string, options *store.Config) error {
	var err error
	sc.prodIpDiscovery, err = registry.NewZookeeperDiscovery(basePath, servicePath+"/prod", zkAddr, options)
	if err != nil {
		return err
	}

	sc.devIpDiscovery, err = registry.NewZookeeperDiscovery(basePath, servicePath+"/dev", zkAddr, options)
	return err
}

func (sc *ServiceCluster) Commit() error {
	go sc.prodServiceDiscovery()
	go sc.devServiceDiscovery()
	return nil
}

func (sc *ServiceCluster) Select(servicePath, serviceMethod, last_select, mode string) string {
	if strings.EqualFold(mode, constants.DEV_MODE) {
		return sc.devSelector.Select(context.Background(), servicePath, serviceMethod, last_select, nil)
	}

	return sc.prodSelector.Select(context.Background(), servicePath, serviceMethod, last_select, nil)
}

func (sc *ServiceCluster) GetProdServices() map[string]string {
	kvpairs := sc.prodIpDiscovery.GetServices()
	prodServerIps := make(map[string]string)
	for _, p := range kvpairs {
		prodServerIps[p.Key] = p.Value
	}
	return prodServerIps
}

func (sc *ServiceCluster) GetDevServices() map[string]string {
	kvpairs := sc.devIpDiscovery.GetServices()
	devServerIps := make(map[string]string)
	for _, p := range kvpairs {
		devServerIps[p.Key] = p.Value
	}
	return devServerIps
}

func (sc *ServiceCluster) prodServiceDiscovery() {
	ch := sc.prodIpDiscovery.WatchService()
	if ch != nil {
		sc.prod_ch = ch
	}

	for pairs := range ch {
		prodServerIps := make(map[string]string)
		for _, p := range pairs {
			prodServerIps[p.Key] = p.Value
		}

		sc.prodServerIps = prodServerIps

		if sc.prodSelector != nil {
			sc.prodSelector.UpdateServer(prodServerIps)
		}
	}
}

func (sc *ServiceCluster) devServiceDiscovery() {
	ch := sc.devIpDiscovery.WatchService()
	if ch != nil {
		sc.prod_ch = ch
	}

	for pairs := range ch {
		devServerIps := make(map[string]string)
		for _, p := range pairs {
			devServerIps[p.Key] = p.Value
		}

		sc.devServerIps = devServerIps

		if sc.devSelector != nil {
			sc.devSelector.UpdateServer(devServerIps)
		}
	}
}

func (sc *ServiceCluster) ProdServerCount() int {
	return len(sc.prodServerIps)
}

func (sc *ServiceCluster) DevServerCount() int {
	return len(sc.devServerIps)
}

func (sc *ServiceCluster) Close() {
	sc.prodIpDiscovery.Close()
}
