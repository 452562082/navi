package routers

import (
	"kuaishangtong/navi/gateway/constants"
	"kuaishangtong/navi/gateway/controllers"
	"kuaishangtong/navi/ipfilter"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/plugins/cors"
	"net"
	"strings"
)

func init() {

	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin"},
		AllowCredentials: true,
	}))

	beego.InsertFilter("/kstAI/:service([\\w|\u4e00-\u9fff|\\.\\-\\:\\_]+)/*", beego.BeforeRouter, func(ctx *context.Context) {

		remoteIp := strings.Split(ctx.Request.RemoteAddr, ":")[0]
		rip := net.ParseIP(remoteIp)
		service_name := ctx.Input.Param(":service")

		isdeny, isdev := ipfilter.IpFilter(service_name, rip)

		// 请求IP不被允许访问，直接返回403
		if isdeny {
			ctx.ResponseWriter.WriteHeader(403)
			return
		}

		if isdev {
			ctx.Request.Header.Set("mode", constants.DEV_MODE)
		} else {
			ctx.Request.Header.Set("mode", constants.PROD_MODE)
		}

	})

	beego.Router("/kstAI/:service([\\w|\u4e00-\u9fff|\\.\\-\\:\\_]+)/:api([\\w|\u4e00-\u9fff|\\.\\-\\:\\_]+)",
		&controllers.ApiController{}, "get,post:Proxy")

}
