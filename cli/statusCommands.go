package cli

import (
	"github.com/symphonyprotocol/p2p/models"
	"fmt"
	"github.com/symphonyprotocol/simple-node/node"
)


type StatusCommand struct {}
func (a *StatusCommand) Text() string { return "status" }
func (a *StatusCommand) Description() string { return "Show current node's status." }
func (a *StatusCommand) Subcommands() []string { return []string{} }
func (a *StatusCommand) SupportedArguments() []string { return []string{} }
func (a *StatusCommand) FollowedBy() []string { return []string{} }
func (a *StatusCommand) Execute(previousCmds []string, args []IArgument) {
	output := "Status:\n"
	output += a.getDashboardOutput(node.GetSimpleNode()) + "\n"
	middlewares := node.GetSimpleNode().P2PServer.GetP2PContext().Middlewares()
	for _, middleware := range middlewares {
		output += a.getDashboardOutput(middleware)
		output += "\n"
	}
	cliLogger.Info(output)
}

func (a *StatusCommand) getDashboardOutput(d models.IDashboardProvider) string {
	output := ""
	if d.DashboardType() == "table" {
		if data, ok := d.DashboardData().([][]string); ok {
			output += fmt.Sprintf("<%v>\n", d.DashboardTitle())
			output += "---------------\n"
			for _, line := range data {
				for _, column := range line {
					output += fmt.Sprintf("\t%v\t", column)
				}
				output += "\n"
			}
		}
	}

	return output
}

var __cmd_inst_status = &StatusCommand{}
