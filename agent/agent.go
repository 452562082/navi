package agent

import (
	"git.oschina.net/kuaishangtong/common/utils/log"
	"time"
)

type Agent struct {
	Plugins PluginContainer
	t       *thrifter
	address string
}

// NewServer returns a server.
func NewAgent(address string, options ...OptionFn) (*Agent, error) {
	var err error

	a := &Agent{
		Plugins: &pluginContainer{},
		address: address,
	}

	for _, op := range options {
		op(a)
	}

	a.t, err = a.NewThrifter()
	if err != nil {
		return nil, err
	}

	return a, nil

}

// Serve starts and listens RPC requests.
func (a *Agent) Serve() (err error) {

	serviceName, err := a.t.ServiceName()
	if err != nil {
		return err
	}

	_, err = a.t.Ping()
	if err != nil {
		return err
	}

	err = a.RegisterName(serviceName, nil, serviceName)
	if err != nil {
		return err
	}

	log.Debugf("register service %s successful", serviceName)

	var service_active bool = true

	pingTicker := time.NewTicker(time.Second)
	defer pingTicker.Stop()

	serviceTicker := time.NewTicker(time.Minute)
	defer serviceTicker.Stop()

	for {

		if a.t != nil {
			a.t.Close()
		}

		a.t, err = a.NewThrifter()
		if err != nil {
			log.Error(err)
		}

		select {
		case <-pingTicker.C:
			if a.t != nil {

				_, err = a.t.Ping()
				if err != nil {
					if service_active {
						service_active = false
						err = a.UnRegisterName(serviceName)
						if err != nil {
							log.Error(err)
							continue
						}
						log.Debugf("unregister service %s successful", serviceName)
					}
					continue
				}

				if !service_active {
					err = a.RegisterName(serviceName, nil, serviceName)
					if err != nil {
						log.Error(err)
						continue
					} else {
						service_active = true
						log.Debugf("register service %s successful", serviceName)
					}
				}
			}
		case <-serviceTicker.C:
		}
	}

	return nil
}

func (a *Agent) ServiceType() (string, error) {
	return a.t.ServiceType()
}
