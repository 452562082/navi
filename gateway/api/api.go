package api

import (
	"git.oschina.net/kuaishangtong/navi/gateway/cluster"
	"git.oschina.net/kuaishangtong/navi/gateway/constants"
	"git.oschina.net/kuaishangtong/navi/lb"
	"git.oschina.net/kuaishangtong/navi/registry"
	"time"
)

type Api struct {
	Name         string
	Cluster      *cluster.ServiceCluster
	urlDiscovery registry.ServiceDiscovery
	ServerURLs   map[string]struct{}
	closed       bool
}

func NewApi(name string, lbmode lb.SelectMode) (*Api, error) {
	api := &Api{
		Name:       name,
		ServerURLs: make(map[string]struct{}),
		closed:     false,
	}

	var err error

	api.urlDiscovery, err = registry.NewZookeeperDiscovery(constants.URLServicePath, name, constants.ZookeeperHosts, nil)
	if err != nil {
		return nil, err
	}

	pairs := api.urlDiscovery.GetServices()
	for _, kv := range pairs {
		api.ServerURLs[kv.Key] = struct{}{}
	}

	api.Cluster = cluster.NewServiceCluster(name).SetApi(api)

	err = api.Cluster.Discovery(constants.URLServicePath, name, constants.ZookeeperHosts, nil)
	if err != nil {
		return nil, err
	}

	selecter := lb.NewSelector(lbmode, nil)
	api.Cluster.SetSelector(selecter)

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
		case p := <-this.urlDiscovery.WatchService():
			serverURLs := make(map[string]struct{})
			for _, kv := range p {
				serverURLs[kv.Key] = struct{}{}
			}
			this.ServerURLs = serverURLs

		case <-ticker.C:
		}
	}

	this.urlDiscovery.Close()
}

func (this *Api) Close() {
	this.closed = true
}
