/*
                      _ooOoo_
                     o8888888o
                     88" . "88
                     (| -_- |)
                     O\  =  /O
                  ____/`---'\____
                .'  \\|     |//  `.
               /  \\|||  :  |||//  \
              /  _||||| -:- |||||-  \
              |   | \\\  -  /// |   |
              | \_|  ''\---/''  |   |
              \  .-\__  `-`  ___/-. /
            ___`. .'  /--.--\  `. . __
         ."" '<  `.___\_<|>_/___.'  >'"".
        | | :  `- \`.;`\ _ /`;.`/ - ` : | |
        \  \ `-.   \_ __\ /__ _/   .-` /  /
   ======`-.____`-.___\_____/___.-`____.-'======
                      `=---='
   ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
				佛祖保佑       永无BUG
*/

package main

import (
	"fmt"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	metrics "github.com/rcrowley/go-metrics"
	"io/ioutil"
	"kuaishangtong/common/utils/daemon"
	"kuaishangtong/common/utils/log"
	"kuaishangtong/navi/agent"
	"kuaishangtong/navi/registry"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/prometheus/node_exporter/collector"
	"net/http"
)

func init() {
	prometheus.MustRegister(version.NewCollector("node_exporter"))
}

func handler(w http.ResponseWriter, r *http.Request) {
	filters := r.URL.Query()["collect[]"]

	nc, err := collector.NewNodeCollector(filters...)
	if err != nil {
		log.Warn("Couldn't create", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Couldn't create %s", err)))
		return
	}

	registry := prometheus.NewRegistry()
	err = registry.Register(nc)
	if err != nil {
		log.Error("Couldn't register collector:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Couldn't register collector: %s", err)))
		return
	}

	gatherers := prometheus.Gatherers{
		prometheus.DefaultGatherer,
		registry,
	}
	// Delegate http serving to Prometheus client library, which will call collector.Collect.
	h := promhttp.HandlerFor(gatherers,
		promhttp.HandlerOpts{
			ErrorLog:      &log.ZeusLogger{},
			ErrorHandling: promhttp.ContinueOnError,
		})

	h.ServeHTTP(w, r)
}

func main() {
	if !initializeFlags() {
		return
	}

	err := initializeConfig(*_flags.Config)
	if err != nil {
		log.Fatal(err)
	}

	if *_flags.Daemon {
		daemon.SetWorkerLogPath(defaultConfig.Log.File)
		daemon.SetLogPath(defaultConfig.Log.File + ".monitor")
		daemon.Exec(daemon.Daemon | daemon.Monitor)
	}

	// log 设置
	logSet := defaultConfig.Log
	log.SetLogFuncCall(logSet.ShowLines)
	log.SetColor(logSet.Coloured)
	log.SetLevel(logSet.Level)
	if *_flags.Daemon || logSet.Enable || defaultConfig.Server.IsDocker {
		log.SetLogFile(
			logSet.File,
			logSet.Level,
			logSet.Daily,
			logSet.Coloured,
			logSet.ShowLines,
			logSet.MaxDays)
	}

	serverhosts := strings.Split(defaultConfig.Server.ServerHosts, ";")

	serverCount := len(serverhosts)

	var agents []*agent.Agent = make([]*agent.Agent, serverCount, serverCount)

	//zkServers, err := env.GetZookeeperHosts()
	//if err != nil {
	//	zkServers = strings.Split(defaultConfig.Zookeeper.ZookeeperHosts, ";")
	//}
	zkServers := strings.Split(defaultConfig.Zookeeper.ZookeeperHosts, ";")

	for i := 0; i < serverCount; i++ {

		agents[i], err = agent.NewAgent(defaultConfig.Server.ServerName, serverhosts[i],
			defaultConfig.Server.ServerType,
			defaultConfig.Server.IsDocker,
			restart_server_in_docker)

		if err != nil {
			log.Fatal(err)
		}

		go func(index int) {
			var basePath string

			if defaultConfig.Server.ServerType == "rpc" {
				basePath = defaultConfig.Zookeeper.ZookeeperRPCServicePath
			} else if defaultConfig.Server.ServerType == "http" {
				basePath = defaultConfig.Zookeeper.ZookeeperHTTPServicePath
			}

			r := &registry.ZooKeeperRegister{
				ServiceAddress:       serverhosts[index],
				ZooKeeperServers:     zkServers,
				BasePath:             basePath,
				Metrics:              metrics.NewRegistry(),
				UpdateInterval:       2 * time.Second,
				PrometheusTargetHost: defaultConfig.PrometheusTarget.Host,
				PrometheusTargetPort: defaultConfig.PrometheusTarget.Port,
			}

			err = r.Start()
			if err != nil {
				log.Fatal(err)
			}

			agents[index].Plugins.Add(r)

			err = agents[index].Serve()
			if err != nil {
				log.Fatal(err)
			}
		}(i)
	}

	// 为 http 服务添加url注册信息
	if defaultConfig.Server.ServerType == "http" {
		data, err := ioutil.ReadFile(defaultConfig.Server.ServerHttpApiJsonFile)
		if err != nil {
			log.Fatal(err)
		}

		urlRegistry, err := libkv.NewStore(store.ZK, zkServers, nil)
		if err != nil {
			log.Fatal(err)
		}

		key := fmt.Sprintf("%s/%s/%s", strings.Trim(defaultConfig.Zookeeper.ZookeeperURLServicePath, "/"),
			defaultConfig.Server.ServerName, defaultConfig.Server.ServerMode)
		err = urlRegistry.Put(key, data, nil)
		if err != nil {
			log.Fatal(err)
		}

		urlRegistry.Close()
	}

	// This instance is only used to check collector creation and logging.
	nc, err := collector.NewNodeCollector()
	if err != nil {
		log.Fatalf("Couldn't create collector: %s", err)
	}
	log.Infof("Enabled collectors:")
	for n := range nc.Collectors {
		log.Infof(" Collector - %s", n)
	}

	http.HandleFunc("/metrics", handler)

	log.Info("Listening on", 9100)
	err = http.ListenAndServe(":9100", nil)
	if err != nil {
		log.Fatal(err)
	}

}
