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
)

var bsLogger = log.GetLogger("blockSyncMiddleware").SetLevel(log.INFO)

var (
	MAX_HEADER_PENDING	=	512
	MAX_BLOCK_PENDING	=	256
	REQUEST_TIMEOUT		=	3 * time.Minute
)

type BlockSyncMiddleware struct {
	*tcp.BaseMiddleware
	syncLoopExitSignal chan struct{}
	downloadBlockChannel chan *block.BlockHeader
	downloadBlockHeaderChannel chan int64
	downloadBlockQueue				*ds.SequentialParallelTaskQueue
	downloadBlockHeaderQueue		*ds.SequentialParallelTaskQueue
	downloadBlockPendingMap			sync.Map
	downloadBlockPendingTimeoutMap	sync.Map
	donwloadHeaderPendingMap		sync.Map
	downloadHeaderPendingTimeoutMap	sync.Map
	knownMaxHeight					int64
}

func NewBlockSyncMiddleware() *BlockSyncMiddleware {
	return &BlockSyncMiddleware{
		BaseMiddleware: tcp.NewBaseMiddleware(),
		syncLoopExitSignal: make(chan struct{}),
		downloadBlockChannel: make(chan *block.BlockHeader, 100),
		downloadBlockHeaderChannel: make(chan int64, 100),
	}
}

func (t *BlockSyncMiddleware) Start(ctx *tcp.P2PContext) {
	t.downloadBlockHeaderQueue = ds.NewSequentialParallelTaskQueue(100, func (tasks []*ds.ParallelTask) {
		// we got headers, need to download blocks.
		bsLogger.Debug("We got finished headers, going to download content of them.")
		for _, task := range tasks {
			if b, ok := task.Result.(*block.BlockHeader); ok {
				bsLogger.Trace("Going to ask for Block: %v: %v", b.HashString(), b)
				t.downloadBlockChannel <- b
				t.donwloadHeaderPendingMap.Delete(b.Height)
			}
		}
	})
	t.downloadBlockQueue = ds.NewSequentialParallelTaskQueue(100, func (tasks []*ds.ParallelTask) {
		// we got blocks, need to save them.
		// defer func() {
		// 	if err := recover(); err != nil {
		// 		bsLogger.Error("%v", err)
		// 	}
		// }()

		bsLogger.Debug("We got finished blocks, going to save them.")
		for _, task := range tasks {
			if b, ok := task.Result.(*block.Block); ok {
				bsLogger.Trace("Going to save Block: %v: %v", b.Header.HashString(), b.Header)
				GetValueChainNode().Chain.SaveBlock(b)
				bsLogger.Trace("Block saved: %v", b.Header.HashString())
				// restart mining if mining is true
				if GetValueChainNode().Miner.IsMining() && !GetValueChainNode().Miner.IsIdle() {
					bsLogger.Info("Cancelled mining !")
					GetValueChainNode().Miner.StopMining()
					GetValueChainNode().Miner.StartMining()
				}
				t.downloadBlockPendingMap.Delete(b.Header.HashString())
			}
		}
	})
	t.downloadBlockHeaderQueue.Execute()
	t.downloadBlockQueue.Execute()
	t.regHandlers()
	go t.syncLoop(ctx)
	go t.mapCheckingLoop(ctx)
}

func (t *BlockSyncMiddleware) regHandlers() {
	t.HandleRequest(diagram.BLOCK_SYNC, func (ctx *tcp.P2PContext) {
		// 1. check the block height with mine,
		//   if higher than me, send BLOCK_SYNC to it
		//   else, send headers by BLOCK_SYNC_RES
		
		var syncDiag diagram.BlockSyncDiagram
		err := ctx.GetDiagram(&syncDiag)
		if err == nil {
			myHeight := GetValueChainNode().Chain.GetMyHeight()
			myLastBlock := GetValueChainNode().Chain.GetMyLastBlock()
			targetHeight := syncDiag.LastBlockHeader.Height
			bsLogger.Debug("Got BLOCK_SYNC diag: %v, myHeight: %v, myLastBlock: %v", syncDiag.LastBlockHeader, myHeight, myLastBlock)
			if targetHeight > myHeight {
				GetValueChainNode().IsSyncing = true
				// ask for headers
				// devide the load to peers
				// t.downloadBlockHeaderQueue.AddTask(&ds.ParallelTask{
				// 	Body: func(params []interface{}, cb func(res interface{})) {
				// 		ctx.SendToPeer()
				// 	},
				// })
				bsLogger.Trace("Asking for headers")
				expectedHeight := utils.Max(myHeight + int64(MAX_HEADER_PENDING), targetHeight)
				for i := myHeight + 1; i <= expectedHeight; i++ {
					t.downloadBlockHeaderChannel <- i
				}
			}

			if targetHeight < myHeight && myLastBlock != nil {
				// Just answer the BLOCK_SYNC message
				ctx.Send(diagram.NewBlockSyncDiagram(ctx, &myLastBlock.Header))
				// bsLogger.Trace("Providing headers")
				// blockHeaders := make([]block.BlockHeader, 0, 0)
				// iterator := GetValueChainNode().Chain.chain.Iterator();
				// for b := iterator.Next(); b != nil; b = iterator.Next() {
				// 	if b.Header.Height > targetHeight {
				// 		blockHeaders = append(blockHeaders, b.Header)
				// 	}
				// }

				// sort.Slice(blockHeaders[:], func(i, j int) bool {
				// 	return blockHeaders[i].Height < blockHeaders[j].Height
				// })
				
				// ctx.Send(diagram.NewBlockHeaderResDiagram(ctx, targetHeight, myHeight, blockHeaders))
				// bsLogger.Trace("Provided headers count: %v", len(blockHeaders))
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

			// blockHeaders := make([]*block.BlockHeader, 0, 0)
			// iterator := GetValueChainNode().Chain.chain.Iterator();
			// for b := iterator.Next(); b != nil; b = iterator.Next() {
			// 	if b.Header.Height >= headerDiag.HeightFrom && b.Header.Height <= headerDiag.HeightTo {
			// 		blockHeaders = append(blockHeaders, b.Header)
			// 	}
			// 	if b.Header.Height < headerDiag.HeightFrom {
			// 		break
			// 	}
			// }

			// sort.Slice(blockHeaders[:], func(i, j int) bool {
			// 	return blockHeaders[i].Height < blockHeaders[j].Height
			// })

			b := GetValueChainNode().Chain.GetBlockByHeight(headerDiag.BlockHeight)
			
			ctx.Send(diagram.NewBlockHeaderResDiagram(ctx, b.Header.Height, &b.Header))
			bsLogger.Trace("Provided header: %v", b.Header.HashString())
		}
	})

	t.HandleRequest(diagram.BLOCK_HEADER_RES, func (ctx *tcp.P2PContext) {
		// 1. store the headers and ask for content with BLOCK_REQ
		var syncDiag diagram.BlockHeaderResDiagram
		err := ctx.GetDiagram(&syncDiag)
		if err == nil {
			bsLogger.Trace("Recieved headers")
			if _cb, _ok := t.donwloadHeaderPendingMap.Load(syncDiag.BlockHeight); _ok {
				bsLogger.Debug("Header Task is in pending map")
				if cb, ok := _cb.(func(res interface{})); ok {
					bsLogger.Debug("Recieved Header and going to ask for block!")
					cb(syncDiag.BlockHeader)
					// remove timeout map
					t.downloadHeaderPendingTimeoutMap.Delete(syncDiag.BlockHeight)
				}
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
				ctx.Send(diagram.NewBlockReqResDiagram(ctx, GetValueChainNode().Chain.GetBlock(bReqDiag.BlockHeader.Hash)))
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
				t.downloadBlockPendingTimeoutMap.Store(bReqResDiag.Block.Header.HashString(), time.Now())
			} else {
				// bsLogger.Info("callback map: %v", t.downloadBlockPendingMap)
				if _cb, _ok := t.downloadBlockPendingMap.Load(bReqResDiag.Block.Header.HashString()); _ok {
					bsLogger.Debug("Task is in pending map")
					if cb, ok := _cb.(func(res interface{})); ok {
						bsLogger.Debug("Recieved Block and going to store !")
						cb(bReqResDiag.Block)
						// remove timeout map
						t.downloadBlockPendingTimeoutMap.Delete(bReqResDiag.Block.Header.HashString())
					}
				}
			}
		}
	})

	t.HandleRequest(diagram.BLOCK_SEND, func (ctx *tcp.P2PContext) {
		// 1. check if I have this block or asking for this block
		//   if not, ask for this block's content with BLOCK_REQ
	})
}

// broadcast BLOCK_SYNC periodly
func (t *BlockSyncMiddleware) syncLoop(ctx *tcp.P2PContext) {
	exit:
	for {
		time.Sleep(10 * time.Second)
		peers := ctx.NodeProvider().GetNearbyNodes(20)
		peersLength := len(peers)
		var randPeer *node.RemoteNode
		if peersLength > 0 {
			randPeer = peers[rand.Intn(peersLength)]
		}
		select {
		case <- t.syncLoopExitSignal:
			break exit
		default:
			bsLogger.Trace("Going to broadcast sync message")
			myLastBlock := GetValueChainNode().Chain.GetMyLastBlock()
			if myLastBlock != nil {
				bsLogger.Trace("Broadcasting sync message")
				ctx.BroadcastToNearbyNodes(diagram.NewBlockSyncDiagram(ctx, &myLastBlock.Header), 20, nil)
			} else {
				ctx.BroadcastToNearbyNodes(diagram.NewBlockSyncDiagram(ctx, &block.BlockHeader{ Height: -1 }), 20, nil)
			}

			t.downloadBlockPendingTimeoutMap.Range(func(k, v interface{}) bool {
				if ti, ok := v.(time.Time); ok {
					if time.Since(ti) >= REQUEST_TIMEOUT {
						// boom
						if header, ok := k.(block.BlockHeader); ok {
							// ask from another guy, it's randomized, maybe it's another guy.
							if randPeer != nil {
								bsLogger.Trace("Block Asking from %v !!!!!", randPeer.GetRemoteIP())
								ctx.SendToPeer(diagram.NewBlockReqDiagram(ctx, &header), randPeer)
								t.downloadBlockPendingTimeoutMap.Store(header.HashString(), time.Now())
							}
						}
					}
				}
				return true
			})

			t.downloadHeaderPendingTimeoutMap.Range(func(k, v interface{}) bool {
				if ti, ok := v.(time.Time); ok {
					if time.Since(ti) >= REQUEST_TIMEOUT {
						// boom
						if height, ok := k.(int64); ok {
							// ask from another guy, it's randomized, maybe it's another guy.
							if randPeer != nil {
								bsLogger.Trace("Header Asking from %v !!!!!", randPeer.GetRemoteIP())
								ctx.SendToPeer(diagram.NewBlockHeaderDiagram(ctx, height), randPeer)
								t.downloadHeaderPendingTimeoutMap.Store(height, time.Now())
							}
						}
					}
				}
				return true
			})
		}
	}
}

// check timeout map periodly
func (t *BlockSyncMiddleware) mapCheckingLoop(ctx *tcp.P2PContext) {
	for {
		peers := ctx.NodeProvider().GetNearbyNodes(20)
		peersLength := len(peers)
		var randPeer *node.RemoteNode
		if peersLength > 0 {
			randPeer = peers[rand.Intn(peersLength)]
		}

		select {
		case header := <- t.downloadBlockChannel:
			bsLogger.Trace("added task to download block: %v", header.HashString())
			t.downloadBlockQueue.AddTask(&ds.ParallelTask{
				Body: func(params []interface{}, cb func(res interface{})) {
					bsLogger.Trace("Going to ask for block!: %v", header.HashString())
					t.downloadBlockPendingTimeoutMap.Store(header.HashString(), time.Now())
					bsLogger.Debug("Storing to downloadBlockPendingMap with key: %v", header.HashString())
					t.downloadBlockPendingMap.Store(header.HashString(), cb)
					if randPeer != nil {
						bsLogger.Trace("Asking for block!: %v", header.HashString())
						ctx.SendToPeer(diagram.NewBlockReqDiagram(ctx, header), randPeer)
					}
				},
			})
		case height := <- t.downloadBlockHeaderChannel:
			bsLogger.Trace("added task to download header: %v", height)
			t.downloadBlockHeaderQueue.AddTask(&ds.ParallelTask{
				Body: func(params []interface{}, cb func(res interface{})) {
					bsLogger.Trace("Going to ask for header!: %v", height)
					t.downloadHeaderPendingTimeoutMap.Store(height, time.Now())
					bsLogger.Debug("Storing to downloadBlockHeaderPendingMap with key: %v", height)
					t.donwloadHeaderPendingMap.Store(height, cb)
					if randPeer != nil {
						bsLogger.Trace("Asking for header!: %v", height)
						ctx.SendToPeer(diagram.NewBlockHeaderDiagram(ctx, height), randPeer)
					}
				},
			})
		default:
			time.Sleep(time.Millisecond)
			continue
		}
	}
}

func (b *BlockSyncMiddleware) DashboardData() interface{} { return [][]string{
	[]string{ "Current Block Height", fmt.Sprintf("%v", GetValueChainNode().Chain.GetMyHeight()) },
	[]string{ "Pending Download Blocks Count", fmt.Sprintf("%v", ds.GetSyncMapSize(&b.downloadBlockPendingMap)) },
} }
func (b *BlockSyncMiddleware) DashboardType() string { return "table" }
func (b *BlockSyncMiddleware) DashboardTitle() string { return "Block Syncing" }
func (b *BlockSyncMiddleware) DashboardTableHasColumnTitles() bool { return false }
