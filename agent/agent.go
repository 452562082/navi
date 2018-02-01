package agent

import (
	"fmt"
	"git.oschina.net/kuaishangtong/common/utils/log"
	"time"
)

type Agent struct {
	Plugins    PluginContainer
	agenter    Agenter
	servername string
	address    string
	typ        string
}

// NewServer returns a server.
func NewAgent(servername, address string, typ string, options ...OptionFn) (*Agent, error) {
	var err error

	a := &Agent{
		Plugins:    &pluginContainer{},
		servername: servername,
		address:    address,
		typ:        typ,
	}

	//if options!=nil{
	//
	//}
	for _, op := range options {
		op(a)
	}

	switch typ {
	case "rpc":
		a.agenter, err = a.NewThriftAgenter()
		if err != nil {
			return nil, err
		}

	case "http":
		a.agenter, err = a.NewHttpAgenter()
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unknown server type")
	}

	return a, nil

}

// Serve starts and listens RPC requests.
func (a *Agent) Serve() (err error) {

	//serviceName, err := a.agenter.ServiceName()
	//if err != nil {
	//	return err
	//}

	_, err = a.agenter.Ping()
	if err != nil {
		return err
	}

	err = a.RegisterName(a.servername, nil, a.servername)
	if err != nil {
		return err
	}

	log.Infof("register service %s successful", a.servername)

	var service_active bool = true

	pingTicker := time.NewTicker(time.Second)
	defer pingTicker.Stop()

	serviceTicker := time.NewTicker(time.Minute)
	defer serviceTicker.Stop()

	for {

		if a.agenter != nil {
			a.agenter.Close()
		}

		switch a.typ {
		case "rpc":
			a.agenter, err = a.NewThriftAgenter()
			if err != nil {
				log.Error(err)
			}

		case "http":
			a.agenter, err = a.NewHttpAgenter()
			if err != nil {
				log.Error(err)
			}

		default:
			log.Error(err)
		}

		//a.agenter, err = a.NewThrifter()
		//if err != nil {
		//	log.Error(err)
		//}

		select {
		case <-pingTicker.C:
			if a.agenter != nil {

				_, err = a.agenter.Ping()
				if err != nil {
					if service_active {
						service_active = false
						err = a.UnRegisterName(a.servername)
						if err != nil {
							log.Error(err)
							continue
						}
						log.Debugf("unregister service %s successful", a.servername)
					}
					continue
				}

				if !service_active {
					err = a.RegisterName(a.servername, nil, a.servername)
					if err != nil {
						log.Error(err)
						continue
					} else {
						service_active = true
						log.Debugf("register service %s successful", a.servername)
					}
				}
			}
		case <-serviceTicker.C:
		}
	}

	return nil
}
