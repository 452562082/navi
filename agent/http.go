package agent

import (
	"fmt"
	"kuaishangtong/common/utils/httplib"
	"kuaishangtong/common/utils/log"
	"time"
)

var defaultSetting = httplib.HTTPSettings{
	UserAgent:        "httpAgenter",
	ConnectTimeout:   10 * time.Second,
	ReadWriteTimeout: 10 * time.Second,
	Gzip:             true,
	DumpBody:         true,
}

func init() {
	httplib.SetDefaultSetting(defaultSetting)
}

func (a *Agent) NewHttpAgenter() (t *httpAgenter, err error) {

	return &httpAgenter{
		Address: a.address,
	}, nil
}

type httpAgenter struct {
	Address string
}

func (ha *httpAgenter) Ping() (string, error) {
	url := fmt.Sprintf("http://%s/ping", ha.Address)
	log.Debugf("Ping url: %s", url)
	return httplib.Get(url).String()
}

func (ha *httpAgenter) ServiceName() (string, error) {
	url := fmt.Sprintf("http://%s/servicename", ha.Address)
	log.Debugf("ServiceName url: %s", url)
	return httplib.Get(url).String()
}

func (ha *httpAgenter) ServiceMode() (string, error) {
	url := fmt.Sprintf("http://%s/servicemode", ha.Address)
	log.Debugf("ServiceMode url: %s", url)
	return httplib.Get(url).String()
}

func (ha *httpAgenter) Close() error {
	return nil
}
