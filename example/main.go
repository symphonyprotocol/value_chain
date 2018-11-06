package main

import (
	"github.com/symphonyprotocol/simple-node/node"
	"github.com/symphonyprotocol/simple-node/cli"
	"github.com/symphonyprotocol/log"
)

func main() {
	// init simple node
	log.SetGlobalLevel(log.TRACE)
	log.Configure(map[string]([]log.Appender){
		"default": []log.Appender{ log.NewConsoleAppender() },
	})
	node.InitSimpleNode()
	(&cli.CLI{}).Init()
}

