package api

import (
	"git.oschina.net/kuaishangtong/common/utils/log"
	"git.oschina.net/kuaishangtong/navi/gateway/constants"
	"git.oschina.net/kuaishangtong/navi/lb"
	"git.oschina.net/kuaishangtong/navi/registry"
	"time"
)

type Api struct {
	Name             string
	Cluster          *ServiceCluster
	prodURLs         registry.ServiceDiscovery
	devURLs          registry.ServiceDiscovery
	ProdServerUrlMap map[string]struct{}
	DevServerUrlMap  map[string]struct{}
	closed           bool
}

func NewApi(name string, lbmode lb.SelectMode) (*Api, error) {
	api := &Api{
		Name:             name,
		ProdServerUrlMap: make(map[string]struct{}),
		closed:           false,
	}

	var err error

	/* 获取 生产版本 /prod  url api接口 */
	api.prodURLs, err = registry.NewZookeeperDiscovery(constants.URLServicePath, name+"/prod", constants.ZookeeperHosts, nil)
	if err != nil {
		return nil, err
	}

	pairs := api.prodURLs.GetServices()
	for _, kv := range pairs {
		api.ProdServerUrlMap[kv.Key] = struct{}{}
		log.Infof("service [%s] add prod api [/%s]", name, kv.Key)
	}

	/* 获取 开发版本 /dev  url api接口 */
	api.devURLs, err = registry.NewZookeeperDiscovery(constants.URLServicePath, name+"/dev", constants.ZookeeperHosts, nil)
	if err != nil {
		return nil, err
	}

	pairs = api.prodURLs.GetServices()
	for _, kv := range pairs {
		api.DevServerUrlMap[kv.Key] = struct{}{}
		log.Infof("service [%s] add dev api [/%s]", name, kv.Key)
	}

	api.Cluster = NewServiceCluster(name).SetApi(api)

	err = api.Cluster.Discovery(constants.HTTPServicePath, name, constants.ZookeeperHosts, nil)
	if err != nil {
		return nil, err
	}

	prodselecter := lb.NewSelector(lbmode, nil)
	api.Cluster.SetProdSelector(prodselecter)
	devselecter := lb.NewSelector(lbmode, nil)
	api.Cluster.SetDevSelector(devselecter)

	err = api.Cluster.Commit()
	if err != nil {
		return nil, err
	}

	go api.watchURLs()

	return api, nil
}

func (this *Api) watchURLs() {
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
}

func (this *Api) Close() {
	this.closed = true
}
