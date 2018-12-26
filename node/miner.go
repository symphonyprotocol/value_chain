package node

import (
	"sync"
	"github.com/symphonyprotocol/scb/block"
	"github.com/symphonyprotocol/value_chain/node/diagram"
	"time"
	"github.com/symphonyprotocol/log"
)

var mLogger = log.GetLogger("miner")

type NodeMiner struct {
	isMining	bool
	isIdle		bool	// when isMining is true and there's no pending tx for packaging

	runningPow	*block.ProofOfWork

	mtx			sync.RWMutex
	stopSign	chan struct{}
}

func NewNodeMiner() *NodeMiner {
	return &NodeMiner{
		stopSign: make(chan struct{}),
	}
}

func (n *NodeMiner) IsMining() bool { 
	n.mtx.Lock()
	defer n.mtx.Unlock()
	return n.isMining 
}

func (n *NodeMiner) StartMining() {
	n.mtx.Lock()
	if n.isMining {
		mLogger.Warn("The miner is already working")
		return
	}

	n.isMining = true
	n.mtx.Unlock()
	go func() {
BREAK_LOOP:
		for {
			select {
			case <- n.stopSign:
				n.mtx.Lock()
				if n.runningPow != nil {
					n.runningPow.Stop()
					n.runningPow = nil
				}
				n.mtx.Unlock()
				break BREAK_LOOP
			default:
				time.Sleep(time.Millisecond)
				sNode := GetValueChainNode()
				currentAccount := sNode.Accounts.CurrentAccount.ToWIFCompressed()
				pendingTxs := make([]*block.Transaction, 0, 0)
				txs := GetValueChainNode().Chain.chain.FindAllUnpackTransaction()
				for _, v := range txs {
					pendingTxs = append(pendingTxs, v...)
				}

				n.mtx.Lock()
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
						// need lock here
						n.mtx.Lock()
						n.runningPow = nil
						n.mtx.Unlock()
					})
				}
				n.mtx.Unlock()
			}
		}
	}()
}

func (n *NodeMiner) StopMining() {
	n.mtx.Lock()
	if !n.isMining {
		mLogger.Warn("The miner already stopped")
		return
	}

	n.isMining = false
	n.mtx.Unlock()
	n.stopSign <- struct{}{}
}

func (n *NodeMiner) IsIdle() bool {
	result := false
	n.mtx.Lock()
	defer n.mtx.Unlock()
	result = n.runningPow == nil
	return result
}
