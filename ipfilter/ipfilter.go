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

var __ipFilterManager *ipFilterManager

/*
 *	zkhosts: 			zookeeper 连接地址
 *	zkIpFilterPath:		ip过滤保存json的zookeeper节点 默认 /navi/ipfilter
 */
func InitFilter(zkHosts []string, zkIpFilterPath string) (err error) {
	__ipFilterManager, err = newIpfilter(zkHosts, zkIpFilterPath)
	return
}

func IpFilter(serviceName string, ip net.IP) (isdeny bool, isdev bool) {

	__ipFilterManager.rwlock.RLock()
	defer __ipFilterManager.rwlock.RUnlock()

	if deny_nets, ok := __ipFilterManager.denyNetMap[serviceName]; ok {
		if len(deny_nets) != 0 {
			for _, on := range deny_nets {
				if on.Contains(ip) {
					return true, false
				}
			}
		}
	}

	if dev_nets, ok := __ipFilterManager.devNetMap[serviceName]; ok {
		if len(dev_nets) != 0 {
			for _, on := range dev_nets {
				if on.Contains(ip) {
					return false, true
				}
			}
		}
	}

	return false, false
}

type ipfilter struct {
	Service string   `json:"service"`
	DevIps  []string `json:"dev_ips"`
	DenyIps []string `json:"deny_ips"`
}

type ipfilters struct {
	Ipfilters []*ipfilter `json:"ipfilters"`
}

type ipFilterManager struct {
	zkIpFilterPath string

	devNetMap   map[string][]*net.IPNet
	denyNetMap  map[string][]*net.IPNet
	rwlock      *sync.RWMutex
	filterStore store.Store
}

func newIpfilter(zkhosts []string, zkIpFilterPath string) (*ipFilterManager, error) {
	store, err := libkv.NewStore(store.ZK, zkhosts, nil)
	if err != nil {
		return nil, err
	}

	_ipfilter := &ipFilterManager{
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

	go _ipfilter.watch()

	return _ipfilter, nil
}

func (p *ipFilterManager) init() error {
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

func (p *ipFilterManager) watch() {
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

func (p *ipFilterManager) addDevNets(serviceName string, nets []string) error {
	if len(nets) == 0 {
		return fmt.Errorf("dev nets is empty")
	}

	var dev_nets []*net.IPNet

	dev_nets = make([]*net.IPNet, 0, len(nets))

	for _, ip := range nets {
		if len(ip) == 0 {
			continue
		}
		_, ipnet, err := net.ParseCIDR(ip)
		if err != nil {
			return fmt.Errorf("ParseCIDR %s err: %v", ip, err)
		}
		log.Info("add dev nets filter", ipnet)
		dev_nets = append(dev_nets, ipnet)
	}

	p.rwlock.Lock()
	p.devNetMap[serviceName] = dev_nets
	p.rwlock.Unlock()

	return nil
}

func (p *ipFilterManager) addDenyNets(serviceName string, nets []string) error {
	if len(nets) == 0 {
		return fmt.Errorf("deny nets is empty")
	}

	deny_nets := make([]*net.IPNet, 0, len(nets))

	for _, ip := range nets {
		if len(ip) == 0 {
			continue
		}
		_, ipnet, err := net.ParseCIDR(ip)
		if err != nil {
			return fmt.Errorf("ParseCIDR %s err: %v", ip, err)
		}
		log.Info("add deny nets filter", ipnet)
		deny_nets = append(deny_nets, ipnet)
	}

	p.rwlock.Lock()
	p.denyNetMap[serviceName] = deny_nets
	p.rwlock.Unlock()
	return nil
}
