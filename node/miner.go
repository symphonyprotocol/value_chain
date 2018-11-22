package node

import (
	"github.com/symphonyprotocol/scb/block"
	"github.com/symphonyprotocol/simple-node/node/diagram"
	"time"
	"github.com/symphonyprotocol/log"
)

var mLogger = log.GetLogger("miner")

type NodeMiner struct {
	isMining	bool
	isIdle		bool	// when isMining is true and there's no pending tx for packaging
	runningPow	*block.ProofOfWork
	stopSign	chan struct{}
}

func NewNodeMiner() *NodeMiner {
	return &NodeMiner{
		stopSign: make(chan struct{}),
	}
}

func (n *NodeMiner) IsMining() bool { return n.isMining }

func (n *NodeMiner) StartMining() {
	if n.isMining {
		mLogger.Warn("The miner is already working")
		return
	}

	n.isMining = true
	go func() {
		for {
			select {
			case <- n.stopSign:
				if n.runningPow != nil {
					n.runningPow.Stop()
					n.runningPow = nil
				}
			default:
				time.Sleep(time.Millisecond)
				sNode := GetSimpleNode()
				currentAccount := sNode.Accounts.CurrentAccount.ToWIFCompressed()
				pendingTxs := make([]*block.Transaction, 0, 0)
				txs := GetSimpleNode().Chain.chain.FindAllUnpackTransaction()
				for _, v := range txs {
					pendingTxs = append(pendingTxs, v...)
				}

				if len(pendingTxs) > 0 && n.runningPow == nil {
					n.runningPow = block.Mine(currentAccount, func(txs []*block.Transaction) {
						// broadcast
						myLastBlock := sNode.Chain.GetMyLastBlock()
						ctx := sNode.P2PServer.GetP2PContext()
						if myLastBlock != nil {
							
							bsLogger.Trace("Broadcasting sync message")
							ctx.BroadcastToNearbyNodes(diagram.NewBlockSyncDiagram(ctx, &myLastBlock.Header), 20, nil)
						} else {
							ctx.BroadcastToNearbyNodes(diagram.NewBlockSyncDiagram(ctx, &block.BlockHeader{ Height: -1 }), 20, nil)
						}
					})
				}
			}
		}
	}()
}

func (n *NodeMiner) StopMining() {
	if !n.isMining {
		mLogger.Warn("The miner already stopped")
		return
	}

	n.isMining = false
	n.stopSign <- struct{}{}
}

func (n *NodeMiner) IsIdle() bool {
	return n.runningPow == nil
}
