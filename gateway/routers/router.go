package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/plugins/cors"
	"kuaishangtong/navi/gateway/constants"
	"kuaishangtong/navi/gateway/controllers"
	"kuaishangtong/navi/ipfilter"
	"net"
	"strings"
	"net/http"
	"fmt"
	"encoding/json"
	"io/ioutil"
)

func init() {

	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin"},
		AllowCredentials: true,
	}))

	//beego.InsertFilter("*", beego.BeforeRouter, func(ctx *context.Context) {
	//	access_token := ctx.Request.Header.Get("access_token")
	//	client_id, scope, err := GetSessionClient(access_token)
	//	if err != nil {
	//		//TODO
	//		//token无效
	//	}
	//
	//	//TODO
	//	//scope权限相关
	//})

	filterFunc := func(ctx *context.Context) {
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
	}

	pattern := "/kstAI/:service([\\w|\u4e00-\u9fff|\\.\\-\\:\\_]+)"

	for index := 1; index <= 6; index++ {
		pattern += "/*"
		beego.InsertFilter(pattern, beego.BeforeRouter, filterFunc)
		beego.Router(pattern, &controllers.ApiController{}, "*:Proxy")
	}

	beego.Get("/", func(ctx *context.Context) {
		ctx.Output.Body([]byte("hello world"))
	})
}

func GetSessionClient(token string) (string, string, error) {
	res, err := http.Get(fmt.Sprintf("http://%s/restricted?token=%s",beego.AppConfig.String("oauth_host"),token))
	defer res.Body.Close()
	if err != nil {
		return "","",err
	}
	client_data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "","",err
	}

	var client_json map[string]interface{}
	err = json.Unmarshal(client_data,&client_json)
	if err != nil {
		return "","",err
	}

	client_id := client_json["ClientID"].(string)
	scope := client_json["Scope"].(string)
	return client_id, scope, nil
}