package diagram

import (
	"github.com/symphonyprotocol/p2p/tcp"
	"github.com/symphonyprotocol/p2p/models"
	"github.com/symphonyprotocol/scb/block"
)

var (
	// msg broadcasted to ask for the block height
	BLOCK_SYNC = "/block/sync"

	// respond to /block/sync with block headers
	BLOCK_SYNC_RES = "/block/sync/res"

	// msg broadcasted with block header from whom mined the block.
	BLOCK_SEND = "/block/send"

	// when node has no such block hash, ask for details
	BLOCK_REQ = "/block/req"

	// send the details to the requester
	BLOCK_REQ_RES = "/block/req/res"
)

type BlockSyncDiagram struct {
	models.TCPDiagram
	
	// TODO: should be blockheader
	LastBlock	block.Block
}

func NewBlockSyncDiagram(ctx *tcp.P2PContext, _block *block.Block) *BlockSyncDiagram {
	tDiag := ctx.NewTCPDiagram()
	tDiag.DType = BLOCK_SYNC
	return &BlockSyncDiagram{
		TCPDiagram: *tDiag,
		LastBlock: *_block,
	}
}


