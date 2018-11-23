package cli

import (
	"strings"
	"fmt"
	"os"
)

func Executor(s string) {
	trimmedS := strings.TrimSpace(s)
	if trimmedS == "" {
		return
	} else if trimmedS == "quit" || trimmedS == "exit" {
		fmt.Println("Bye!")
		os.Exit(0)
		return
	}

	words := splitWithQuotes(s)
	cmdIndex, lastCmd := findLastCommand(words, true)
	allCmds := findAllCommands(words)

	for _, cmd := range commands {
		if lastCmd == cmd.Text() {
			// check its followedby
			failed := true
			for j := cmdIndex - 1; j >= 0; j-- {
				followedBy := cmd.FollowedBy()[len(cmd.FollowedBy()) - cmdIndex + j]
				for _, _cmd := range commands {
					if _cmd.Text() == followedBy && followedBy != "" && cmdIndex > 0 && allCmds[cmdIndex - 1] == followedBy {
						// good
						cliLogger.Trace("Good for %v", followedBy)
						failed = false
						break
					} else {
						// boom
						cliLogger.Debug("BOOM")
					}
				}
			}
			if !failed || len(cmd.FollowedBy()) == 0 {
				cliLogger.Debug("Executing")
				cmd.Execute(words[:cmdIndex], getArgumentsByStrings(words[cmdIndex + 1:]))
			}
		}
	}
}
