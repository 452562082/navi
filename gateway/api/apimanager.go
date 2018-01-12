package api

import (
	"git.oschina.net/kuaishangtong/common/utils/log"
	"git.oschina.net/kuaishangtong/navi/gateway/constants"
	"git.oschina.net/kuaishangtong/navi/lb"
	"git.oschina.net/kuaishangtong/navi/registry"
	"sync"
)

var GlobalApiManager *ApiManager

func Init() error {
	var err error
	GlobalApiManager, err = newApiManager()
	return err
}

type ApiManager struct {
	apis         map[string]*Api
	apiDiscovery registry.ServiceDiscovery
	lock         *sync.RWMutex
}

func newApiManager() (*ApiManager, error) {
	apiManager := &ApiManager{
		apis: make(map[string]*Api, 10),
		lock: new(sync.RWMutex),
	}

	var err error
	apiManager.apiDiscovery, err = registry.NewZookeeperDiscovery(constants.ServiceListPath, "", constants.ZookeeperHosts, nil)
	if err != nil {
		return nil, err
	}

	err = apiManager.init()
	if err != nil {
		return nil, err
	}

	return apiManager, nil
}

func (this *ApiManager) init() error {
	pairs := this.apiDiscovery.GetServices()

	for _, kv := range pairs {
		api, err := NewApi(kv.Key, lb.RoundRobin)
		if err != nil {
			return err
		}
		this.AddApi(kv.Key, api)
	}

	go this.watch()
	return nil
}

func (this *ApiManager) watch() {
	for {
		select {
		case p := <-this.apiDiscovery.WatchService():
			// 动态刷新API
			var oldmap, newmap map[string]struct{} = make(map[string]struct{}), make(map[string]struct{})
			for k, _ := range this.apis {
				oldmap[k] = struct{}{}
			}

			for _, kv := range p {
				newmap[kv.Key] = struct{}{}
			}

			for k, _ := range newmap {
				if _, ok := oldmap[k]; ok {
					delete(oldmap, k)
					delete(newmap, k)
				}
			}

			for key, _ := range oldmap {
				this.DelApi(key)
			}

			for key, _ := range newmap {
				api, err := NewApi(key, lb.RoundRobin)
				if err != nil {
					log.Error(err)
					continue
				}

				this.AddApi(key, api)
			}
		}
	}
}

func (this *ApiManager) AddApi(name string, api *Api) {
	this.lock.Lock()
	this.apis[name] = api
	this.lock.Unlock()
}

func (this *ApiManager) DelApi(name string) {
	api := this.GetApi(name)
	if api != nil {
		api.Close()
	}

	this.lock.Lock()
	delete(this.apis, name)
	this.lock.Unlock()
}

func (this *ApiManager) GetApi(name string) *Api {
	this.lock.RLock()
	if api, ok := this.apis[name]; ok {
		this.lock.RUnlock()
		return api
	}
	this.lock.RUnlock()
	return nil
}
