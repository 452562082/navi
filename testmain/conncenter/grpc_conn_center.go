package conncenter

import (
	"google.golang.org/grpc"
	"kuaishangtong/common/utils/log"
	"kuaishangtong/navi/testmain/navi_grpc"
	"time"
)

type GrpcConn struct {
	scpool   *ServerConnPool
	host     string
	interval int
	navi_grpc.NaviClient
	closed    bool
	available bool
	conn      *grpc.ClientConn

	reConnFlag chan struct{}
}

func newGrpcConn(pool *ServerConnPool) (*GrpcConn, error) {
	var err error
	gconn := &GrpcConn{
		scpool:     pool,
		host:       pool.host,
		interval:   pool.interval,
		closed:     false,
		available:  true,
		reConnFlag: make(chan struct{}),
	}

	gconn.NaviClient, err = gconn.connect()
	if err != nil {
		return nil, err
	}

	go gconn.check()
	return gconn, nil
}

func (gc *GrpcConn) check() {

	for !gc.closed {
		select {
		case <-gc.reConnFlag:
			gc.close()

			newClient, err := gc.connect()
			if err != nil {
				log.Errorf("reconnect to %s err: %v", gc.host, err)

				reconnectOK := false
				for !reconnectOK && !gc.closed {
					ticker := time.NewTicker(1 * time.Second)

					select {
					case <-ticker.C:
						newClient, err := gc.connect()
						if err != nil {
							log.Errorf("reconnect to %s err: %v", gc.host, err)
						} else {
							gc.NaviClient = newClient
							gc.available = true
							reconnectOK = true
						}
					}
					ticker.Stop()
				}
			} else {
				gc.NaviClient = newClient
				gc.available = true
			}
		}
	}
}

func (gc *GrpcConn) GetServerConnPool() *ServerConnPool {
	return gc.scpool
}

func (gc *GrpcConn) HostName() string {
	return gc.host
}

func (gc *GrpcConn) connect() (navi_grpc.NaviClient, error) {
	var err error
	gc.conn, err = grpc.Dial(gc.host, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return navi_grpc.NewNaviClient(gc.conn), nil
}

func (gc *GrpcConn) Available() bool {
	return gc.available
}

// 当外部调用rpc接口报错err，一般为rpc连接出问题，需要重新连接
func (gc *GrpcConn) Reconnect() {
	gc.available = false
	gc.reConnFlag <- struct{}{}
}

// 关闭Conn下的thrift rpc连接
func (gc *GrpcConn) close() {
	gc.available = false
	if gc.conn != nil {
		gc.conn.Close()
	}
}

// 关闭Conn连接
func (gc *GrpcConn) Close() {
	gc.closed = true
	gc.close()
}
