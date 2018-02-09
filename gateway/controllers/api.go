package controllers

import (
	"git.oschina.net/kuaishangtong/common/utils/log"
	"git.oschina.net/kuaishangtong/navi/gateway/api"
	"git.oschina.net/kuaishangtong/navi/gateway/httpproxy"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"net/http"
)

type ApiController struct {
	beego.Controller
}

func (this *ApiController) Init(ct *context.Context, controllerName, actionName string, app interface{}) {
	this.Controller.Init(ct, controllerName, actionName, app)
}

func (this *ApiController) Proxy() {
	service := this.Ctx.Input.Param(":service")
	apiurl := this.Ctx.Input.Param(":api")

	api := api.GlobalApiManager.GetApi(service)
	if api != nil {
		if _, ok := api.ServerURLs[apiurl]; ok {

			var err error
			var firstCall bool = true
			var host string

			for err != nil || firstCall {
				firstCall = false
				director := func(req *http.Request) *http.Request {
					req = this.Ctx.Request
					req.URL.Scheme = "http"
					host = api.Cluster.Select(service+"/"+apiurl, req.Method, host)
					req.URL.Host = host
					req.URL.Path = "/" + apiurl
					req.Header.Set("RemoteAddr", this.Ctx.Request.RemoteAddr)
					log.Infof("remote %s, proxy service %s api /%s to host %s", this.Ctx.Request.RemoteAddr, service, apiurl, host)
					return req
				}
				proxy := &httpproxy.ReverseProxy{Director: director}
				err = proxy.ServeHTTP(this.Ctx.ResponseWriter, this.Ctx.Request)
			}

			return
		}
	}

	respstr := "{\"responseCode\":404,\"responseJSON\":\"\"}"
	this.Ctx.ResponseWriter.Write([]byte(respstr))
}
