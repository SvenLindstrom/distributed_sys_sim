package worker

import "net/rpc"

type RPCClient interface {
	Call(serviceName string, args any, reply any) error
}

type RPCDialer interface {
	Dial(address string) (RPCClient, error)
}

type RealRPCDialer struct{}

func (d *RealRPCDialer) Dial(address string) (RPCClient, error) {
	return rpc.DialHTTP("tcp", address)
}
