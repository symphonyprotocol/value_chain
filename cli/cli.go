package cli

import (
	"github.com/c-bata/go-prompt"
	"github.com/c-bata/go-prompt/completer"
	"github.com/symphonyprotocol/log"
)

var cliLogger = log.GetLogger("cli").SetLevel(log.INFO)

type CLI struct {}

func (c *CLI) Init() {
	// prompt.Input("> ", func (d prompt.Document) []prompt.Suggest {
	// 	return prompt.FilterHasPrefix([]prompt.Suggest{
	// 		{ Text: "NewTransaction", Description: "Create a new transaction" },
	// 		{ Text: "Mine", Description: "Switch mine on/off" },
	// 		{ Text: "GetBalance", Description: "Get balance of current account" },
	// 		{ Text: "Print", Description: "Print the node's current statistics" },
	// 	}, d.GetWordBeforeCursor(), true)
	// })

	p := prompt.New(
		Executor,
		Completer,
		prompt.OptionTitle("Symphony protocol: symple-node\n"),
		prompt.OptionPrefix("> "),
		prompt.OptionCompletionWordSeparator(completer.FilePathCompletionSeparator),
	)
	p.Run()
}

