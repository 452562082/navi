package service

import (
	"encoding/json"
	"fmt"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"kuaishangtong/common/utils/log"
	"kuaishangtong/navi/gateway/constants"
	"kuaishangtong/navi/lb"
	"strings"
	"time"
)

type Service struct {
	Name    string
	Cluster *ServiceCluster

	prodApiURLs store.Store
	devApiURLs  store.Store

	prodApiUrlMap map[string]struct{}
	devApiUrlMap  map[string]struct{}

	closed bool
}

func NewService(name string, lbmode lb.SelectMode) (*Service, error) {
	srv := &Service{
		Name:          name,
		prodApiUrlMap: make(map[string]struct{}),
		devApiUrlMap:  make(map[string]struct{}),
		closed:        false,
	}

	var err error

	/* 获取 生产版本 /prod api url */
	srv.prodApiURLs, err = libkv.NewStore(store.ZK, constants.ZookeeperHosts, nil)
	if err != nil {
		log.Errorf("cannot create store: %v", err)
		return nil, err
	}

	kv, err := srv.prodApiURLs.Get(fmt.Sprintf("%s/%s", constants.URLServicePath, name+"/prod"))
	if err != nil {
		log.Errorf("prodApiURLs cannot get kv err: %v", err)
		return nil, err
	}

	var urlInfo UrlInfo
	err = json.Unmarshal(kv.Value, &urlInfo)
	if err != nil {
		return nil, err
	}

	for _, apiurl := range urlInfo.ApiUrls {
		url := strings.Trim(apiurl.URL, "/")
		srv.prodApiUrlMap[url] = struct{}{}
		log.Infof("service [%s] add prod api url [/%s]", name, url)
	}

	//srv.prodApiURLs, err = registry.NewZookeeperDiscovery(constants.URLServicePath, name+"/prod", constants.ZookeeperHosts, nil)
	//if err != nil {
	//	return nil, err
	//}

	//pairs := srv.prodApiURLs.GetServices()
	//for _, kv := range pairs {
	//	srv.prodApiUrlMap[kv.Key] = struct{}{}
	//	log.Infof("service [%s] add prod api url [/%s]", name, kv.Key)
	//}

	/* 获取 开发版本 /dev api url */
	srv.devApiURLs, err = libkv.NewStore(store.ZK, constants.ZookeeperHosts, nil)
	if err != nil {
		log.Errorf("cannot create store: %v", err)
		return nil, err
	}

	kv, err = srv.devApiURLs.Get(fmt.Sprintf("%s/%s", constants.URLServicePath, name+"/dev"))
	if err != nil {
		log.Errorf("devApiURLs cannot get kv err: %v", err)
		return nil, err
	}

	err = json.Unmarshal(kv.Value, &urlInfo)
	if err != nil {
		return nil, err
	}

	for _, apiurl := range urlInfo.ApiUrls {
		url := strings.Trim(apiurl.URL, "/")
		srv.devApiUrlMap[url] = struct{}{}
		log.Infof("service [%s] add dev api url [/%s]", name, url)
	}

	//srv.devApiURLs, err = registry.NewZookeeperDiscovery(constants.URLServicePath, name+"/dev", constants.ZookeeperHosts, nil)
	//if err != nil {
	//	return nil, err
	//}
	//
	//pairs = srv.devApiURLs.GetServices()
	//for _, kv := range pairs {
	//	srv.devApiUrlMap[kv.Key] = struct{}{}
	//	log.Infof("service [%s] add dev api url [/%s]", name, kv.Key)
	//}

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

	stopCh1 := make(<-chan struct{})
	stopCh2 := make(<-chan struct{})

	prodEvent, err := this.prodApiURLs.Watch(fmt.Sprintf("%s/%s", constants.URLServicePath, this.Name+"/prod"), stopCh1)
	if err != nil {
		log.Errorf("prodApiURLs watchURLs err: %v", err)
		return
	}

	devEvent, err := this.devApiURLs.Watch(fmt.Sprintf("%s/%s", constants.URLServicePath, this.Name+"/dev"), stopCh2)
	if err != nil {
		log.Errorf("devApiURLs watchURLs err: %v", err)
		return
	}

	for !this.closed {
		select {

		// 生产版本 /prod
		case p := <-prodEvent:

			var urlInfo UrlInfo
			err = json.Unmarshal(p.Value, &urlInfo)
			if err != nil {
				log.Errorf("prodApiURLs watch json.Unmarshal err: %v", err)
				continue
			}

			prodApiUrlMap := make(map[string]struct{})
			urls := make([]string, 0, len(urlInfo.ApiUrls))
			for _, apiurl := range urlInfo.ApiUrls {
				url := strings.Trim(apiurl.URL, "/")
				prodApiUrlMap[url] = struct{}{}
				urls = append(urls, url)
				log.Infof("service [%s] add prod api url [/%s]", this.Name, url)
			}

			this.prodApiUrlMap = prodApiUrlMap
			log.Infof("service [%s] update prod api urls %v", this.Name, urls)
			// 开发版本 /dev
		case p := <-devEvent:

			var urlInfo UrlInfo
			err = json.Unmarshal(p.Value, &urlInfo)
			if err != nil {
				log.Errorf("prodApiURLs watch json.Unmarshal err: %v", err)
				continue
			}

			devApiUrlMap := make(map[string]struct{})
			urls := make([]string, 0, len(urlInfo.ApiUrls))
			for _, apiurl := range urlInfo.ApiUrls {
				url := strings.Trim(apiurl.URL, "/")
				devApiUrlMap[url] = struct{}{}
				urls = append(urls, url)
			}
			this.devApiUrlMap = devApiUrlMap
			log.Infof("service [%s] update dev api urls %v", this.Name, urls)

		case <-ticker.C:
		}
	}

	this.prodApiURLs.Close()
	this.devApiURLs.Close()
}

func (this *Service) Close() {
	this.closed = true
}

type UrlInfo struct {
	ApiUrls []ApiUrl `json:"api_urls"`
}

type ApiUrl struct {
	URL    string `json:"url"`
	Method string `json:"method"`
}
