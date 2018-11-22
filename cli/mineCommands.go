package cli

import (
	"github.com/symphonyprotocol/simple-node/node"
)


type MineCommand struct {
}
func (a *MineCommand) Text() string { return "mine" }
func (a *MineCommand) Description() string { return "Start/stop mining on this node." }
func (a *MineCommand) Subcommands() []string {
	return []string{
		"start", "stop",
	}
}

func (a *MineCommand) SupportedArguments() []string { return []string{} }
func (a *MineCommand) FollowedBy() []string { return []string{} }
func (a *MineCommand) Execute(previousCmds []string, args []IArgument) {
	sNode := node.GetSimpleNode()
	isMining := sNode.Miner.IsMining()
	cliLogger.Info("Current mining state: %v", isMining)
	if isMining {
		sNode.Miner.StopMining()
	} else {
		sNode.Miner.StartMining()
	}
	cliLogger.Info("Mining state changed from %v to %v", isMining, sNode.Miner.IsMining())
}

var __cmd_inst_mine = &MineCommand{}


