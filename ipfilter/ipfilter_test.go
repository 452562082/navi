package ipfilter

import (
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/zookeeper"
	"testing"
)

var zkhosts []string = []string{"192.168.1.16:2181"}

const _ipFilterRules = `{
  "ip_filter_rules": [
    {
      "service_name": "MyTest",
      "dev_ips": [
        "192.168.1.1/24",
		"127.0.0.1/32"
      ],
      "deny_ips": [
      ]
    },
    {
      "service_name": "test1",
      "dev_ips": [
      ],
      "deny_ips": [
      ]
    }
  ]
}

`

func TestAddIpFilterRule(t *testing.T) {

	zookeeper.Register()

	ipfStore, err := libkv.NewStore(store.ZK, zkhosts, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = ipfStore.Put("navi/ipfilter", []byte(_ipFilterRules), nil)
	if err != nil {
		t.Fatal(err)
	}
}
