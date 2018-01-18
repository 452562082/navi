package engine

import "testing"

func TestEngine_GetConn(t *testing.T) {
	err := InitEngine("", "", []string{"192.168.1.16:2181"}, 1, 15)
	if err != nil {
		t.Fatal(err)
	}

	conn, err := XEngine.GetConn()
	if err != nil {
		t.Fatal(err)
	}

	conn.ServiceName()
}
