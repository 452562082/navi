package main

import (
	"git.oschina.net/kuaishangtong/common/utils/log"
	"git.oschina.net/kuaishangtong/navi/gateway/api"
	"git.oschina.net/kuaishangtong/navi/gateway/ipfilter"
	_ "git.oschina.net/kuaishangtong/navi/gateway/routers"
	"github.com/astaxie/beego"
	_ "net/http/pprof"
)

func main() {
	log.SetLevel(beego.AppConfig.DefaultInt("logLevel", 6))
	err := api.Init()
	if err != nil {
		log.Fatal(err)
	}

	err = ipfilter.InitFilter()
	if err != nil {
		log.Fatal(err)
	}

	beego.Run()
}
