package main

import (
	"github.com/symphonyprotocol/simple-node/node"
)

func main() {
	// init simple node
	node.InitSimpleNode()
	(&node.CLI{}).Init()
}

