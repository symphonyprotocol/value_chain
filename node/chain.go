package node

import (
	"github.com/symphonyprotocol/log"
	"github.com/symphonyprotocol/scb/block"
	"bytes"
)

var chainLogger = log.GetLogger("chain")

type NodeChain struct {
	chain	*block.Blockchain
	// pool	*block.BlockchainPendingPool
}

func LoadNodeChain() (result *NodeChain) {
	theChain := block.CreateEmptyBlockchain()
	// thePool := block.LoadPendingPool()
	return &NodeChain{
		chain: theChain,
		// pool: thePool,
	}
}

func (c *NodeChain) GetMyHeight() int64 {
	if c.chain != nil {
		chainLogger.Trace("Chain is not nil")
		lastBlock := c.GetMyLastBlock()
		if lastBlock != nil {
			chainLogger.Trace("LastBlock is not nil: %v", lastBlock.Header.Height)
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

func (c *NodeChain) GetBlock(hash []byte) *block.Block {
	if c.chain != nil {
		return c.chain.GetBlockByHash(hash)
	}

	return nil
}

func (c *NodeChain) GetBlockByHeight(height int64) *block.Block {
	if c.chain != nil {
		return c.chain.GetBlockByHeight(height)
	}

	return nil
}

func (c *NodeChain) SaveBlock(theBlock *block.Block) {
	if c.chain != nil { 
		accountTree := theBlock.GetAccountTree(true)
		checkRes := bytes.Compare(accountTree.MerkleRoot(), theBlock.Header.MerkleRootAccountHash)
		chainLogger.Info("Account tree equal to I got from net? : %v\n caclulated: %v\n got: %v", 
			checkRes,
			accountTree.MerkleRoot(),
			theBlock.Header.MerkleRootAccountHash)
		if checkRes == 0 || len(block.GetAllAccount()) == 0 {
			c.chain.AcceptNewBlock(theBlock, accountTree)
			chainLogger.Info("Clearing Pending pool") 
			block.ClearPendingPool()
		}
	}
}

func (c *NodeChain) SavePendingTx(theTx *block.Transaction) {
	if c.chain != nil {
		c.chain.SaveTransaction(theTx)
	}
}

func (c *NodeChain) SavePendingBlock(b *block.Block) {
	pool := block.LoadPendingPool()
	if pool != nil {
		
		pendingblockChain := pool.AcceptBlock(b)
		if pendingblockChain != nil{
			bc := block.LoadBlockchain()
			bc.AcceptNewPendingChain(pendingblockChain)
		}
		b.DeleteTransactions()
	}
}

