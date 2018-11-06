package cli

import (
	"strings"
	"fmt"
	"os"
)

func Executor(s string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return
	} else if s == "quit" || s == "exit" {
		fmt.Println("Bye!")
		os.Exit(0)
		return
	}

	words := splitWithQuotes(s)
	if len(words) > 0 {
		cmdIndex := 0
		for i := len(words) - 1; i >= 0; i-- {
			if strings.HasPrefix(words[i], "-") {
				continue
			} else {
				// this is a command
				cmdIndex = i
				cliLogger.Trace("found last command: %v", words[cmdIndex])
				break
			}
		}
		if cmd, ok := commands[words[cmdIndex]]; ok {
			cmd.Execute(words[:cmdIndex], getArgumentsByStrings(words[cmdIndex + 1:]))
		}
	}
}
