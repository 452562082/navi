package ipfilter

import (
	"encoding/json"
	"fmt"
	"git.oschina.net/kuaishangtong/common/utils/log"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"net"
	"sync"
)

var __ipfilterctl *ipfilterController

func InitFilter(zkhosts []string, zkIpFilterPath string) (err error) {
	__ipfilterctl, err = newIpfilter(zkhosts, zkIpFilterPath)
	return
}

func IpFilter(serviceName string, ip net.IP) (isdeny bool, isdev bool) {

	__ipfilterctl.rwlock.RLock()
	defer __ipfilterctl.rwlock.RUnlock()

	if deny_nets, ok := __ipfilterctl.denyNetMap[serviceName]; ok {
		if len(deny_nets) != 0 {
			for _, on := range deny_nets {
				if on.Contains(ip) {
					return false, false
				}
			}
		}
	}

	if dev_nets, ok := __ipfilterctl.devNetMap[serviceName]; ok {
		if len(dev_nets) != 0 {
			for _, on := range dev_nets {
				if on.Contains(ip) {
					return false, true
				}
			}
		}
	}

	return true, false
}

type ipfilter struct {
	Service string   `json:"service"`
	DevIps  []string `json:"dev_ips"`
	DenyIps []string `json:"deny_ips"`
}

type ipfilters struct {
	Ipfilters []*ipfilter `json:"ipfilters"`
}

type ipfilterController struct {
	zkIpFilterPath string

	devNetMap   map[string][]*net.IPNet
	denyNetMap  map[string][]*net.IPNet
	rwlock      *sync.RWMutex
	filterStore store.Store
}

func newIpfilter(zkhosts []string, zkIpFilterPath string) (*ipfilterController, error) {
	store, err := libkv.NewStore(store.ZK, zkhosts, nil)
	if err != nil {
		return nil, err
	}

	_ipfilter := &ipfilterController{
		zkIpFilterPath: zkIpFilterPath,
		devNetMap:      make(map[string][]*net.IPNet),
		denyNetMap:     make(map[string][]*net.IPNet),
		rwlock:         &sync.RWMutex{},
		filterStore:    store,
	}

	err = _ipfilter.init()
	if err != nil {
		return nil, err
	}

	go _ipfilter.watchIpfilters()

	return _ipfilter, nil
}

func (p *ipfilterController) init() error {
	ipfilterJsonStr, err := p.filterStore.Get(p.zkIpFilterPath)
	if err != nil {
		return err
	}

	var _ipfilters ipfilters
	err = json.Unmarshal(ipfilterJsonStr.Value, &_ipfilters)
	if err != nil {
		return err
	}

	for _, _ipfilter := range _ipfilters.Ipfilters {
		err = p.addDevNets(_ipfilter.Service, _ipfilter.DevIps)
		if err != nil {
			return err
		}

		err = p.addDenyNets(_ipfilter.Service, _ipfilter.DenyIps)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *ipfilterController) watchIpfilters() {
	for {
		event, err := p.filterStore.Watch(p.zkIpFilterPath, nil)
		if err != nil {
			log.Error(err)
			continue
		}

		ipfilterJsonStr := <-event
		var _ipfilters ipfilters
		err = json.Unmarshal(ipfilterJsonStr.Value, &_ipfilters)
		if err != nil {
			log.Error(err)
			continue
		}

		for _, _ipfilter := range _ipfilters.Ipfilters {
			err = p.addDevNets(_ipfilter.Service, _ipfilter.DevIps)
			if err != nil {
				log.Error(err)
				continue
			}

			err = p.addDenyNets(_ipfilter.Service, _ipfilter.DenyIps)
			if err != nil {
				log.Error(err)
				continue
			}
		}

	}
}

func (p *ipfilterController) addDevNets(serviceName string, nets []string) error {
	if len(nets) == 0 {
		return fmt.Errorf("dev nets is empty")
	}

	var dev_nets []*net.IPNet
	//var exist bool

	//if dev_net, exist = devNetMap[serviceName]; !exist {
	dev_nets = make([]*net.IPNet, 0, len(nets))
	//}

	for _, n := range nets {
		if len(n) == 0 {
			continue
		}
		_, ipnet, err := net.ParseCIDR(n)
		if err != nil {
			return fmt.Errorf("ParseCIDR %s err: %v", n, err)
		}
		log.Info("add dev nets filter", ipnet)
		dev_nets = append(dev_nets, ipnet)
	}

	p.rwlock.Lock()
	p.devNetMap[serviceName] = dev_nets
	p.rwlock.Unlock()

	return nil
}

func (p *ipfilterController) addDenyNets(serviceName string, nets []string) error {
	deny_nets := make([]*net.IPNet, 0, len(nets))

	for _, n := range nets {
		if len(n) == 0 {
			continue
		}
		_, ipnet, err := net.ParseCIDR(n)
		if err != nil {
			return fmt.Errorf("ParseCIDR %s err: %v", n, err)
		}
		log.Info("add deny nets filter", ipnet)
		deny_nets = append(deny_nets, ipnet)
	}

	p.rwlock.Lock()
	p.denyNetMap[serviceName] = deny_nets
	p.rwlock.Unlock()
	return nil
}
