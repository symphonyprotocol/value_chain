package middleware

import (
	"time"
	"github.com/symphonyprotocol/scb/block"
	"github.com/symphonyprotocol/p2p/tcp"
	"github.com/symphonyprotocol/simple-node/node/diagram"
)

type BlockSyncMiddleware struct {
	*tcp.BaseMiddleware
	syncLoopExitSignal chan struct{}
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
	})

	t.HandleRequest(diagram.BLOCK_SYNC_RES, func (ctx *tcp.P2PContext) {
		// 1. store the headers and ask for content with BLOCK_REQ
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
			ctx.Broadcast(*diagram.NewBlockSyncDiagram(ctx, &block.Block{}))
		}
	}
}
