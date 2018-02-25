package service

import (
	"git.oschina.net/kuaishangtong/common/utils/log"
	"git.oschina.net/kuaishangtong/navi/gateway/constants"
	"git.oschina.net/kuaishangtong/navi/lb"
	"git.oschina.net/kuaishangtong/navi/registry"
	"time"
)

type Service struct {
	Name    string
	Cluster *ServiceCluster

	prodURLs registry.ServiceDiscovery
	devURLs  registry.ServiceDiscovery

	ProdServerUrlMap map[string]struct{}
	DevServerUrlMap  map[string]struct{}

	closed bool
}

func NewService(name string, lbmode lb.SelectMode) (*Service, error) {
	srv := &Service{
		Name:             name,
		ProdServerUrlMap: make(map[string]struct{}),
		closed:           false,
	}

	var err error

	/* 获取 生产版本 /prod  url */
	srv.prodURLs, err = registry.NewZookeeperDiscovery(constants.URLServicePath, name+"/prod", constants.ZookeeperHosts, nil)
	if err != nil {
		return nil, err
	}

	pairs := srv.prodURLs.GetServices()
	for _, kv := range pairs {
		srv.ProdServerUrlMap[kv.Key] = struct{}{}
		log.Infof("service [%s] add prod srv [/%s]", name, kv.Key)
	}

	/* 获取 开发版本 /dev  url */
	srv.devURLs, err = registry.NewZookeeperDiscovery(constants.URLServicePath, name+"/dev", constants.ZookeeperHosts, nil)
	if err != nil {
		return nil, err
	}

	pairs = srv.prodURLs.GetServices()
	for _, kv := range pairs {
		srv.DevServerUrlMap[kv.Key] = struct{}{}
		log.Infof("service [%s] add dev srv [/%s]", name, kv.Key)
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

func (this *Service) watchURLs() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for !this.closed {
		select {
		// 生产版本 /prod
		case p := <-this.prodURLs.WatchService():
			ProdServerUrlMap := make(map[string]struct{})
			for _, kv := range p {
				ProdServerUrlMap[kv.Key] = struct{}{}
			}
			this.ProdServerUrlMap = ProdServerUrlMap

			// 开发版本 /dev
		case p := <-this.devURLs.WatchService():
			ProdServerUrlMap := make(map[string]struct{})
			for _, kv := range p {
				ProdServerUrlMap[kv.Key] = struct{}{}
			}
			this.DevServerUrlMap = ProdServerUrlMap

		case <-ticker.C:
		}
	}

	this.prodURLs.Close()
	this.devURLs.Close()
}

func (this *Service) Close() {
	this.closed = true
}
