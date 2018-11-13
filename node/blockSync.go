package node

import (
	"github.com/symphonyprotocol/log"
	"github.com/symphonyprotocol/scb/block"
	"time"
	"github.com/symphonyprotocol/p2p/tcp"
	"github.com/symphonyprotocol/simple-node/node/diagram"
)

var bsLogger = log.GetLogger("blockSyncMiddleware")

type BlockSyncMiddleware struct {
	*tcp.BaseMiddleware
	syncLoopExitSignal chan struct{}
	downloadBlockChannel chan *block.BlockHeader
	downloadBlockHeaderChannel chan int64
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

			}

			if targetHeight < myHeight && myLastBlock != nil {
				ctx.Send(diagram.NewBlockSyncDiagram(ctx, &myLastBlock.Header))
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
	})

	t.HandleRequest(diagram.BLOCK_REQ_RES, func (ctx *tcp.P2PContext) {
		// 1. verify and store the block
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
		select {
		case <- t.syncLoopExitSignal:
			break exit
		default:
			time.Sleep(5 * time.Minute)
			myLastBlock := GetSimpleNode().Chain.GetMyLastBlock()
			if myLastBlock != nil {
				ctx.BroadcastToNearbyNodes(diagram.NewBlockSyncDiagram(ctx, &myLastBlock.Header), 20, nil)
			}
		}
	}
}
