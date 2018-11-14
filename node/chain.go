package node

import (
	"bytes"
	"github.com/symphonyprotocol/log"
	"github.com/symphonyprotocol/scb/block"
)

var chainLogger = log.GetLogger("chain")

type NodeChain struct {
	chain	*block.Blockchain
}

func LoadNodeChain() (result *NodeChain) {
	the_chain := block.CreateEmptyBlockchain()
	return &NodeChain{
		chain: the_chain,
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
	return false
}

func (c *NodeChain) GetBlock(hash []byte) *block.Block {
	if c.chain != nil {
		iterator := c.chain.Iterator()
		for b := iterator.Next(); b != nil; {
			if bytes.Compare(b.Header.Hash, hash) == 0 {
				return b
			}
		}
	}

	return nil
}

func (c *NodeChain) SaveBlock(the_block *block.Block) {

}

