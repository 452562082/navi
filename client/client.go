package client

import (
	"context"
	"time"
	"fmt"
	"sync"
	"github.com/smallnest/rpcx/share"
)

// Breaker is a CircuitBreaker interface
type Breaker interface {
	Call(func() error, time.Duration) error
}

type seqKey struct{}

// RPCClient is interface that defines one client to call one server
type RPCClient interface {
	Connect(network, address string) error
	Go(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}, done chan *Call) *Call
	Call(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) error
	Close() error

	IsClosing() bool
	IsShutdown() bool
}

// Client represents a RPC client
type Client struct {
	option	Option

	mutex		sync.Mutex // protects following
	pending		map[uint64]*Call
	closing 	bool // user has called Close
	shutdown 	bool // server has told us to stop
}

// Option contains all options for creating clients
type Option struct {
	// Retries retries to send
	Retries int
	//RPCPath for http connection
	RPCPath	string

	// Breaker is used to config CircuitBreaker
	Breaker Breaker

	Heartbeat 			bool
	HeartbeatInterval	time.Duration
}

// Call represents an active RPC
type Call struct {
	ServicePath   string            // The name of the service and method to call.
	ServiceMethod string            // The name of the service and method to call.
	Metadata      map[string]string //metadata
	ResMetadata   map[string]string
	Args          interface{} // The argument to the function (*struct).
	Reply         interface{} // The reply from the function (*struct).
	Error         error       // After completion, the error status.
	Done          chan *Call  // Strobes when call is complete.
	Raw           bool        // raw message or not
}

func (call *Call) done() {
	select {
	case call.Done <- call:
		// ok
	default:
		fmt.Printf("rpc: discarding Call reply due to insufficient Done chan capacity")
	}
}

// IsClosing client is closing or not
func (client *Client) IsClosing() bool {
	return client.closing
}

// IsShutdown client is shutdown or not
func (client *Client) IsShutdown() bool {
	return client.shutdown
}

// Go invokes the function asynchronously. It returns the Call structure representing
// the invocation. The done channel will signal when the call is complete by returning
// the same Call object. If done is nil, Go will allocate a new channel.
// If non-nil, done must be buffered or Go will deliberately crash.
func (client *Client) Go(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}, done chan *Call) *Call {
	call := new(Call)
	call.ServicePath = servicePath
	call.ServiceMethod = serviceMethod
	meta := ctx.Value(share.ReqMetaDataKey)
	if meta != nil { //copy meta in context to meta in requests
		call.Metadata = meta.(map[string]string)
	}
	call.Args = args
	call.Reply = reply
	if done == nil {
		done = make(chan *Call, 10) // buffered.
	} else {
		// If caller passes done != nil, it must arrange that
		// done has enough buffer for the number of simultaneous
		// RPCs that will be using that channel. If the channel
		// is totally unbuffered, it's best not to run at all.
		if cap(done) == 0 {
			log.Panic("rpc: done channel is unbuffered")
		}
	}
	call.Done = done
	client.send(ctx, call)
	return call
}

// Call invokes the named function, waits for it to complete, and returns its error status.
func (client *Client) Call(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) error {
	if client.option.Breaker != nil {
		return client.option.Breaker.Call(func() error {
			return client.call(ctx, servicePath, serviceMethod, args, reply)
		}, 0)
	}

	return client.call(ctx, servicePath, serviceMethod, args, reply)
}

func (client *Client) call(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) error {
	seq := new(uint64)
	ctx = context.WithValue(ctx, seqKey{}, seq)
	Done := client.Go(ctx, servicePath, serviceMethod, args, reply, make(chan *Call, 1)).Done

	var err error
	select {
	case <-ctx.Done(): //cancel by context
		client.mutex.Lock()
		call := client.pending[*seq]
		delete(client.pending, *seq)
		client.mutex.Unlock()
		if call != nil {
			call.Error = ctx.Err()
			call.done()
		}

		return ctx.Err()
	case call := <-Done:
		err = call.Error
		meta := ctx.Value(share.ResMetaDataKey)
		if meta != nil && len(call.ResMetadata) > 0 {
			resMeta := meta.(map[string]string)
			for k, v := range call.ResMetadata {
				resMeta[k] = v
			}
		}
	}

	return err
}

//func (client *Client) heartbeat() {
//	t := time.NewTicker(client.option.HeartbeatInterval)
//
//	for range t.C {
//		if client.shutdown || client.closing {
//			return
//		}
//
//		err := client.Call(context.Background(), "", "", nil, nil)
//		if err != nil {
//			fmt.Printf("failed to heartbeat to %s", client.Conn.RemoteAddr().String())
//		}
//	}
//}