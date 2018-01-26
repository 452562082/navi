package controllers

import (
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

	//transport := http.DefaultTransport
	//
	//// step 1
	//outReq := new(http.Request)
	//*outReq = *this.Ctx.Request // this only does shallow copies of maps
	//
	//if clientIP, _, err := net.SplitHostPort(this.Ctx.Request.RemoteAddr); err == nil {
	//	if prior, ok := outReq.Header["X-Forwarded-For"]; ok {
	//		clientIP = strings.Join(prior, ", ") + ", " + clientIP
	//	}
	//	outReq.Header.Set("X-Forwarded-For", clientIP)
	//}
	//
	//// step 2
	//res, err := transport.RoundTrip(outReq)
	//if err != nil {
	//	this.Ctx.ResponseWriter.WriteHeader(http.StatusBadGateway)
	//	return
	//}
	//
	//// step 3
	//for key, value := range res.Header {
	//	for _, v := range value {
	//		this.Ctx.ResponseWriter.Header().Add(key, v)
	//	}
	//}
	//
	//this.Ctx.ResponseWriter.WriteHeader(res.StatusCode)
	//io.Copy(this.Ctx.ResponseWriter, res.Body)
	//res.Body.Close()

	api := api.GlobalApiManager.GetApi(service)
	if api != nil {
		if _, ok := api.ServerURLs[apiurl]; ok {
			director := func(req *http.Request) *http.Request {
				var newreq *http.Request
				newreq = req.WithContext(req.Context())
				//req = this.Ctx.Request
				//req.URL.Scheme = "http"
				//host := api.Cluster.Select(service+"/"+apiurl, req.Method)
				//req.URL.Host = host
				//log.Debugf("service %s api %s, host %s", service, apiurl, host)
				//log.Debug("2  -->", req)
				//log.Debug("2  URL -->", req.URL.Scheme)
				newreq.URL.Scheme = "http"
				host := api.Cluster.Select(service+"/"+apiurl, req.Method)
				newreq.URL.Host = host
				return newreq
			}
			proxy := &httpproxy.ReverseProxy{Director: director}
			proxy.ServeHTTP(this.Ctx.ResponseWriter, this.Ctx.Request)
		}
	}
}
