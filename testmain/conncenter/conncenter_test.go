package conncenter

import (
	"kuaishangtong/common/utils/log"
	"kuaishangtong/navi/lb"
	"testing"
)

func TestEngine_GetConn(t *testing.T) {
	err := InitConnCenter("/navi/rpcservice", "MyTest/dev", []string{"192.168.1.16:2181"}, 3, 1, lb.Failover)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 32; i++ {

		go func() {
			conn, err := connCenter.GetConn()
			if err != nil {
				t.Fatal(err)
			}

			s, err := conn.(*Conn).Ping()
			if err != nil {
				t.Fatal(err)
			}

			log.Infof(s)
			err = PutConn(conn)
			if err != nil {
				t.Fatal(err)
			}
		}()

	}

	select {}
}
