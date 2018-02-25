package service

import (
	"git.oschina.net/kuaishangtong/common/utils/log"
	"git.oschina.net/kuaishangtong/navi/gateway/constants"
	"git.oschina.net/kuaishangtong/navi/lb"
	"git.oschina.net/kuaishangtong/navi/registry"
	"sync"
)

var GlobalServiceManager *ServiceManager

func InitServiceManager() error {
	var err error
	GlobalServiceManager, err = newServiceManager()
	return err
}

// ServiceManager 用于动态管理 gateway 可支持的services
// services信息存储在zookeeper上，后续由平台的管理后台来管理
type ServiceManager struct {
	services     map[string]*Service
	apiDiscovery registry.ServiceDiscovery
	lock         *sync.RWMutex
}

func newServiceManager() (*ServiceManager, error) {
	srvManager := &ServiceManager{
		services: make(map[string]*Service, 10),
		lock:     new(sync.RWMutex),
	}

	var err error
	srvManager.apiDiscovery, err = registry.NewZookeeperDiscovery(constants.URLServicePath, "", constants.ZookeeperHosts, nil)
	if err != nil {
		return nil, err
	}
	log.Infof("Discovery Service list in %s", constants.URLServicePath)

	err = srvManager.init()
	if err != nil {
		return nil, err
	}

	return srvManager, nil
}

func (this *ServiceManager) init() error {
	pairs := this.apiDiscovery.GetServices()

	for _, kv := range pairs {
		api, err := NewService(kv.Key, lb.RoundRobin)
		if err != nil {
			return err
		}
		this.AddService(kv.Key, api)
	}

	go this.watch()
	return nil
}

func (this *ServiceManager) watch() {
	for {
		select {
		case p := <-this.apiDiscovery.WatchService():
			// 动态刷新API
			var oldmap, newmap map[string]struct{} = make(map[string]struct{}), make(map[string]struct{})
			for k, _ := range this.services {
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
				this.DelService(key)
			}

			for key, _ := range newmap {
				api, err := NewService(key, lb.RoundRobin)
				if err != nil {
					log.Error(err)
					continue
				}

				this.AddService(key, api)
			}
		}
	}
}

func (this *ServiceManager) AddService(name string, api *Service) {
	this.lock.Lock()
	this.services[name] = api
	log.Infof("Add Service [%s]", name)
	this.lock.Unlock()
}

func (this *ServiceManager) DelService(name string) {
	api := this.GetService(name)
	if api != nil {
		api.Close()
	}

	this.lock.Lock()
	delete(this.services, name)
	log.Infof("Del Service [%s]", name)
	this.lock.Unlock()
}

func (this *ServiceManager) GetService(name string) *Service {
	this.lock.RLock()
	if api, ok := this.services[name]; ok {
		this.lock.RUnlock()
		return api
	}
	this.lock.RUnlock()
	return nil
}
