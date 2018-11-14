package node

import (
	"math/rand"
	"sync"
	"github.com/symphonyprotocol/log"
	"github.com/symphonyprotocol/scb/block"
	"time"
	"github.com/symphonyprotocol/p2p/tcp"
	"github.com/symphonyprotocol/simple-node/node/diagram"
	"github.com/symphonyprotocol/sutil/ds"
)

var bsLogger = log.GetLogger("blockSyncMiddleware")

type BlockSyncMiddleware struct {
	*tcp.BaseMiddleware
	syncLoopExitSignal chan struct{}
	downloadBlockChannel chan *block.BlockHeader
	downloadBlockHeaderChannel chan int64
	downloadBlockQueue *ds.SequentialParallelTaskQueue
	downloadBlockHeaderQueue *ds.SequentialParallelTaskQueue
	downloadBlockPendingMap	sync.Map
	downloadBlockPendingTimeoutMap	sync.Map
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
	})
	t.downloadBlockQueue = ds.NewSequentialParallelTaskQueue(100, func (tasks []*ds.ParallelTask) {
		// we got blocks, need to save them.

	})
	t.downloadBlockHeaderQueue.Execute()
	t.downloadBlockQueue.Execute()
	t.regHandlers()
	go t.syncLoop(ctx)
}

func (t *BlockSyncMiddleware) regHandlers() {
	t.HandleRequest(diagram.BLOCK_SYNC, func (ctx *tcp.P2PContext) {
		// 1. check the block height with mine,
		//   if higher than me, send BLOCK_SYNC to it
		//   else, send headers by BLOCK_SYNC_RES
		
		var syncDiag diagram.BlockSyncDiagram
		err := ctx.GetDiagram(&syncDiag)
		if err == nil {
			myHeight := GetSimpleNode().Chain.GetMyHeight()
			myLastBlock := GetSimpleNode().Chain.GetMyLastBlock()
			targetHeight := syncDiag.LastBlockHeader.Height
			if targetHeight > myHeight {
				// ask for headers
				// devide the load to peers
				// t.downloadBlockHeaderQueue.AddTask(&ds.ParallelTask{
				// 	Body: func(params []interface{}, cb func(res interface{})) {
				// 		ctx.SendToPeer()
				// 	},
				// })

				ctx.Send(diagram.NewBlockSyncDiagram(ctx, &myLastBlock.Header))
			}

			if targetHeight < myHeight && myLastBlock != nil {
				blockHeaders := make([]block.BlockHeader, 0, 0)
				iterator := GetSimpleNode().Chain.chain.Iterator();
				for b := iterator.Next(); b != nil; {
					blockHeaders = append(blockHeaders, b.Header)
				}
				
				ctx.Send(diagram.NewBlockHeaderResDiagram(ctx, targetHeight, myHeight, blockHeaders))
			}

			if targetHeight < myHeight && myLastBlock == nil {
				// ???????
				bsLogger.Error("??????? lower height than me, but I have no blocks???????")
			}
		}
	})

	t.HandleRequest(diagram.BLOCK_HEADER_RES, func (ctx *tcp.P2PContext) {
		// 1. store the headers and ask for content with BLOCK_REQ
		var syncDiag diagram.BlockHeaderResDiagram
		err := ctx.GetDiagram(&syncDiag)
		if err == nil {
			for _, header := range syncDiag.BlockHeaders {
				t.downloadBlockChannel <- &header
			}
		}
	})

	t.HandleRequest(diagram.BLOCK_REQ, func (ctx *tcp.P2PContext) {
		// 1. If I have the block, send the block with BLOCK_REQ_RES
		var bReqDiag diagram.BlockReqDiagram
		err := ctx.GetDiagram(&bReqDiag)
		if err == nil {
			myHeight := GetSimpleNode().Chain.GetMyHeight()
			targetHeight := bReqDiag.BlockHeader.Height
			if myHeight <= targetHeight {
				// boom
			} else {
				ctx.Send(diagram.NewBlockReqResDiagram(ctx, GetSimpleNode().Chain.GetBlock(bReqDiag.BlockHeader.Hash)))
			}
		}
	})

	t.HandleRequest(diagram.BLOCK_REQ_RES, func (ctx *tcp.P2PContext) {
		// 1. verify and store the block
		var bReqResDiag diagram.BlockReqResDiagram
		err := ctx.GetDiagram(&bReqResDiag)
		if err == nil {
			if _cb, _ok := t.downloadBlockPendingMap.Load(bReqResDiag.Block.Header); _ok {
				if cb, ok := _cb.(func(res interface{})); ok {
					cb(&bReqResDiag.Block)
					// remove timeout map
					t.downloadBlockPendingTimeoutMap.Delete(bReqResDiag.Block.Header)
				}
			}
		}
	})

	t.HandleRequest(diagram.BLOCK_SEND, func (ctx *tcp.P2PContext) {
		// 1. check if I have this block or asking for this block
		//   if not, ask for this block's content with BLOCK_REQ
	})
}

// broadcast BLOCK_SYNC periodly, check timeout map periodly
func (t *BlockSyncMiddleware) syncLoop(ctx *tcp.P2PContext) {
	exit:
	for {
		peers := ctx.NodeProvider().GetNearbyNodes(20)
		randPeer := peers[rand.Intn(len(peers))]
		select {
		case <- t.syncLoopExitSignal:
			break exit
		case header := <- t.downloadBlockChannel:
			t.downloadBlockQueue.AddTask(&ds.ParallelTask{
				Body: func(params []interface{}, cb func(res interface{})) {
					t.downloadBlockPendingTimeoutMap.Store(header, time.Now())
					t.downloadBlockPendingMap.Store(header, cb)
					ctx.SendToPeer(diagram.NewBlockReqDiagram(ctx, header), randPeer)
				},
			})
		default:
			time.Sleep(5 * time.Minute)
			myLastBlock := GetSimpleNode().Chain.GetMyLastBlock()
			if myLastBlock != nil {
				ctx.BroadcastToNearbyNodes(diagram.NewBlockSyncDiagram(ctx, &myLastBlock.Header), 20, nil)
			}

			t.downloadBlockPendingTimeoutMap.Range(func(k, v interface{}) bool {
				if ti, ok := v.(time.Time); ok {
					if time.Since(ti) >= time.Minute * 10 {
						// boom
						if header, ok := k.(block.BlockHeader); ok {
							// ask from another guy, it's randomized, maybe it's another guy.
							ctx.SendToPeer(diagram.NewBlockReqDiagram(ctx, &header), randPeer)
							t.downloadBlockPendingTimeoutMap.Store(header, time.Now())
						}
					}
				}
				return true
			})
		}
	}
}
