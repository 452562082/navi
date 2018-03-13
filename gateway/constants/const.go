package constants

import (
	"github.com/astaxie/beego"
	"kuaishangtong/common/env"
	"strings"
)

var ZookeeperHosts []string
var URLServicePath string  // 存储service的提供的 api 接口,即 url
var HTTPServicePath string // 存储service的服务IP
var IPFilterPath string

const PROD_MODE = "prod"
const DEV_MODE = "dev"

func init() {

	if zkHosts, err := env.GetZookeeperHosts(); err == nil {
		ZookeeperHosts = zkHosts
	} else {
		ZookeeperHosts = beego.AppConfig.Strings("zookeeper.hosts")
	}

	URLServicePath = beego.AppConfig.String("zookeeper.url_service_path")
	HTTPServicePath = beego.AppConfig.String("zookeeper.http_service_path")
	IPFilterPath = beego.AppConfig.String("zookeeper.ip_filter_path")

	if len(ZookeeperHosts) == 0 {
		panic("do not configurate zookeeper.hosts")
	}

	if len(URLServicePath) == 0 {
		panic("do not configurate zookeeper.service_base_path")
	}

	URLServicePath = strings.Trim(URLServicePath, "/")

	if len(ZookeeperHosts) == 0 {
		panic("do not configurate zookeeper.service_list_path")
	}
	HTTPServicePath = strings.Trim(HTTPServicePath, "/")

	if len(IPFilterPath) == 0 {
		panic("do not configurate zookeeper.ip_filter_path")
	}
	IPFilterPath = strings.Trim(IPFilterPath, "/")

}
