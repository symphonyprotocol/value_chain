package node

import (
	"github.com/symphonyprotocol/p2p"
)

type SimpleNode struct {
	p2pServer	*p2p.P2PServer
}

func InitSimpleNode() *SimpleNode {
	n := &SimpleNode{
		p2pServer:	p2p.NewP2PServer(),
	}

	return n
}
