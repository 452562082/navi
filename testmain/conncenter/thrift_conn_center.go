package conncenter

import (
	"git.apache.org/thrift.git/lib/go/thrift"
	"kuaishangtong/common/utils/log"
	"kuaishangtong/navi/testmain/navi_thrift"
	"time"
)

type ThriftConn struct {
	scpool   *ServerConnPool
	host     string
	interval int
	*navi_thrift.NaviServiceClient
	closed    bool
	available bool

	reConnFlag chan struct{}
}

func newThriftConn(pool *ServerConnPool) (*ThriftConn, error) {
	var err error
	conn := &ThriftConn{
		scpool:     pool,
		host:       pool.host,
		interval:   pool.interval,
		closed:     false,
		available:  true,
		reConnFlag: make(chan struct{}),
	}

	conn.NaviServiceClient, err = conn.connect()
	if err != nil {
		return nil, err
	}

	go conn.check()
	return conn, nil
}

func (tc *ThriftConn) check() {

	for !tc.closed {
		select {
		case <-tc.reConnFlag:
			tc.close()

			newClient, err := tc.connect()
			if err != nil {
				log.Errorf("reconnect to %s err: %v", tc.host, err)

				reconnectOK := false
				for !reconnectOK && !tc.closed {
					ticker := time.NewTicker(1 * time.Second)

					select {
					case <-ticker.C:
						newClient, err := tc.connect()
						if err != nil {
							log.Errorf("reconnect to %s err: %v", tc.host, err)
						} else {
							tc.NaviServiceClient = newClient
							tc.available = true
							reconnectOK = true
						}
					}
					ticker.Stop()
				}
			} else {
				tc.NaviServiceClient = newClient
				tc.available = true
			}
		}
	}
}

func (tc *ThriftConn) GetServerConnPool() *ServerConnPool {
	return tc.scpool
}

func (tc *ThriftConn) connect() (*navi_thrift.NaviServiceClient, error) {
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()

	transport, err := thrift.NewTSocketTimeout(tc.host, time.Duration(gTimeout)*time.Second)
	if err != nil {
		return nil, err
	}

	useTransport, err := transportFactory.GetTransport(transport)
	if err != nil {
		transport.Close()
		return nil, err
	}

	client := navi_thrift.NewNaviServiceClientFactory(useTransport, protocolFactory)
	if err := transport.Open(); err != nil {
		transport.Close()
		useTransport.Close()
		return nil, err
	}

	return client, nil
}

func (tc *ThriftConn) HostName() string {
	return tc.host
}

func (tc *ThriftConn) Available() bool {
	return tc.available
}

// 当外部调用rpc接口报错err，一般为rpc连接出问题，需要重新连接
func (tc *ThriftConn) Reconnect() {
	tc.available = false
	tc.reConnFlag <- struct{}{}
}

// 关闭Conn下的thrift rpc连接
func (tc *ThriftConn) close() {
	tc.available = false
	if tc.NaviServiceClient != nil {
		tc.NaviServiceClient.InputProtocol.Transport().Close()
		tc.NaviServiceClient.OutputProtocol.Transport().Close()
		tc.NaviServiceClient.Transport.Close()
	}
}

// 关闭Conn连接
func (tc *ThriftConn) Close() {
	tc.closed = true
	tc.close()
	//close(tc.reConnFlag)
}
