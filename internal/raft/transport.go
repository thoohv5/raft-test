package raft

import (
	"net"
	"os"
	"time"

	"github.com/hashicorp/raft"
)

func newRaftTransport(o *options) (*raft.NetworkTransport, error) {
	address, err := net.ResolveTCPAddr("tcp", o.addr)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(address.String(), address, 3, 10*time.Second, os.Stdout)
	if err != nil {
		return nil, err
	}
	return transport, nil
}
