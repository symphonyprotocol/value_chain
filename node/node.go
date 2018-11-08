package node

import (
	"github.com/symphonyprotocol/p2p/tcp"
	"github.com/symphonyprotocol/p2p"
	"github.com/symphonyprotocol/simple-node/node/middleware"
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
	
		_SimpleNode.P2PServer.Use(&middleware.BlockSyncMiddleware{ BaseMiddleware: &tcp.BaseMiddleware{} })
		_SimpleNode.P2PServer.Use(&middleware.TransactionMiddleware{ BaseMiddleware: &tcp.BaseMiddleware{} })
		go _SimpleNode.P2PServer.Start()
	}

	return _SimpleNode
}
