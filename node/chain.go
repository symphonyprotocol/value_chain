package node

import (
	"github.com/symphonyprotocol/log"
	"github.com/symphonyprotocol/scb/block"
	"github.com/symphonyprotocol/sutil/utils"
	"bytes"
)

var chainLogger = log.GetLogger("chain")

type NodeChain struct {
	chain	*block.Blockchain
	// pool	*block.BlockchainPendingPool
	pendingBlockChan	chan *block.Block
}

// be called only once.
func LoadNodeChain() (result *NodeChain) {
	theChain := block.CreateEmptyBlockchain()
	// thePool := block.LoadPendingPool()
	c := &NodeChain{
		chain: theChain,
		pendingBlockChan: make(chan *block.Block),
		// pool: thePool,
	}
	go c.loop()
	return c
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
	c.chain.SaveTransaction(theTx)
}

func (c *NodeChain) SavePendingBlock(b *block.Block) {
	c.pendingBlockChan <- b
}

func (c *NodeChain) loop() {
	for {
		select {
		case b := <- c.pendingBlockChan:
			chainLogger.Trace("====== received block with hash: %v \n    and prevHash: %v", b.Header.HashString(), utils.BytesToString(b.Header.PrevBlockHash))
			if GetValueChainNode().Miner.IsMining() {
				bsLogger.Debug("Cancelled mining")
				GetValueChainNode().Miner.StopMining()
				c.savePendingBlock(b)
				GetValueChainNode().Miner.StartMining()
			} else {
				c.savePendingBlock(b)
			}
		}
	}
}

func (c *NodeChain) savePendingBlock(b *block.Block) {
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

