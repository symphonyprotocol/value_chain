package main

import (
	"github.com/symphonyprotocol/value_chain/node"
	"github.com/symphonyprotocol/value_chain/cli"
	"github.com/symphonyprotocol/log"
)

func main() {
	// init simple node
	log.SetGlobalLevel(log.DEBUG)
	log.Configure(map[string]([]log.Appender){
		"default": []log.Appender{ log.NewConsoleAppender() },
	})
	node.InitValueChainNode()
	(&cli.CLI{}).Init()
}

