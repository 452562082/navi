package registry

// KVPair contains a key and a string.
type KVPair struct {
	Key   string
	Value string
}

// ServiceDiscovery defines ServiceDiscovery of zookeeper, etcd and consul
type ServiceDiscovery interface {
	GetServices() []*KVPair
	WatchService() chan []*KVPair
	RemoveWatcher(ch chan []*KVPair)
	Clone(servicePath string) (ServiceDiscovery, error)
	Close()
}
