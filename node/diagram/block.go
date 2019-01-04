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
	// BLOCK_SYNC_RES = "/block/sync/res"

	BLOCK_HEADER = "/block/header"

	BLOCK_HEADER_RES = "/block/header/res"

	// msg broadcasted with block header from whom mined the block.
	BLOCK_SEND = "/block/send"

	// when node has no such block, ask for details
	BLOCK_REQ = "/block/req"

	// send the details to the requester
	BLOCK_REQ_RES = "/block/req/res"
)

type BlockSyncDiagram struct {
	*models.TCPDiagram
	LastBlockHeader	*block.BlockHeader
}

func NewBlockSyncDiagram(ctx *tcp.P2PContext, _blockHeader *block.BlockHeader) *BlockSyncDiagram {
	tDiag := ctx.NewTCPDiagram()
	tDiag.DType = BLOCK_SYNC
	return &BlockSyncDiagram{
		TCPDiagram: tDiag,
		LastBlockHeader: _blockHeader,
	}
}

type BlockHeaderDiagram struct {
	*models.TCPDiagram
	BlockHeight	int64
}

func NewBlockHeaderDiagram(ctx *tcp.P2PContext, _blockHeight int64) *BlockHeaderDiagram {
	tDiag := ctx.NewTCPDiagram()
	tDiag.DType = BLOCK_HEADER
	return &BlockHeaderDiagram{
		TCPDiagram: tDiag,
		BlockHeight: _blockHeight,
	}
}

type BlockHeaderResDiagram struct {
	*models.TCPDiagram
	BlockHeight	int64
	BlockHeader *block.BlockHeader
}

func NewBlockHeaderResDiagram(ctx *tcp.P2PContext, _blockHeight int64, _blockHeader *block.BlockHeader) *BlockHeaderResDiagram {
	tDiag := ctx.NewTCPDiagram()
	tDiag.DType = BLOCK_HEADER_RES
	return &BlockHeaderResDiagram{
		TCPDiagram: tDiag,
		BlockHeight: _blockHeight,
		BlockHeader: _blockHeader,
	}
}

type BlockReqDiagram struct {
	*models.TCPDiagram
	BlockHeader	*block.BlockHeader
}

func NewBlockReqDiagram(ctx *tcp.P2PContext, _blockHeader *block.BlockHeader) *BlockReqDiagram {
	tDiag := ctx.NewTCPDiagram()
	tDiag.DType = BLOCK_REQ
	return &BlockReqDiagram{
		TCPDiagram: tDiag,
		BlockHeader: _blockHeader,
	}
}

type BlockReqResDiagram struct {
	*models.TCPDiagram
	Block 	*block.Block
}

func NewBlockReqResDiagram(ctx *tcp.P2PContext, _block *block.Block) *BlockReqResDiagram {
	tDiag := ctx.NewTCPDiagram()
	tDiag.DType = BLOCK_REQ_RES
	return &BlockReqResDiagram{
		TCPDiagram: tDiag,
		Block:	_block,
	}
}

type BlockSendDiagram struct {
	*models.TCPDiagram
	BlockHeader *block.BlockHeader
}

func NewBlockSendDiagram(ctx *tcp.P2PContext, _blockHeader *block.BlockHeader) *BlockSendDiagram {
	tDiag := ctx.NewTCPDiagram()
	tDiag.DType = BLOCK_SEND
	return &BlockSendDiagram{
		TCPDiagram: tDiag,
		BlockHeader: _blockHeader,
	}
}

