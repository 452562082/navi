/*
 * Copyright © 2017 Xiao Zhang <zzxx513@gmail.com>.
 * Use of this source code is governed by an MIT-style
 * license that can be found in the LICENSE file.
 */
package navicli

import (
	"git.apache.org/thrift.git/lib/go/thrift"
	"git.oschina.net/kuaishangtong/common/utils/log"
)

type thriftClient struct {
	thriftService interface{}
	transport     thrift.TTransport
	factory       thrift.TProtocolFactory
}

func (t *thriftClient) init(addr string, clientCreator func(trans thrift.TTransport, f thrift.TProtocolFactory) interface{}) {
	if t.thriftService != nil {
		return
	}
	log.Debugf("connecting thrift addr: %s", addr)
	t.connect(addr)
	t.thriftService = clientCreator(t.transport, t.factory)
}

func (t *thriftClient) connect(hostPort string) {
	tSocket, err := thrift.NewTSocket(hostPort)
	logPanicIf(err)

	t.transport, err = thrift.NewTTransportFactory().GetTransport(tSocket)
	logPanicIf(err)

	err = t.transport.Open()
	logPanicIf(err)

	t.factory = thrift.NewTBinaryProtocolFactoryDefault()
}

func (t *thriftClient) close() error {
	if t.transport == nil {
		return nil
	}
	return t.transport.Close()
}
