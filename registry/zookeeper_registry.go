package registry

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/zookeeper"
	"github.com/rcrowley/go-metrics"
	"kuaishangtong/common/utils/log"
)

func init() {
	zookeeper.Register()
}

var serverset string = `{"serviceEndpoint":{"host":"%s","port":%s},"additionalEndpoints":{},"status":"ALIVE"}`

// ZooKeeperRegister implements zookeeper registry.
type ZooKeeperRegister struct {
	// service address, for example, tcp@127.0.0.1:8972, quic@127.0.0.1:1234
	ServiceAddress string

	// prometheus target host
	PrometheusTargetHost string

	// prometheus target host
	PrometheusTargetPort string

	// zookeeper addresses
	ZooKeeperServers []string
	// base path for rpcx server, for example com/example/rpcx
	BasePath string
	Mode     string
	Metrics  metrics.Registry
	// Registered services
	Services       []string
	metasLock      sync.RWMutex
	metas          map[string]string
	UpdateInterval time.Duration

	Options *store.Config
	kv      store.Store
}

// Start starts to connect zookeeper cluster
func (p *ZooKeeperRegister) Start() error {
	if p.kv == nil {
		kv, err := libkv.NewStore(store.ZK, p.ZooKeeperServers, p.Options)
		if err != nil {
			log.Errorf("cannot create zk registry: %v", err)
			return err
		}
		p.kv = kv
	}

	if p.BasePath[0] == '/' {
		p.BasePath = p.BasePath[1:]
	}

	err := p.kv.Put(p.BasePath, []byte("navi_path"), &store.WriteOptions{IsDir: true})
	if err != nil {
		log.Errorf("cannot create zk path %s: %v", p.BasePath, err)
		return err
	}

	if p.UpdateInterval > 0 {
		ticker := time.NewTicker(p.UpdateInterval)
		go func() {
			defer p.kv.Close()

			// refresh service TTL
			for range ticker.C {

				//set this same metrics for all services at this server
				for _, name := range p.Services {

					var nodePath string
					nodePath = fmt.Sprintf("%s", p.BasePath)
					if name != "" {
						nodePath += "/" + name
					}
					if p.Mode != "" {
						nodePath += "/" + p.Mode
					}
					nodePath += "/" + p.ServiceAddress

					_, err := p.kv.Get(nodePath)
					if err != nil {
						log.Errorf("can't get data of node: %s, because of %v", nodePath, err.Error())

						p.metasLock.RLock()
						metadata := p.metas[name]
						p.metasLock.RUnlock()

						if len(p.PrometheusTargetHost) > 0 && len(p.PrometheusTargetPort) > 0 {
							metadata = fmt.Sprintf(serverset, p.PrometheusTargetHost, p.PrometheusTargetPort)
						}

						err = p.kv.Put(nodePath, []byte(metadata), &store.WriteOptions{TTL: 1})
						if err != nil {
							log.Errorf("cannot re-create zookeeper path %s: %v", nodePath, err)
						}
					} else {
						var metadata string
						if len(p.PrometheusTargetHost) > 0 && len(p.PrometheusTargetPort) > 0 {
							metadata = fmt.Sprintf(serverset, p.PrometheusTargetHost, p.PrometheusTargetPort)
						}

						p.kv.Put(nodePath, []byte(metadata), &store.WriteOptions{TTL: 1})
					}
				}

			}
		}()
	}

	return nil
}

// HandleConnAccept handles connections from clients
func (p *ZooKeeperRegister) HandleConnAccept(conn net.Conn) (net.Conn, bool) {
	if p.Metrics != nil {
		clientMeter := metrics.GetOrRegisterMeter("clientMeter", p.Metrics)
		clientMeter.Mark(1)
	}
	return conn, true
}

// Register handles registering event.
// this service is registered at BASE/serviceName/thisIpAddress node
func (p *ZooKeeperRegister) Register(name string, rcvr interface{}, data string) (err error) {
	//if "" == strings.TrimSpace(name) {
	//	err = errors.New("Register service `name` can't be empty")
	//	return
	//}

	if p.kv == nil {
		zookeeper.Register()
		kv, err := libkv.NewStore(store.ZK, p.ZooKeeperServers, nil)
		if err != nil {
			log.Errorf("cannot create zk registry: %v", err)
			return err
		}
		p.kv = kv
	}

	if p.BasePath[0] == '/' {
		p.BasePath = p.BasePath[1:]
	}
	err = p.kv.Put(p.BasePath, []byte("navi_path"), &store.WriteOptions{IsDir: true})
	if err != nil {
		log.Errorf("cannot create zk path %s: %v", p.BasePath, err)
		return err
	}

	nodePath := fmt.Sprintf("%s", p.BasePath)
	if name != "" {
		nodePath += "/" + name
		err = p.kv.Put(nodePath, []byte(name), &store.WriteOptions{IsDir: true})
		if err != nil {
			log.Errorf("cannot create zk path %s: %v", nodePath, err)
			return err
		}
	}

	if p.Mode != "" {
		nodePath += "/" + p.Mode
		err = p.kv.Put(nodePath, []byte(name), &store.WriteOptions{IsDir: true})
		if err != nil {
			log.Errorf("cannot create zk path %s: %v", nodePath, err)
			return err
		}
	}

	nodePath += "/" + p.ServiceAddress

	var metadata string
	if len(p.PrometheusTargetHost) > 0 && len(p.PrometheusTargetPort) > 0 {
		metadata = fmt.Sprintf(serverset, p.PrometheusTargetHost, p.PrometheusTargetPort)
	} else {
		metadata = data
	}

	err = p.kv.Put(nodePath, []byte(metadata), &store.WriteOptions{TTL: 1})
	if err != nil {
		log.Errorf("cannot create zk path %s: %v", nodePath, err)
		return err
	}

	p.metasLock.Lock()
	if p.metas == nil {
		p.metas = make(map[string]string)
	}

	if _, ok := p.metas[name]; !ok {
		p.Services = append(p.Services, name)
	}

	p.metas[name] = metadata
	p.metasLock.Unlock()
	return
}

func (p *ZooKeeperRegister) UnRegister(name string) (err error) {
	//if "" == strings.TrimSpace(name) {
	//	err = fmt.Errorf("Register service `name` can't be empty")
	//	return
	//}

	if p.kv == nil {
		zookeeper.Register()
		kv, err := libkv.NewStore(store.ZK, p.ZooKeeperServers, nil)
		if err != nil {
			log.Errorf("cannot create zk registry: %v", err)
			return err
		}
		p.kv = kv
	}

	if p.BasePath[0] == '/' {
		p.BasePath = p.BasePath[1:]
	}

	var nodePath string
	nodePath = fmt.Sprintf("%s", p.BasePath)
	if name != "" {
		nodePath += "/" + name
	}
	if p.Mode != "" {
		nodePath += "/" + p.Mode
	}
	nodePath += "/" + p.ServiceAddress

	err = p.kv.Delete(nodePath)
	if err != nil {
		log.Errorf("cannot delete zk path %s: %v", nodePath, err)
		return err
	}

	p.metasLock.Lock()
	if p.metas != nil {
		delete(p.metas, name)
	}
	for i, v := range p.Services {
		if v == name {
			p.Services[i] = p.Services[len(p.Services)-1]
			p.Services = p.Services[:len(p.Services)-1]
		}
	}
	p.metasLock.Unlock()
	return
}
