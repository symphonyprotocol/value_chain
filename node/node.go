package node

import (
	"github.com/symphonyprotocol/p2p/tcp"
	"github.com/symphonyprotocol/p2p"
	"fmt"
)

var _SimpleNode *SimpleNode

type SimpleNode struct {
	P2PServer	*p2p.P2PServer
	Accounts	*NodeAccounts
	Chain		*NodeChain
	Miner		*NodeMiner
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
			Chain: LoadNodeChain(),
			Miner: NewNodeMiner(),
		}
	
		_SimpleNode.P2PServer.Use(NewBlockSyncMiddleware())
		_SimpleNode.P2PServer.Use(&TransactionMiddleware{ BaseMiddleware: tcp.NewBaseMiddleware() })
		go _SimpleNode.P2PServer.Start()
	}

	return _SimpleNode
}

func (b *SimpleNode) DashboardData() interface{} { return [][]string{
	[]string{ "ID: ", fmt.Sprintf("%v", b.P2PServer.GetP2PContext().LocalNode().GetID()) },
	[]string{ "Is Mining: ", fmt.Sprintf("%v", GetSimpleNode().Chain.GetMyHeight()) },
} }
func (b *SimpleNode) DashboardType() string { return "table" }
func (b *SimpleNode) DashboardTitle() string { return "Simple Node" }
func (b *SimpleNode) DashboardTableHasColumnTitles() bool { return false }

