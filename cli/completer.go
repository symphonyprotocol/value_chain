package cli

import (
	"strings"
	"github.com/c-bata/go-prompt"
)

var commands = []ICommand {
	__cmd_inst_account,
	__cmd_inst_account_new,
	__cmd_inst_account_getkey,
	__cmd_inst_account_derive,
}

var arguments = []IArgument {
	__arg_inst_account_getkey_m,
	__arg_inst_account_getkey_p,
	__arg_inst_account_derive_path,
	__arg_inst_account_derive_pwd,
}

func Completer(d prompt.Document) []prompt.Suggest {
	words := splitWithQuotes(d.TextBeforeCursor())
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
		hasEmptySuffix := strings.HasSuffix(d.TextBeforeCursor(), " ")
		if lenWords == 1 && !hasEmptySuffix {
			// root commands
			if len(cmd.FollowedBy()) == 0 {
				suggests = appendSuggests(suggests, cmd)
			}
		}

		if lenWords > 1 || hasEmptySuffix {
			_, lastCmd := findLastCommand(words, hasEmptySuffix)
			cliLogger.Trace("last cmd: %v", lastCmd)
			if cmd.Text() == lastCmd {
				for _, subcmd := range cmd.Subcommands() {
					isSub := false
					cliLogger.Trace("checking sub cmd: %v", subcmd)
					var _subcmd ICommand = nil
					SUBFLAG:
					for _, _cmd := range commands {
						if _cmd.Text() == subcmd {
							cliLogger.Trace("found sub cmd: %v", _cmd.Text())
							for _, followed := range _cmd.FollowedBy() {
								cliLogger.Trace("checking follwed by: %v", followed)
								if followed == cmd.Text() {
									cliLogger.Debug("found the subcmd: %v", _cmd.Text())
									isSub = true
									_subcmd = _cmd
									break SUBFLAG
								}
							}
						}
					}
					if isSub {
						cliLogger.Debug("Appending cmd: %v", _subcmd.Text())
						suggests = appendSuggests(suggests, _subcmd)
						cliLogger.Debug("Current suggests: %v", suggests)
					}
				}
							
				for _, arg := range cmd.SupportedArguments() {
					for _, _arg := range arguments {
						if arg == _arg.Text() {
							suggests = appendSuggests(suggests, _arg)
						}
					}
				}
			}
		}
	}

	if lenWords > 0 {
		cliLogger.Debug("I'm returning suggests: %v", suggests)
		return prompt.FilterHasPrefix(suggests, d.GetWordBeforeCursor(), true)
	} else {
		return nil
	}
}

func appendSuggests(sl []prompt.Suggest, s ISuggest) []prompt.Suggest {
	return append(sl, prompt.Suggest{ Text: s.Text(), Description: s.Description() })
}
