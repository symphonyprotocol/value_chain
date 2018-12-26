package node

import (
	"github.com/symphonyprotocol/log"
	"github.com/symphonyprotocol/scb/block"
)

var chainLogger = log.GetLogger("chain")

type NodeChain struct {
	chain	*block.Blockchain
}

func LoadNodeChain() (result *NodeChain) {
	theChain := block.CreateEmptyBlockchain()
	return &NodeChain{
		chain: theChain,
	}
}

func (c *NodeChain) GetMyHeight() int64 {
	if c.chain != nil {
		lastBlock := c.GetMyLastBlock()
		if lastBlock != nil {
			return lastBlock.Header.Height
		}
	}
	
	return -1
}

func (c *NodeChain) GetMyLastBlock() *block.Block {
	if c.chain != nil {
		lastBlock := c.chain.Iterator().Next()
		return lastBlock
	}

	return nil
}

func (c *NodeChain) HasBlock(hash []byte) bool {
	//TODO: implementation
	if c.chain != nil {
		return c.chain.HasBlock(hash) != nil
	}
	return false
}

func (c *NodeChain) HasPendingTransaction(id []byte) bool {
	return c.chain.FindUnpackTransactionById(id) != nil
}

// boom. should not be this way.
func (c *NodeChain) GetBlock(hash []byte) *block.Block {
	if c.chain != nil {
		return c.chain.GetBlockByHash(hash)
	}

	return nil
}

func (c *NodeChain) SaveBlock(theBlock *block.Block) {
	if c.chain != nil { 
		c.chain.AcceptNewBlock(theBlock, theBlock.GetAccountTree(true))
	}
}

func (c *NodeChain) SavePendingTx(theTx *block.Transaction) {
	if c.chain != nil {
		c.chain.SaveTransaction(theTx)
	}
}

