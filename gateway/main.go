/*
                      _ooOoo_
                     o8888888o
                     88" . "88
                     (| -_- |)
                     O\  =  /O
                  ____/`---'\____
                .'  \\|     |//  `.
               /  \\|||  :  |||//  \
              /  _||||| -:- |||||-  \
              |   | \\\  -  /// |   |
              | \_|  ''\---/''  |   |
              \  .-\__  `-`  ___/-. /
            ___`. .'  /--.--\  `. . __
         ."" '<  `.___\_<|>_/___.'  >'"".
        | | :  `- \`.;`\ _ /`;.`/ - ` : | |
        \  \ `-.   \_ __\ /__ _/   .-` /  /
   ======`-.____`-.___\_____/___.-`____.-'======
                      `=---='
   ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
				佛祖保佑       永无BUG
*/

package main

import (
	"git.oschina.net/kuaishangtong/common/utils/log"
	"git.oschina.net/kuaishangtong/navi/gateway/api"
	"git.oschina.net/kuaishangtong/navi/gateway/constants"
	_ "git.oschina.net/kuaishangtong/navi/gateway/routers"
	"git.oschina.net/kuaishangtong/navi/ipfilter"
	"github.com/astaxie/beego"
	_ "net/http/pprof"
)

func main() {
	log.SetLevel(beego.AppConfig.DefaultInt("logLevel", 6))
	err := api.Init()
	if err != nil {
		log.Fatal(err)
	}

	err = ipfilter.InitFilter(constants.ZookeeperHosts, constants.IPFilterPath)
	if err != nil {
		log.Fatal(err)
	}

	beego.Run()
}
