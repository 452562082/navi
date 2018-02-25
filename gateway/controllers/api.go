package controllers

import (
	"git.oschina.net/kuaishangtong/common/utils/log"
	"git.oschina.net/kuaishangtong/navi/gateway/httpproxy"
	"git.oschina.net/kuaishangtong/navi/gateway/service"
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
	service_name := this.Ctx.Input.Param(":service")
	api_url := this.Ctx.Input.Param(":api")
	mode := this.Ctx.Input.Header("mode")

	srv := service.GlobalServiceManager.GetService(service_name)
	if srv != nil {

		if api_exist := srv.ExistApi(api_url, mode); !api_exist {
			respstr := "{\"responseCode\":404,\"responseJSON\":\"\"}"
			this.Ctx.ResponseWriter.Write([]byte(respstr))
			return
		}

		servercounts := srv.GetServerCount(mode)

		var err error
		var firstCall bool = true
		var host string

		for retries := 0; (err != nil || firstCall) && retries < servercounts; retries++ {

			firstCall = false
			director := func(req *http.Request) *http.Request {
				req = this.Ctx.Request
				req.URL.Scheme = "http"
				// 由 mode 来决定请求时转发到prod的集群上或dev的集群上
				host = srv.Cluster.Select(service_name+"/"+api_url, req.Method, host, mode)
				req.URL.Host = host
				req.URL.Path = "/" + api_url
				req.Header.Set("RemoteAddr", this.Ctx.Request.RemoteAddr)
				log.Infof("remote addr %s, proxy %s service %s api /%s to host %s",
					this.Ctx.Request.RemoteAddr, mode, service_name, api_url, host)
				return req
			}
			proxy := &httpproxy.ReverseProxy{Director: director}
			err = proxy.ServeHTTP(this.Ctx.ResponseWriter, this.Ctx.Request)
		}

		if err != nil {
			this.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		}
		return
	}

	respstr := "{\"responseCode\":404,\"responseJSON\":\"\"}"
	this.Ctx.ResponseWriter.Write([]byte(respstr))
	return
}
