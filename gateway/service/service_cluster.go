package service

import (
	"context"
	"git.oschina.net/kuaishangtong/navi/gateway/constants"
	"git.oschina.net/kuaishangtong/navi/lb"
	"git.oschina.net/kuaishangtong/navi/registry"
	"github.com/docker/libkv/store"
	"strings"
)

// ServiceCluster 主要用于管理对应 Service 下的集群
// 包括 prod , dev 两个版本的服务机器的发现功能，还有负载均衡功能
type ServiceCluster struct {
	Name    string
	service *Service // 集群所从属的服务

	prodServerIps map[string]string
	devServerIps  map[string]string

	prodIpDiscovery registry.ServiceDiscovery
	devIpDiscovery  registry.ServiceDiscovery

	prodSelector lb.Selector
	devSelector  lb.Selector

	prod_ch chan []*registry.KVPair
	dev_ch  chan []*registry.KVPair
}

func NewServiceCluster(name string, service *Service) *ServiceCluster {
	return &ServiceCluster{
		Name:          name,
		service:       service,
		prodServerIps: make(map[string]string, 10),
		devServerIps:  make(map[string]string, 10),
	}
}

func (sc *ServiceCluster) SetProdSelector(s lb.Selector) *ServiceCluster {
	s.UpdateServer(sc.GetProdServers())
	sc.prodSelector = s
	return sc
}

func (sc *ServiceCluster) SetDevSelector(s lb.Selector) *ServiceCluster {
	s.UpdateServer(sc.GetDevServers())
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
	go sc.prodServerDiscovery()
	go sc.devServerDiscovery()
	return nil
}

func (sc *ServiceCluster) Select(servicePath, serviceMethod, last_select, mode string) string {
	if strings.EqualFold(mode, constants.DEV_MODE) {
		return sc.devSelector.Select(context.Background(), servicePath, serviceMethod, last_select, nil)
	}

	return sc.prodSelector.Select(context.Background(), servicePath, serviceMethod, last_select, nil)
}

func (sc *ServiceCluster) GetProdServers() map[string]string {
	kvpairs := sc.prodIpDiscovery.GetServices()
	prodServerIps := make(map[string]string)
	for _, p := range kvpairs {
		prodServerIps[p.Key] = p.Value
	}
	return prodServerIps
}

func (sc *ServiceCluster) GetDevServers() map[string]string {
	kvpairs := sc.devIpDiscovery.GetServices()
	devServerIps := make(map[string]string)
	for _, p := range kvpairs {
		devServerIps[p.Key] = p.Value
	}
	return devServerIps
}

// 监听zookeeper，发现 服务 prod 版本的机器
func (sc *ServiceCluster) prodServerDiscovery() {
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

// 监听zookeeper，发现 服务 dev 版本的机器
func (sc *ServiceCluster) devServerDiscovery() {
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


func (sc *ServiceCluster) Close() {
	sc.prodIpDiscovery.Close()
}
