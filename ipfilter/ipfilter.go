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

var ip_filter *ipFilter

/*
 *	zkhosts: 			zookeeper 连接地址
 *	zkIpFilterPath:		ip过滤保存json的zookeeper节点 默认 /navi/ipfilter
 */
func InitIpFilter(zkHosts []string, zkIpFilterPath string) (err error) {
	ip_filter, err = newIpFilter(zkHosts, zkIpFilterPath)
	return
}

func IpFilter(serviceName string, ip net.IP) (isdeny bool, isdev bool) {

	ip_filter.rwlock.RLock()
	defer ip_filter.rwlock.RUnlock()

	if deny_nets, ok := ip_filter.denyNetMap[serviceName]; ok {
		if len(deny_nets) != 0 {
			for _, on := range deny_nets {
				if on.Contains(ip) {
					return true, false
				}
			}
		}
	}

	if dev_nets, ok := ip_filter.devNetMap[serviceName]; ok {
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

type ipFilterRule struct {
	ServiceName string   `json:"service_name"`
	DevIps      []string `json:"dev_ips"`
	DenyIps     []string `json:"deny_ips"`
}

type ipFilterRules struct {
	IpFilterRules []*ipFilterRule `json:"ip_filter_rules"`
}

type ipFilter struct {
	zkIpFilterPath string

	devNetMap   map[string][]*net.IPNet
	denyNetMap  map[string][]*net.IPNet
	rwlock      *sync.RWMutex
	filterStore store.Store
}

func newIpFilter(zkhosts []string, zkIpFilterPath string) (*ipFilter, error) {
	store, err := libkv.NewStore(store.ZK, zkhosts, nil)
	if err != nil {
		return nil, err
	}

	_ipfilter := &ipFilter{
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

	return _ipfilter, nil
}

func (p *ipFilter) init() error {
	event, err := p.filterStore.Watch(p.zkIpFilterPath, nil)
	if err != nil {
		return err
	}

	go p.watch(event)

	return nil
}

func (p *ipFilter) watch(event <-chan *store.KVPair) {
	for {
		select {
		case ipfilterJsonStr := <-event:
			var ip_filter_rules ipFilterRules
			err := json.Unmarshal(ipfilterJsonStr.Value, &ip_filter_rules)
			if err != nil {
				log.Error(err)
				continue
			}

			for _, ifrule := range ip_filter_rules.IpFilterRules {
				err = p.addDevNets(ifrule.ServiceName, ifrule.DevIps)
				if err != nil {
					log.Error(err)
					continue
				}

				err = p.addDenyNets(ifrule.ServiceName, ifrule.DenyIps)
				if err != nil {
					log.Error(err)
					continue
				}
			}
		}

	}
}

func (p *ipFilter) addDevNets(serviceName string, nets []string) error {
	//if len(nets) == 0 {
	//	return fmt.Errorf("dev nets is empty")
	//}

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
		log.Infof("service %s add dev network %v", serviceName, ipnet)
		dev_nets = append(dev_nets, ipnet)
	}

	p.rwlock.Lock()
	p.devNetMap[serviceName] = dev_nets
	p.rwlock.Unlock()

	return nil
}

func (p *ipFilter) addDenyNets(serviceName string, nets []string) error {
	//if len(nets) == 0 {
	//	return fmt.Errorf("deny network is empty")
	//}

	deny_nets := make([]*net.IPNet, 0, len(nets))

	for _, ip := range nets {
		if len(ip) == 0 {
			continue
		}
		_, ipnet, err := net.ParseCIDR(ip)
		if err != nil {
			return fmt.Errorf("ParseCIDR %s err: %v", ip, err)
		}
		log.Infof("service %s add deny network %v", serviceName, ipnet)
		deny_nets = append(deny_nets, ipnet)
	}

	p.rwlock.Lock()
	p.denyNetMap[serviceName] = deny_nets
	p.rwlock.Unlock()
	return nil
}
