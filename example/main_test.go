package main

import (
	"github.com/symphonyprotocol/value_chain/node"
	"github.com/symphonyprotocol/value_chain/cli"
	"github.com/symphonyprotocol/log"
	"github.com/symphonyprotocol/scb/block"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	// init simple node
	log.SetGlobalLevel(log.TRACE)
	log.Configure(map[string]([]log.Appender){
		"default": []log.Appender{ log.NewConsoleAppender() },
	})
	vNode := node.InitValueChainNode()
	go func() {
		// defer func() {
		// 	if err := recover(); err != nil {
		// 		//goto LOOP
		// 		fmt.Printf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!! %v", err)
		// 	}
		// }()
		// start mining
		vNode.Miner.StartMining()
		// make transfers
//LOOP:
		for {
			time.Sleep(10 * time.Second)
			if vNode.P2PServer.GetP2PContext() != nil && vNode.P2PServer.GetP2PContext().LocalNode() != nil {
				for i := 0; i < 1; i++ {
					from := vNode.Accounts.Accounts[0].ECPubKey().ToAddressCompressed()
					to := vNode.Accounts.Accounts[1].ECPubKey().ToAddressCompressed()
					block.SendTo(from, to, 1, vNode.Accounts.ExportAccount(from))
					// fmt.Printf("Initialized transaction: %v\n", tx.IDString())
					// tmpCtx := vNode.P2PServer.GetP2PContext()
					// tmpCtx.BroadcastToNearbyNodes(diagram.NewTransactionSendDiagram(tmpCtx, tx), 20, nil)
					// fmt.Printf("Broadcasted: %v\n", tx.IDString())
				}
			}
		}
	}()
	(&cli.CLI{}).Init()
}

