
package agent

import (
	"fmt"
	"kuaishangtong/common/utils/httplib"
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
	return httplib.Get(url).String()
}

func (ha *httpAgenter) ServiceName() (string, error) {
	url := fmt.Sprintf("http://%s/servicename", ha.Address)
	return httplib.Get(url).String()
}

func (ha *httpAgenter) ServiceMode() (string, error) {
	url := fmt.Sprintf("http://%s/servicemode", ha.Address)
	return httplib.Get(url).String()
}

func (ha *httpAgenter) Close() error {
	return nil
}
