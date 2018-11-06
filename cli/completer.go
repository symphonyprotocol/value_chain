package cli

import (
	"strings"
	"github.com/c-bata/go-prompt"
)

var commands = map[string]ICommand {
	__cmd_inst_account.Text(): __cmd_inst_account,
	__cmd_inst_account_new.Text(): __cmd_inst_account_new,
	__cmd_inst_account_getkey.Text(): __cmd_inst_account_getkey,
	__cmd_inst_account_derive.Text(): __cmd_inst_account_derive,
}

var arguments = map[string]IArgument {
	__arg_inst_account_getkey_m.Text(): __arg_inst_account_getkey_m,
	__arg_inst_account_getkey_p.Text(): __arg_inst_account_getkey_p,
	__arg_inst_account_derive_path.Text(): __arg_inst_account_derive_path,
	__arg_inst_account_derive_pwd.Text(): __arg_inst_account_derive_pwd,
}

func Completer(d prompt.Document) []prompt.Suggest {
	words := strings.Split(d.TextBeforeCursor(), " ")
	if d.TextBeforeCursor() == "" {
		return nil
	}

	// if len(words) == 1 {
	// 	// commands
	// 	return prompt.FilterHasPrefix([]prompt.Suggest{
	// 		{ Text: "account", Description: "Account related commands" },
	// 		{ Text: "transaction", Description: "Transaction related commands" },
	// 		{ Text: "status", Description: "Show current status" },
	// 	}, words[0], true)
	// }

	lenWords := len(words)
	suggests := make([]prompt.Suggest, 0, 0)
	for _, cmd := range commands {
		if lenWords == 1 {
			// root commands
			if len(cmd.FollowedBy()) == 0 {
				suggests = appendSuggests(suggests, cmd)
			}
		}
	}
		
	if lenWords > 1 {
		if cmd, ok := commands[words[lenWords - 2]]; ok {
			for _, subcmd := range cmd.Subcommands() {
				suggests = appendSuggests(suggests, commands[subcmd])
			}

			for _, arg := range cmd.SupportedArguments() {
				suggests = appendSuggests(suggests, arguments[arg])
			}
		}
	}

	if lenWords > 0 {
		return prompt.FilterHasPrefix(suggests, words[lenWords - 1], true)
	} else {
		return nil
	}
}

func appendSuggests(sl []prompt.Suggest, s ISuggest) []prompt.Suggest {
	return append(sl, prompt.Suggest{ Text: s.Text(), Description: s.Description() })
}
