package service

import (
	"context"
	"fmt"
	"git.oschina.net/kuaishangtong/common/utils/log"
	"git.oschina.net/kuaishangtong/navi/gateway/constants"
	"git.oschina.net/kuaishangtong/navi/lb"
	"git.oschina.net/kuaishangtong/navi/registry"
	"github.com/docker/libkv/store"
	"strings"
	"time"
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

	closed bool
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
	servers := sc.getProdServers()
	s.UpdateServer(servers)
	for ip, _ := range servers {
		log.Infof("service [%s] cluster add prod server ip %s", sc.service.Name, ip)
	}

	sc.prodSelector = s
	return sc
}

func (sc *ServiceCluster) SetDevSelector(s lb.Selector) *ServiceCluster {
	servers := sc.getDevServers()
	s.UpdateServer(servers)
	for ip, _ := range servers {
		log.Infof("service [%s] cluster add dev server ip %s", sc.service.Name, ip)
	}

	sc.devSelector = s
	return sc
}

// 发现服务集群IP
func (sc *ServiceCluster) Discovery(basePath string, servicePath string, zkAddr []string, options *store.Config) error {
	var err1, err2 error
	sc.prodIpDiscovery, err1 = registry.NewZookeeperDiscovery(basePath, servicePath+"/prod", zkAddr, options)

	sc.devIpDiscovery, err2 = registry.NewZookeeperDiscovery(basePath, servicePath+"/dev", zkAddr, options)

	if err1 != nil && err2 != nil {
		log.Errorf("cat not find service %s [prod] or [dev] IP err: %v; %v", servicePath, err1, err2)
		return fmt.Errorf("cat not find service %s [prod] or [dev] IP err: %v; %v", servicePath, err1, err2)
	}
	return nil
}

func (sc *ServiceCluster) Commit() error {
	go sc.srverDiscovery()
	return nil
}

func (sc *ServiceCluster) Select(servicePath, serviceMethod, last_select, mode string) string {
	if strings.EqualFold(mode, constants.DEV_MODE) {
		return sc.devSelector.Select(context.Background(), servicePath, serviceMethod, last_select, nil)
	}

	return sc.prodSelector.Select(context.Background(), servicePath, serviceMethod, last_select, nil)
}

func (sc *ServiceCluster) getProdServers() map[string]string {
	kvpairs := sc.prodIpDiscovery.GetServices()
	prodServerIps := make(map[string]string)
	for _, p := range kvpairs {
		prodServerIps[p.Key] = p.Value
	}
	return prodServerIps
}

func (sc *ServiceCluster) getDevServers() map[string]string {
	kvpairs := sc.devIpDiscovery.GetServices()
	devServerIps := make(map[string]string)
	for _, p := range kvpairs {
		devServerIps[p.Key] = p.Value
	}
	return devServerIps
}

// 监听zookeeper，发现 服务 prod 版本的机器
func (sc *ServiceCluster) srverDiscovery() {

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for !sc.closed {
		select {
		// 监听zookeeper，发现服务 prod版本的机器更变
		case prod_ch := <-sc.prodIpDiscovery.WatchService():
			prodServerIps := make(map[string]string)
			ips := make([]string, 0, len(prod_ch))
			for _, p := range prod_ch {
				prodServerIps[p.Key] = p.Value
				ips = append(ips, p.Key)
			}

			sc.prodServerIps = prodServerIps

			if sc.prodSelector != nil {
				sc.prodSelector.UpdateServer(prodServerIps)
				log.Infof("service [%s] cluster update prod servers %v", sc.service.Name, ips)
			}

			// 监听zookeeper，发现服务 dev 版本的机器更变
		case dev_ch := <-sc.devIpDiscovery.WatchService():
			devServerIps := make(map[string]string)
			ips := make([]string, 0, len(dev_ch))
			for _, p := range dev_ch {
				devServerIps[p.Key] = p.Value
				ips = append(ips, p.Key)
			}

			sc.devServerIps = devServerIps

			if sc.devSelector != nil {
				sc.devSelector.UpdateServer(devServerIps)
				log.Infof("service [%s] cluster update dev servers %v", sc.service.Name, ips)
			}
		case <-ticker.C:
		}
	}

}

func (sc *ServiceCluster) Close() {
	sc.closed = true

	sc.prodIpDiscovery.Close()
	sc.devIpDiscovery.Close()
}
