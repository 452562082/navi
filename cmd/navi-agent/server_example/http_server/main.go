package main

import (
	"encoding/json"
	"fmt"
	glog "git.oschina.net/kuaishangtong/common/utils/log"
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"io/ioutil"
	"net/http"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/ping", pingHandler)
	r.HandleFunc("/servicename", serviceNameHandler)
	r.HandleFunc("/servicemode", serviceModeHandler)
	r.HandleFunc("/hello", helloHandler)
	http.Handle("/", r)

	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		glog.Fatal("ListenAndServe: ", err.Error())
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	glog.Infof("req remoteAddr: %s for url /%s", r.RemoteAddr, r.URL.Path)
	fmt.Fprintf(w, "pong")
}

func serviceNameHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	glog.Infof("req remoteAddr: %s for url /%s", r.RemoteAddr, r.URL.Path)
	fmt.Fprintf(w, "MyHttpTest")
}

func serviceModeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	glog.Infof("req remoteAddr: %s for url /%s", r.RemoteAddr, r.URL.Path)
	fmt.Fprintf(w, "dev")
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		glog.Errorf("req Method is not POST")
		return
	}

	textCarrier := opentracing.HTTPHeadersCarrier(r.Header)
	wireSpanContext, err := opentracing.GlobalTracer().Extract(
		opentracing.TextMap, textCarrier)
	if err != nil {
		panic(err)
	}
	serverSpan := opentracing.GlobalTracer().StartSpan(
		"POST sayHello",
		ext.RPCServerOption(wireSpanContext))
	serverSpan.SetTag("component", "server")
	defer serverSpan.Finish()

	glog.Infof("req remoteAddr: %s for url /%s", r.RemoteAddr, r.URL.Path)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		glog.Errorf("read req body err: %v", err)
		serverSpan.LogFields(log.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var body map[string]interface{}

	err = json.Unmarshal(data, &body)
	if err != nil {
		glog.Errorf("Unmarshal req json body err: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	serverSpan.LogFields(log.String("request body", body["name"].(string)))

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "hello %s", body["name"].(string))
}
