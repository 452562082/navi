package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"kuaishangtong/common/utils/log"
	"kuaishangtong/navi/gateway/httpproxy"
	"kuaishangtong/navi/gateway/service"
	"net/http"
	"time"
)

type ApiController struct {
	beego.Controller
}

func (this *ApiController) Init(ct *context.Context, controllerName, actionName string, app interface{}) {
	this.Controller.Init(ct, controllerName, actionName, app)
}

func (this *ApiController) Proxy() {
	service_name := this.Ctx.Input.Param(":service")
	api_url := this.Ctx.Input.URL()[len(service_name)+8:]
	mode := this.Ctx.Input.Header("mode")

	srv := service.GlobalServiceManager.GetService(service_name)
	if srv != nil {
		if api_exist := srv.ExistApi(api_url, mode); !api_exist {
			this.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
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
				var is_hash_select bool
				var key string
				if this.Ctx.Request.URL.Query().Get("key") != ""{
					is_hash_select = true
					key = this.Ctx.Request.URL.Query().Get("key")
				}else {
					is_hash_select = false
				}

				host = srv.Cluster.Select(service_name+"/"+api_url, req.Method, host, mode, is_hash_select, key)
				log.Infof("host:%v",host)
				req.URL.Host = host
				req.URL.Path = "/" + api_url
				req.Header.Set("RemoteAddr", this.Ctx.Request.RemoteAddr)
				req.Header.Set("service", service_name)
				//log.Infof("remote addr %s, proxy service [%s] %s api /%s to host %s",
				//	this.Ctx.Request.RemoteAddr, service_name, mode, api_url, host)

				return req
			}
			starttime := time.Now()
			proxy := &httpproxy.ReverseProxy{Director: director, Transport: &nethttp.Transport{}}
			err = proxy.ServeHTTP(this.Ctx.ResponseWriter, this.Ctx.Request)
			if err != nil {
				log.Errorf("remote addr %s, proxy service [%s] %s api /%s to host %s err: %v",
					this.Ctx.Request.RemoteAddr, service_name, mode, api_url, host, err)
			}
			timeconsumer := time.Now().Sub(starttime)
			log.Infof("remote addr %s, proxy service [%s] %s api /%s to host %s success | %s",
				this.Ctx.Request.RemoteAddr, service_name, mode, api_url, host, timeconsumer.String())
		}

		if err != nil || servercounts == 0 {
			if servercounts == 0 {
				log.Errorf("service [%s] %s server count is 0", service_name, mode)
			}
			this.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
		}
		return
	}

	this.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
	return
}
