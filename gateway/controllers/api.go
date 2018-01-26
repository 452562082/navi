package controllers

import (
	"git.oschina.net/kuaishangtong/common/utils/log"
	"git.oschina.net/kuaishangtong/navi/gateway/api"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"net/http"
	"net/http/httputil"
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
			director := func(req *http.Request) {
				req = this.Ctx.Request
				req.URL.Scheme = "http"
				host := api.Cluster.Select(service+"/"+apiurl, req.Method)
				log.Debugf("service %s api %s, host %s", service, api, host)
				req.URL.Host = host
			}
			proxy := &httputil.ReverseProxy{Director: director}
			proxy.ServeHTTP(this.Ctx.ResponseWriter, this.Ctx.Request)
		}
	}
}
