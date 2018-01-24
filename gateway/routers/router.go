package routers

import (
	"git.oschina.net/kuaishangtong/navi/gateway/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/plugins/cors"
)

func init() {

	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin"},
		AllowCredentials: true,
	}))

	// project
	beego.Router("/kst/:service([\\w|\u4e00-\u9fff|\\.\\-\\:\\_]+)/:api([\\w|\u4e00-\u9fff|\\.\\-\\:\\_]+)",
		&controllers.ApiController{}, "get,post:Proxy")

}
