package network

import "net/rpc"

type RPCClient interface {
	Call(serviceName string, args any, reply any) error
}

type RPCDialer interface {
	Dial(address string) (RPCClient, error)
}

type realRPCDialer struct{}

func RealRPCDialer() *realRPCDialer {
	return &realRPCDialer{}
}

func (d *realRPCDialer) Dial(address string) (RPCClient, error) {
	return rpc.DialHTTP("tcp", address)
}
