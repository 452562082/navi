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
			director := func(req *http.Request) *http.Request {
				//var newreq *http.Request
				//newreq = req.WithContext(req.Context())
				req = this.Ctx.Request
				req.URL.Scheme = "http"
				host := api.Cluster.Select(service+"/"+apiurl, req.Method)
				req.URL.Host = host
				req.URL.Path = "/" + apiurl
				log.Debugf("proxy service %s api /%s to host %s", service, apiurl, host)
				//log.Debug("2  -->", req)
				//log.Debug("2  URL -->", req.URL.Scheme)
				//newreq.URL.Scheme = "http"
				//host := api.Cluster.Select(service+"/"+apiurl, req.Method)
				//newreq.URL.Host = host
				//newreq.URL.Path = "/" + apiurl
				//log.Debug("URL -->", newreq.URL)
				return req
			}
			proxy := &httpproxy.ReverseProxy{Director: director}
			proxy.ServeHTTP(this.Ctx.ResponseWriter, this.Ctx.Request)
		}
	}
}
