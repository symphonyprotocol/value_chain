package node

import (
	"github.com/symphonyprotocol/p2p"
)

var _SimpleNode *SimpleNode

type SimpleNode struct {
	P2PServer	*p2p.P2PServer
	Accounts	*NodeAccounts
}

func GetSimpleNode() *SimpleNode {
	if _SimpleNode == nil {
		_SimpleNode = InitSimpleNode()
	}

	return _SimpleNode
}

func InitSimpleNode() *SimpleNode {
	n := &SimpleNode{
		P2PServer:	p2p.NewP2PServer(),
		Accounts: LoadNodeAccounts(),
	}

	// n.P2PServer.Start()

	return n
}
