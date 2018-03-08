package conncenter

type GrpcConn struct {
	scpool   *ServerConnPool
	host     string
	interval int
	client
	closed    bool
	available bool

	reConnFlag chan struct{}
}
