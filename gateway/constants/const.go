package constants

import (
	"github.com/astaxie/beego"
	"strings"
)

var ZookeeperHosts []string
var ServiceBasePath string
var ServiceListPath string

func init() {
	ZookeeperHosts = beego.AppConfig.Strings("zookeeper.hosts")
	ServiceBasePath = beego.AppConfig.String("zookeeper.service_base_path")
	ServiceListPath = beego.AppConfig.String("zookeeper.service_list_path")

	if len(ZookeeperHosts) == 0 {
		panic("do not configurate zookeeper.hosts")
	}

	if len(ServiceBasePath) == 0 {
		panic("do not configurate zookeeper.service_base_path")
	}

	ServiceBasePath = strings.Trim(ServiceBasePath, "/")

	if len(ZookeeperHosts) == 0 {
		panic("do not configurate zookeeper.service_list_path")
	}
	ServiceListPath = strings.Trim(ServiceListPath, "/")
}
