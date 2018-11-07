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
	if _SimpleNode == nil {
		_SimpleNode = &SimpleNode{
			P2PServer:	p2p.NewP2PServer(),
			Accounts: LoadNodeAccounts(),
		}
	
		go _SimpleNode.P2PServer.Start()
	}

	return _SimpleNode
}
