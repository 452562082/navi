package service

import (
	"git.oschina.net/kuaishangtong/common/utils/log"
	"git.oschina.net/kuaishangtong/navi/gateway/constants"
	"git.oschina.net/kuaishangtong/navi/lb"
	"git.oschina.net/kuaishangtong/navi/registry"
	"strings"
	"time"
)

type Service struct {
	Name    string
	Cluster *ServiceCluster

	prodApiURLs registry.ServiceDiscovery
	devApiURLs  registry.ServiceDiscovery

	prodApiUrlMap map[string]struct{}
	devApiUrlMap  map[string]struct{}

	closed bool
}

func NewService(name string, lbmode lb.SelectMode) (*Service, error) {
	srv := &Service{
		Name:          name,
		prodApiUrlMap: make(map[string]struct{}),
		closed:        false,
	}

	var err error

	/* 获取 生产版本 /prod  api url */
	srv.prodApiURLs, err = registry.NewZookeeperDiscovery(constants.URLServicePath, name+"/prod", constants.ZookeeperHosts, nil)
	if err != nil {
		return nil, err
	}

	pairs := srv.prodApiURLs.GetServices()
	for _, kv := range pairs {
		srv.prodApiUrlMap[kv.Key] = struct{}{}
		log.Infof("service [%s] add prod api url [/%s]", name, kv.Key)
	}

	/* 获取 开发版本 /dev api url */
	srv.devApiURLs, err = registry.NewZookeeperDiscovery(constants.URLServicePath, name+"/dev", constants.ZookeeperHosts, nil)
	if err != nil {
		return nil, err
	}

	pairs = srv.devApiURLs.GetServices()
	for _, kv := range pairs {
		srv.devApiUrlMap[kv.Key] = struct{}{}
		log.Infof("service [%s] add dev api url [/%s]", name, kv.Key)
	}

	srv.Cluster = NewServiceCluster(name, srv)

	err = srv.Cluster.Discovery(constants.HTTPServicePath, name, constants.ZookeeperHosts, nil)
	if err != nil {
		return nil, err
	}

	prodselecter := lb.NewSelector(lbmode, nil)
	srv.Cluster.SetProdSelector(prodselecter)

	devselecter := lb.NewSelector(lbmode, nil)
	srv.Cluster.SetDevSelector(devselecter)

	err = srv.Cluster.Commit()
	if err != nil {
		return nil, err
	}

	go srv.watchURLs()

	return srv, nil
}

func (this *Service) ExistApi(api string, mode string) bool {

	if strings.EqualFold(mode, constants.DEV_MODE) {
		_, exist := this.devApiUrlMap[api]
		return exist
	} else if strings.EqualFold(mode, constants.PROD_MODE) {
		_, exist := this.prodApiUrlMap[api]
		return exist
	}

	return false
}

func (this *Service) GetServerCount(mode string) int {
	if strings.EqualFold(mode, constants.DEV_MODE) {
		return len(this.Cluster.devServerIps)
	} else if strings.EqualFold(mode, constants.PROD_MODE) {
		return len(this.Cluster.prodServerIps)
	}
	return 0
}

// 监听 service 的 api 更变
func (this *Service) watchURLs() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for !this.closed {
		select {
		// 生产版本 /prod
		case p := <-this.prodApiURLs.WatchService():
			prodApiUrlMap := make(map[string]struct{})
			for _, kv := range p {
				prodApiUrlMap[kv.Key] = struct{}{}
			}
			this.prodApiUrlMap = prodApiUrlMap

			// 开发版本 /dev
		case p := <-this.devApiURLs.WatchService():
			devApiUrlMap := make(map[string]struct{})
			for _, kv := range p {
				devApiUrlMap[kv.Key] = struct{}{}
			}
			this.devApiUrlMap = devApiUrlMap

		case <-ticker.C:
		}
	}

	this.prodApiURLs.Close()
	this.devApiURLs.Close()
}

func (this *Service) Close() {
	this.closed = true
}
