package controllers

import (
	"git.oschina.net/kuaishangtong/common/utils/log"
	"git.oschina.net/kuaishangtong/navi/gateway/api"
	"git.oschina.net/kuaishangtong/navi/gateway/constants"
	"git.oschina.net/kuaishangtong/navi/gateway/httpproxy"
	"git.oschina.net/kuaishangtong/navi/gateway/ipfilter"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"net"
	"net/http"
	"strings"
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

	remoteIp := strings.Split(this.Ctx.Request.RemoteAddr, ":")[0]
	rip := net.ParseIP(remoteIp)
	version := constants.PROD_VERSION
	servercounts := 0

	isdeny, isdev := ipfilter.IpFilter(service, rip)

	if isdeny {
		this.Ctx.ResponseWriter.WriteHeader(403)
		return
	}

	api := api.GlobalApiManager.GetApi(service)
	if api != nil {

		if isdev {
			// 开发版本
			servercounts = api.Cluster.DevServerCount()
			version = constants.DEV_VERSION
			if _, ok := api.DevServerUrlMap[apiurl]; !ok {
				respstr := "{\"responseCode\":404,\"responseJSON\":\"\"}"
				this.Ctx.ResponseWriter.Write([]byte(respstr))
				return
			}
		} else {
			// 生产版本
			servercounts = api.Cluster.ProdServerCount()
			if _, ok := api.ProdServerUrlMap[apiurl]; !ok {
				respstr := "{\"responseCode\":404,\"responseJSON\":\"\"}"
				this.Ctx.ResponseWriter.Write([]byte(respstr))
				return
			}
		}

		var err error
		var firstCall bool = true
		var host string

		for retries := 0; (err != nil || firstCall) && retries < servercounts; retries++ {

			firstCall = false
			director := func(req *http.Request) *http.Request {
				req = this.Ctx.Request
				req.URL.Scheme = "http"
				host = api.Cluster.Select(service+"/"+apiurl, req.Method, host, version)
				req.URL.Host = host
				req.URL.Path = "/" + apiurl
				req.Header.Set("RemoteAddr", this.Ctx.Request.RemoteAddr)
				log.Infof("remote IP %s, proxy prod service %s api /%s to host %s", remoteIp, service, apiurl, host)
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
