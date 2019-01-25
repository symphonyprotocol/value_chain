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
	mineSign	chan struct{}
}

func NewNodeMiner() *NodeMiner {
	miner := &NodeMiner{
		stopSign: make(chan struct{}),
		mineSign: make(chan struct{}),
	}
	go miner.mineLoop()
	go miner.mineTrigger()

	return miner
}

func (n *NodeMiner) IsMining() bool { 
	n.mtx.Lock()
	defer n.mtx.Unlock()
	return n.isMining 
}

func (n *NodeMiner) SetIsMining(v bool) {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	n.isMining = v
}

func (n *NodeMiner) StartMining() {
	if n.IsMining() {
		mLogger.Warn("The miner is already working")
		return
	}

	n.SetIsMining(true)
}

func (n *NodeMiner) mineLoop() {
	for {
		select {
		case <- n.stopSign:
			n.ClearRunningPow()
		case <- n.mineSign:
			sNode := GetValueChainNode()
			currentAccount := sNode.Accounts.CurrentAccount.ToWIFCompressed()
			pendingTxs := make([]*block.Transaction, 0, 0)
			txs := block.FindAllUnpackTransaction()
			for _, v := range txs {
				pendingTxs = append(pendingTxs, v...)
			}

			if len(pendingTxs) > 0 && n.IsRunningPowEmpty() && sNode.P2PServer != nil && sNode.P2PServer.GetP2PContext() != nil && sNode.P2PServer.GetP2PContext().LocalNode() != nil {
				doneSignal := make(chan struct{})
				n.runningPow = block.Mine(currentAccount, func(b *block.Block) {
					bsLogger.Trace("Mined !!!!!!!!!")
					// broadcast
					// myLastBlock := sNode.Chain.GetMyLastBlock()
					ctx := sNode.P2PServer.GetP2PContext()
					bsLogger.Trace("Broadcasting new block message")
					go ctx.BroadcastToNearbyNodes(diagram.NewBlockSendDiagram(ctx, b), 20, nil)
					bsLogger.Trace("Broadcasting new block message done.")
					// need lock here
					n.ClearRunningPow()
					bsLogger.Trace("Running pow cleared")
				}, func() {
					doneSignal <- struct{}{}
				})
				select {
				case <- n.runningPow.Cancelled:
					bsLogger.Debug("Running pow cancelled")
				case <- doneSignal:
					bsLogger.Debug("Running pow done")
				}
			}
		}
	}
}

func (n *NodeMiner) mineTrigger() {
	for {
		time.Sleep(time.Millisecond)
		// bsLogger.Trace("Going to mine - 0")
		if n.IsMining() {
			// bsLogger.Trace("Going to mine - 1")
			n.mineSign <- struct{}{}
		} else {
			for ;len(n.mineSign) > 0; {
				<- n.mineSign
			}
		}
	}
}

func (n *NodeMiner) StopMining() {
	if !n.IsMining() {
		mLogger.Warn("The miner already stopped")
		return
	}

	n.SetIsMining(false)
	n.stopSign <- struct{}{}
}

func (n *NodeMiner) IsIdle() bool {
	result := false
	n.mtx.Lock()
	defer n.mtx.Unlock()
	result = n.runningPow == nil
	return result
}

func (n *NodeMiner) ClearRunningPow() {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	bsLogger.Trace("Clearing running pow")
	if n.runningPow != nil {
		if !n.runningPow.IsFinished() {
			bsLogger.Trace("Stopping running pow")
			n.runningPow.Stop()
			bsLogger.Trace("Stopped running pow")
		}
		n.runningPow = nil
	}
}

func (n *NodeMiner) IsRunningPowEmpty() bool {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	return n.runningPow == nil
}


