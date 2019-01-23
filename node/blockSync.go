package node

import (
	"fmt"
	"github.com/symphonyprotocol/p2p/node"
	"math/rand"
	"sync"
	"github.com/symphonyprotocol/log"
	"github.com/symphonyprotocol/scb/block"
	"time"
	"github.com/symphonyprotocol/p2p/tcp"
	"github.com/symphonyprotocol/value_chain/node/diagram"
	"github.com/symphonyprotocol/sutil/ds"
	"github.com/symphonyprotocol/sutil/utils"
	"bytes"
	"context"
)

var bsLogger = log.GetLogger("blockSyncMiddleware").SetLevel(log.TRACE)

var (
	MAX_HEADER_PENDING	=	128
	MAX_BLOCK_PENDING	=	128
	BLOCK_REQUEST_TIMEOUT		=	30 * time.Second
	HEADER_REQUEST_TIMEOUT		=	30 * time.Second
	NEED_NEGOTIATION_THRESHOLD	=	10

	SYNC_STATE_IDLE		=	0
	SYNC_STATE_SYNCING	=	1
	SYNC_STATE_NEGOTIATION	=	2
)

type BlockSyncMiddleware struct {
	*tcp.BaseMiddleware
	syncLoopExitSignal chan struct{}
	acceptBlockChannel chan *block.Block
	acceptHeaderChannel chan *block.BlockHeader
	downloadBlockChannel chan *block.BlockHeader
	downloadBlockHeaderChannel chan int64
	downloadBlockQueue				*ds.SequentialParallelTaskQueue
	downloadBlockHeaderQueue		*ds.SequentialParallelTaskQueue
	downloadBlockPendingMap			sync.Map
	downloadHeaderPendingMap		sync.Map
	knownMaxHeight					int64
	myLastHeader					*block.BlockHeader
	lastRequestedHeight				int64
	selectedPeer					*node.RemoteNode
	previousPeer					*node.RemoteNode
	peersHeights					sync.Map				// TODO: should use ip+port as key not only ip
	askForHeaderChan				chan int64
	negotiateCounter				int
	currentState					int
	mtx								sync.RWMutex
	negotiationHeaderChan			chan *block.BlockHeader
	peerBlackList					[]*node.RemoteNode
}

func NewBlockSyncMiddleware() *BlockSyncMiddleware {
	return &BlockSyncMiddleware{
		BaseMiddleware: tcp.NewBaseMiddleware(),
		syncLoopExitSignal: make(chan struct{}),
		acceptBlockChannel: make(chan *block.Block),
		acceptHeaderChannel: make(chan *block.BlockHeader),
		downloadBlockChannel: make(chan *block.BlockHeader, 512),
		downloadBlockHeaderChannel: make(chan int64, 512),
		askForHeaderChan: make(chan int64),
		lastRequestedHeight: -1,
		negotiateCounter: 0,
		negotiationHeaderChan: make(chan *block.BlockHeader),
		peerBlackList: make([]*node.RemoteNode, 0, 0),
	}
}

func (t *BlockSyncMiddleware) Start(ctx *tcp.P2PContext) {
	t.SetSyncState(SYNC_STATE_IDLE)
	t.BaseMiddleware.Start(ctx)
	t.downloadBlockHeaderQueue = ds.NewSequentialParallelTaskQueue(MAX_HEADER_PENDING, func (tasks []*ds.ParallelTask) {
		// we got headers, need to download blocks.
		bsLogger.Info("We got finished headers, going to download content of them.")
		myLastBlock := GetValueChainNode().Chain.GetMyLastBlock()
		if myLastBlock != nil && t.myLastHeader == nil {
			t.myLastHeader = &myLastBlock.Header
		}
		for _, task := range tasks {
			if b, ok := task.Result.(*block.BlockHeader); ok {
				bsLogger.Info("Going to save Header: %v: %v", b.HashString(), b.Height)
				t.acceptHeaderChannel <- b
			}
		}
	}, func (tasks []*ds.ParallelTask) {
		// timedout
		bsLogger.Warn("Block downloading tasks (count: %v) timed out, will retry", len(tasks))
		for _, t := range tasks {
			t.Retry()
		}
	})
	t.downloadBlockQueue = ds.NewSequentialParallelTaskQueue(MAX_BLOCK_PENDING, func (tasks []*ds.ParallelTask) {
		// we got blocks, need to save them.
		// defer func() {
		// 	if err := recover(); err != nil {
		// 		bsLogger.Error("%v", err)
		// 	}
		// }()

		bsLogger.Info("We got finished blocks (count: %v), going to save them.", len(tasks))
		for _, task := range tasks {
			if b, ok := task.Result.(*block.Block); ok {
				bsLogger.Info("Going to save Block: %v: %v", b.Header.HashString(), b.Header.Height)
				t.acceptBlockChannel <- b
			}
		}
	}, func (tasks []*ds.ParallelTask) {
		// timedout
		bsLogger.Warn("Header downloading tasks (count: %v) timed out, will retry", len(tasks))
		for _, t := range tasks {
			t.Retry()
		}
	})
	t.downloadBlockHeaderQueue.Execute()
	t.downloadBlockQueue.Execute()
	t.regHandlers()
	go t.syncLoop(ctx)
	go t.downloadLoop(ctx)
	go t.acceptBlockLoop(ctx)
	go t.acceptHeaderLoop(ctx)
}

func (t *BlockSyncMiddleware) regHandlers() {
	t.HandleRequest(diagram.BLOCK_SYNC, func (ctx *tcp.P2PContext) {
		// 1. check the block height with mine,
		//   if higher than me, send BLOCK_SYNC to it
		//   else, send headers by BLOCK_SYNC_RES
		
		var syncDiag diagram.BlockSyncDiagram
		err := ctx.GetDiagram(&syncDiag)
		if err == nil && t.GetSyncState() != SYNC_STATE_NEGOTIATION {
			myHeight := GetValueChainNode().Chain.GetMyHeight()
			myLastBlock := GetValueChainNode().Chain.GetMyLastBlock()
			targetHeight := syncDiag.LastBlockHeader.Height
			bsLogger.Debug("Got BLOCK_SYNC diag: %v, myHeight: %v", syncDiag.LastBlockHeader.Height, myHeight)
			if targetHeight > myHeight {
				GetValueChainNode().IsSyncing = true
				t.SetSyncState(SYNC_STATE_SYNCING)
				t.peersHeights.Store(ctx.Params().GetTCPRemoteAddr().IP.String(), targetHeight)
			}

			if targetHeight < myHeight && myLastBlock != nil {
				// Just answer the BLOCK_SYNC message
				ctx.Send(diagram.NewBlockSyncDiagram(ctx, &myLastBlock.Header))
			}

			if targetHeight <= myHeight && 
			myLastBlock != nil && 
			t.selectedPeer != nil && 
			ctx.Params().GetTCPRemoteAddr().IP.String() == t.selectedPeer.GetRemoteIP().String() {
				GetValueChainNode().IsSyncing = false
				t.SetSyncState(SYNC_STATE_IDLE)
			}

			if targetHeight < myHeight && myLastBlock == nil {
				// ???????
				bsLogger.Error("??????? lower height than me, but I have no blocks???????")
			}
		}
	})

	t.HandleRequest(diagram.BLOCK_HEADER, func (ctx *tcp.P2PContext) {
		var headerDiag diagram.BlockHeaderDiagram
		err := ctx.GetDiagram(&headerDiag)
		if err == nil { 
			bsLogger.Trace("Going to provide header: %v", headerDiag.BlockHeight)

			b := GetValueChainNode().Chain.GetBlockByHeight(headerDiag.BlockHeight)
			
			if b != nil {
				ctx.Send(diagram.NewBlockHeaderResDiagram(ctx, b.Header.Height, &b.Header))
				bsLogger.Trace("Provided header: %v", b.Header.HashString())
			} else {
				bsLogger.Error("did not find requested header by height: %v", headerDiag.BlockHeight)
			}
		}
	})

	t.HandleRequest(diagram.BLOCK_HEADER_RES, func (ctx *tcp.P2PContext) {
		// 1. store the headers and ask for content with BLOCK_REQ
		var syncDiag diagram.BlockHeaderResDiagram
		err := ctx.GetDiagram(&syncDiag)
		if err == nil {
			bsLogger.Trace("Recieved headers")
			if t.GetSyncState() == SYNC_STATE_SYNCING {
				if _cb, _ok := t.downloadHeaderPendingMap.Load(syncDiag.BlockHeight); _ok {
					bsLogger.Debug("Header Task is in pending map")
					if cb, ok := _cb.(func(res interface{})); ok {
						bsLogger.Debug("Recieved Header %v and going to ask for block when headers are in a line!", syncDiag.BlockHeader.Height)
						cb(syncDiag.BlockHeader)
						// remove timeout map
						// t.downloadHeaderPendingTimeoutMap.Delete(syncDiag.BlockHeight)
					}
				}
			} else if t.GetSyncState() == SYNC_STATE_NEGOTIATION {
				t.negotiationHeaderChan <- syncDiag.BlockHeader
			} else {
				bsLogger.Fatal("????? what state am I %v ?????", t.GetSyncState())
			}
		}
	})

	t.HandleRequest(diagram.BLOCK_REQ, func (ctx *tcp.P2PContext) {
		// 1. If I have the block, send the block with BLOCK_REQ_RES
		var bReqDiag diagram.BlockReqDiagram
		err := ctx.GetDiagram(&bReqDiag)
		if err == nil {
			bsLogger.Trace("Going to provide block: %v", bReqDiag.BlockHeader.HashString())
			myHeight := GetValueChainNode().Chain.GetMyHeight()
			targetHeight := bReqDiag.BlockHeader.Height
			if myHeight < targetHeight {
				bsLogger.Warn("BOOM, my: %v, target: %v", myHeight, targetHeight)
				// boom
				ctx.Send(diagram.NewBlockReqResDiagram(ctx, &block.Block{ Header: block.BlockHeader{ Hash: bReqDiag.BlockHeader.Hash, Height: bReqDiag.BlockHeader.Height, Signature: nil } }))
			} else {
				bsLogger.Trace("Providing blocks")
				the_block := GetValueChainNode().Chain.GetBlock(bReqDiag.BlockHeader.Hash)
				resDiag := diagram.NewBlockReqResDiagram(ctx, the_block)
				bsLogger.Debug("block:%v, txs inside: %v", the_block.Header.Height, len(the_block.Transactions))
				ctx.Send(resDiag)
			}
		}
	})

	t.HandleRequest(diagram.BLOCK_REQ_RES, func (ctx *tcp.P2PContext) {
		// 1. verify and store the block
		var bReqResDiag diagram.BlockReqResDiagram
		err := ctx.GetDiagram(&bReqResDiag)
		if err == nil {
			bsLogger.Trace("Recieved Block !")
			if bReqResDiag.Block.Transactions == nil && bReqResDiag.Block.Header.Signature == nil {
				bsLogger.Warn("The target has no such a block, need to retry !!!!!")
				
				peers := ctx.NodeProvider().GetNearbyNodes(20)
				peersLength := len(peers)
				var randPeer *node.RemoteNode
				if peersLength > 0 {
					randPeer = peers[rand.Intn(peersLength)]
				}
				bsLogger.Trace("Retrying to ask from %v !!!!!", randPeer.GetRemoteIP())
				
				ctx.SendToPeer(diagram.NewBlockReqDiagram(ctx, &bReqResDiag.Block.Header), randPeer)
				// t.downloadBlockPendingTimeoutMap.Store(bReqResDiag.Block.Header.HashString(), time.Now())
			} else {
				// bsLogger.Info("callback map: %v", t.downloadBlockPendingMap)
				if _cb, _ok := t.downloadBlockPendingMap.Load(bReqResDiag.Block.Header.HashString()); _ok {
					bsLogger.Debug("Task is in pending map")
					if cb, ok := _cb.(func(res interface{})); ok {
						bsLogger.Debug("Recieved Block and going to store !")
						bsLogger.Debug("The block: %v, txs inside: %v", bReqResDiag.Block.Header.Height, len(bReqResDiag.Block.Transactions))
						cb(bReqResDiag.Block)
						// remove timeout map
						// t.downloadBlockPendingTimeoutMap.Delete(bReqResDiag.Block.Header.HashString())
					}
				}
			}
		}
	})

	t.HandleRequest(diagram.BLOCK_SEND, func (ctx *tcp.P2PContext) {
		// add to pending pool.
		var bSendDiag diagram.BlockSendDiagram
		err := ctx.GetDiagram(&bSendDiag)
		if err == nil && t.GetSyncState() == SYNC_STATE_IDLE {
			if GetValueChainNode().Miner.IsMining() {
				bsLogger.Debug("Cancelled mining")
				GetValueChainNode().Miner.StopMining()
			}
			GetValueChainNode().Chain.SavePendingBlock(bSendDiag.Block)
			GetValueChainNode().Miner.StartMining()
		}

		if err == nil {
			go ctx.BroadcastToNearbyNodes(bSendDiag, 20, func(_p *node.RemoteNode) bool {
				// will not broadcast back to where the msg is from.
				return _p.GetID() != bSendDiag.GetNodeID()
			})
		}
	})
}

// broadcast BLOCK_SYNC periodly
func (t *BlockSyncMiddleware) syncLoop(ctx *tcp.P2PContext) {
	exit:
	for {
		time.Sleep(10 * time.Second)
		select {
		case <- t.syncLoopExitSignal:
			bsLogger.Trace("Syncing stopped")
			break exit
		default:
			bsLogger.Trace("Going to broadcast sync message")
			myLastBlock := GetValueChainNode().Chain.GetMyLastBlock()
			var myHeight int64 = -1
			if myLastBlock != nil {
				myHeight = myLastBlock.Header.Height
				bsLogger.Trace("Broadcasting sync message")
				ctx.BroadcastToNearbyNodes(diagram.NewBlockSyncDiagram(ctx, &myLastBlock.Header), 20, nil)
			} else {
				ctx.BroadcastToNearbyNodes(diagram.NewBlockSyncDiagram(ctx, &block.BlockHeader{ Height: -1 }), 20, nil)
			}

			bestPeer := t.GetBestPeer(ctx)
			if bestPeer != nil {
				if t.selectedPeer == nil || t.selectedPeer.GetID() != bestPeer.GetID() {
					t.previousPeer = t.selectedPeer
					t.selectedPeer = bestPeer
				}
				if _h, ok := t.peersHeights.Load(bestPeer.GetRemoteIP().String()); ok {
					if h, ok := _h.(int64); ok {				
						expectedHeight := utils.Min(myHeight + int64(MAX_HEADER_PENDING), h)
						bsLogger.Trace("Asking for headers, myHeight: %v, target Height: %v, expectedHeght: %v, pendingCount: %v, lastRequestedHeight: %v", 
						myHeight, 
						h, 
						expectedHeight, 
						ds.GetSyncMapSize(&t.downloadHeaderPendingMap),
						t.lastRequestedHeight)
						for i := utils.Max(myHeight, t.lastRequestedHeight) + 1; i <= expectedHeight; i++ {
							if _, ok := t.downloadHeaderPendingMap.Load(i); !ok {
								t.downloadBlockHeaderChannel <- i
								t.lastRequestedHeight = i
							}
						}
					}
				}
			}
		}
	}
}

// check timeout map periodly
func (t *BlockSyncMiddleware) downloadLoop(ctx *tcp.P2PContext) {
	for {

		select {
		case header := <- t.downloadBlockChannel:
			bsLogger.Trace("added task to download block: %v", header.HashString())
			t.downloadBlockQueue.AddTask(&ds.ParallelTask{
				Body: func(params []interface{}, cb func(res interface{})) {
					bsLogger.Trace("Going to ask for block!: %v with height: %v", header.HashString(), header.Height)
					// t.downloadBlockPendingTimeoutMap.Store(header.HashString(), time.Now())
					bsLogger.Debug("Storing to downloadBlockPendingMap with key: %v", header.HashString())
					t.downloadBlockPendingMap.Store(header.HashString(), cb)
					if params != nil && len(params) > 0 {
						if peer, ok := params[0].(*node.RemoteNode); ok {
							bsLogger.Trace("Asking for block!: %v", header.HashString())
							ctx.SendToPeer(diagram.NewBlockReqDiagram(ctx, header), peer)
						}
					}
				},
				Params: []interface{}{
					t.selectedPeer,
				},
				Timeout: BLOCK_REQUEST_TIMEOUT,
			})
		case height := <- t.downloadBlockHeaderChannel:
			bsLogger.Trace("added task to download header: %v", height)
			t.downloadBlockHeaderQueue.AddTask(&ds.ParallelTask{
				Body: func(params []interface{}, cb func(res interface{})) {
					bsLogger.Trace("Going to ask for header!: %v", height)
					// t.downloadHeaderPendingTimeoutMap.Store(height, time.Now())
					bsLogger.Debug("Storing to downloadBlockHeaderPendingMap with key: %v", height)
					t.downloadHeaderPendingMap.Store(height, cb)
					if params != nil && len(params) > 0 {
						if peer, ok := params[0].(*node.RemoteNode); ok {
							bsLogger.Trace("Asking for header!: %v", height)
							ctx.SendToPeer(diagram.NewBlockHeaderDiagram(ctx, height), peer)
						}
					}
				},
				Params: []interface{}{
					t.selectedPeer,
				},
				Timeout: HEADER_REQUEST_TIMEOUT,
			})
		}
	}
}

func (t *BlockSyncMiddleware) acceptBlockLoop(ctx *tcp.P2PContext) {
	for {
		select {		
			case b := <- t.acceptBlockChannel:
			bsLogger.Info("Got block to be saved from acceptBlockChannel: %v (%v)", b.Header.HashString(), b.Header.Height)
			GetValueChainNode().Chain.SaveBlock(b)
			bsLogger.Info("Block saved: %v", b.Header.HashString())
			// restart mining if mining is true
			if GetValueChainNode().Miner.IsMining() && !GetValueChainNode().Miner.IsIdle() {
				bsLogger.Info("Cancelled mining !")
				GetValueChainNode().Miner.StopMining()
				GetValueChainNode().Miner.StartMining()
			}
			t.downloadBlockPendingMap.Delete(b.Header.HashString())
		}
	}
}

func (t *BlockSyncMiddleware) acceptHeaderLoop(ctx *tcp.P2PContext) {
	for {
		select {		
			case b := <- t.acceptHeaderChannel:
			bsLogger.Info("Got header to be processed from acceptHeaderChannel: %v (%v)", b.HashString(), b.Height)
			// check if these headers can be connected
			if t.myLastHeader != nil {
				bsLogger.Trace("Comparing my last header (%v): %v with recieved header's(%v) previous hash: %v",
				t.myLastHeader.Height, 
					utils.BytesToString(t.myLastHeader.Hash), 
					b.Height,
					utils.BytesToString(b.PrevBlockHash))
			}
			if t.myLastHeader == nil || bytes.Compare(t.myLastHeader.Hash, b.PrevBlockHash) == 0 {
				bsLogger.Trace("Going to ask for Block: %v: %v", b.HashString(), b)
				t.downloadBlockChannel <- b
				t.myLastHeader = b
				bsLogger.Trace("My Last hash set to: %v with height: %v", utils.BytesToString(t.myLastHeader.Hash), t.myLastHeader.Height)
				t.downloadHeaderPendingMap.Delete(b.Height)
			} else {
				t.negotiateCounter++
				bsLogger.Error("BOOOOOM ! got a header that has the same height but cannot be connected to the chain, negotitateCounter: %v", t.negotiateCounter)
				if t.negotiateCounter > NEED_NEGOTIATION_THRESHOLD {
					// TODO: negotiate with the best peer and ready to revert.
					bsLogger.Fatal("Fatal ! got BIFURCATION")
					go t.Negotiate(ctx)
				} else {
					// clean up the queue and resync
					t.downloadBlockHeaderQueue.Clear()
					ds.ClearSyncMap(&t.downloadHeaderPendingMap)
					bsLogger.Error("Cleared. :%v left", ds.GetSyncMapSize(&t.downloadHeaderPendingMap))
					myLastBlock := GetValueChainNode().Chain.GetMyLastBlock()
					if myLastBlock != nil {
						t.myLastHeader = &myLastBlock.Header
						t.lastRequestedHeight = t.myLastHeader.Height
					} else {
						t.lastRequestedHeight = -1
					}
					bsLogger.Error("Cleared.")
				}
			}
		}
	}
}

func (t *BlockSyncMiddleware) GetBestPeer(ctx *tcp.P2PContext) *node.RemoteNode {
	var ret *node.RemoteNode = nil
	peers := ctx.NodeProvider().GetNearbyNodes(20)
	var maxHeight int64 = -1
	for _, p := range peers {
		if _h, ok := t.peersHeights.Load(p.GetRemoteIP().String()); ok {
			if h, ok := _h.(int64); ok {
				if maxHeight <= h {
					maxHeight = h
					ret = p
				}
			}
		}
	}

	return ret
}

// 1. stop the syncing
// 2. find the bifurcation --- log(n)
//		a. compare hash of current height / 2 with same height on bestpeer
//			same:
//				compare hash of current height * 3 / 4 
//			different:
//				compare hash of current height / 4
//		b. loop through the cursor stopped at the bifurcation
//
// 3. revert the chain to the bifurcation
// 4. start syncing
func (t *BlockSyncMiddleware) Negotiate(ctx *tcp.P2PContext) {
	t.SetSyncState(SYNC_STATE_NEGOTIATION)
	// 1 stop
	bsLogger.Trace("Negotiate - stopping the sync")
	t.syncLoopExitSignal <- struct{}{}

	// 2 find
	bsLogger.Trace("Negotiate - finding the bifurcation")
	leftHeight := int64(0)
	rightHeight := GetValueChainNode().Chain.GetMyHeight()
	var height int64
	for {
		height = (rightHeight + leftHeight) / 2
		if (leftHeight == height) {
			// got it
			break
		}
		ctx.SendToPeer(diagram.NewBlockHeaderDiagram(ctx, height), t.selectedPeer)

		_ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
		defer cancel()
		
		// wait for response - read header from chan.
		select {
		case header := <- t.negotiationHeaderChan:
			if header.Height == height {
				if bytes.Compare(header.Hash, GetValueChainNode().Chain.GetBlockByHeight(height).Header.Hash) == 0 {
					// good
					leftHeight = height
				} else {
					rightHeight = height
				}
			} else {
				// ??? response from previous state.
			}
		case <- _ctx.Done():
			continue
		}
	}

	// 3 revert
	bsLogger.Trace("Negotiate - found bifurcation at height: %v, reverting my chain", height)
	block.RevertTo(height)

	t.SetSyncState(SYNC_STATE_IDLE)
	// 4 sync
	bsLogger.Trace("Negotiate - start sync again")
	go t.syncLoop(ctx)
}


func (t *BlockSyncMiddleware) DropConnection(conn *tcp.TCPConnection) {
	t.BaseMiddleware.DropConnection(conn)
}

func (t *BlockSyncMiddleware) DashboardData() interface{} { return [][]string{
	[]string{ "Current Block Height", fmt.Sprintf("%v", GetValueChainNode().Chain.GetMyHeight()) },
	[]string{ "Pending Download Headers Count", fmt.Sprintf("%v", ds.GetSyncMapSize(&t.downloadHeaderPendingMap)) },
	[]string{ "Pending Download Headers Task Count", fmt.Sprintf("%v", t.downloadBlockHeaderQueue.GetRunningTasksCount()) },
	// []string{ "Pending Download Headers Timeout Count", fmt.Sprintf("%v", ds.GetSyncMapSize(&b.downloadHeaderPendingTimeoutMap)) },
	[]string{ "Pending Download Blocks Count", fmt.Sprintf("%v", ds.GetSyncMapSize(&t.downloadBlockPendingMap)) },
	[]string{ "Pending Download Blocks Task Count", fmt.Sprintf("%v", t.downloadBlockQueue.GetRunningTasksCount()) },
	[]string{ "Current Sync State", fmt.Sprintf("%v", t.GetSyncState()) },
} }
func (t *BlockSyncMiddleware) DashboardType() string { return "table" }
func (t *BlockSyncMiddleware) DashboardTitle() string { return "Block Syncing" }
func (t *BlockSyncMiddleware) DashboardTableHasColumnTitles() bool { return false }

func (t *BlockSyncMiddleware) GetSyncState() int {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	bsLogger.Trace("Retrived state: %v", t.currentState)
	return t.currentState
}

func (t *BlockSyncMiddleware) SetSyncState(state int) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	if t.currentState != state {
		bsLogger.Trace("State changed to %v", state)
		t.currentState = state
	}
}
