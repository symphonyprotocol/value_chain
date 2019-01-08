package node

import (
	"github.com/symphonyprotocol/log"
	"github.com/symphonyprotocol/scb/block"
	"bytes"
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

func (c *NodeChain) GetBlockByHeight(height int64) *block.Block {
	// TODO: use index in db to get
	if c.chain != nil {
		it := c.chain.Iterator()
		b := it.Next()
		for ; b != nil; b = it.Next() {
			// chainLogger.Trace("Looping block height: %v", b.Header.Height)
			if b.Header.Height == height {
				return b
			} else if b.Header.Height < height {
				break
			}
		}
	}

	return nil
}

func (c *NodeChain) SaveBlock(theBlock *block.Block) {
	if c.chain != nil { 
		accountTree := theBlock.GetAccountTree(true)
		chainLogger.Info("Account tree equal to I got from net? : %v\n caclulated: %v\n got: %v", 
			bytes.Compare(accountTree.MerkleRoot(), theBlock.Header.MerkleRootAccountHash),
			accountTree.MerkleRoot(),
			theBlock.Header.MerkleRootAccountHash)
		c.chain.AcceptNewBlock(theBlock, accountTree)
	}
}

func (c *NodeChain) SavePendingTx(theTx *block.Transaction) {
	if c.chain != nil {
		c.chain.SaveTransaction(theTx)
	}
}

