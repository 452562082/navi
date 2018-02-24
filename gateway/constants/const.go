package constants

import (
	"github.com/astaxie/beego"
	"strings"
)

var ZookeeperHosts []string
var URLServicePath string
var HTTPServicePath string
var IPFilterPath string

const PROD_MODE = "prod"
const DEV_MODE = "dev"

func init() {
	ZookeeperHosts = beego.AppConfig.Strings("zookeeper.hosts")
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
