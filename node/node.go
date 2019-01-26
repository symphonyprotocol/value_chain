package node

import (
	"time"
	"github.com/symphonyprotocol/p2p/tcp"
	"github.com/symphonyprotocol/p2p"
	"fmt"
)

var _ValueChainNode *ValueChainNode

type ValueChainNode struct {
	P2PServer	*p2p.P2PServer
	Accounts	*NodeAccounts
	Chain		*NodeChain
	Miner		*NodeMiner
	startTime	time.Time
	IsSyncing	bool 
}

func GetValueChainNode() *ValueChainNode {
	if _ValueChainNode == nil {
		_ValueChainNode = InitValueChainNode()
	}

	return _ValueChainNode
}

func InitValueChainNode() *ValueChainNode {
	if _ValueChainNode == nil {
		_ValueChainNode = &ValueChainNode{
			P2PServer:	p2p.NewP2PServer(),
			Accounts: LoadNodeAccounts(),
			Chain: LoadNodeChain(),
			Miner: NewNodeMiner(),
			startTime: time.Now(),
		}
	
		_ValueChainNode.P2PServer.Use(NewBlockSyncMiddleware())
		_ValueChainNode.P2PServer.Use(&TransactionMiddleware{ BaseMiddleware: tcp.NewBaseMiddleware() })
		go _ValueChainNode.P2PServer.Start()
	}

	return _ValueChainNode
}

func (b *ValueChainNode) DashboardData() interface{} { return [][]string{
	[]string{ "ID: ", fmt.Sprintf("%v", b.P2PServer.GetP2PContext().LocalNode().GetID()) },
	[]string{ "Is Mining: ", fmt.Sprintf("%v", GetValueChainNode().Miner.IsMining()) },
	[]string{ "PubKey:", b.P2PServer.GetP2PContext().LocalNode().GetPublicKey()},
	[]string{ "Local Address:", fmt.Sprintf("%v:%v", b.P2PServer.GetP2PContext().LocalNode().GetLocalIP().String(), b.P2PServer.GetP2PContext().LocalNode().GetLocalPort())},
	[]string{ "Remote Address:", fmt.Sprintf("%v:%v", b.P2PServer.GetP2PContext().LocalNode().GetRemoteIP().String(), b.P2PServer.GetP2PContext().LocalNode().GetRemotePort())},
	[]string{ "Up time:", fmt.Sprintf("%v", time.Since(b.startTime))},
	[]string{ "Is Syncing:", fmt.Sprintf("%v", b.IsSyncing)},
} }
func (b *ValueChainNode) DashboardType() string { return "table" }
func (b *ValueChainNode) DashboardTitle() string { return "Simple Node" }
func (b *ValueChainNode) DashboardTableHasColumnTitles() bool { return false }

