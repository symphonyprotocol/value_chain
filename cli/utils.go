package cli

import (
	"strings"
)

func getArgumentsByStrings(s []string) []IArgument {
	defer func() {
		if err := recover(); err != nil {
			cliLogger.Error("%v", err)
		}
	}()

	cliLogger.Trace("Getting args from %v", s)
	res := make([]IArgument, 0, 0)
	for _, _arg := range s {
		for _, _predefined_arg := range arguments {
			if strings.HasPrefix(_arg, _predefined_arg.Text()) {
				cliLogger.Trace("analyzing the arg: %v, with %v", _arg, _predefined_arg.Text())
				value := ""
				remains := strings.TrimPrefix(_arg, _predefined_arg.Text())
				cliLogger.Trace("remaining: %v", remains)
				if strings.HasPrefix(remains, "=") || strings.HasPrefix(remains, ":") {
					value = remains[1:]
					if strings.HasPrefix(value, "\"") {
						value = value[1:]
					} 
					if strings.HasSuffix(value, "\"") {
						value = value[:len(value)-1]
					}
				}
				cliLogger.Trace("result: %v", value)
				_predefined_arg.SetValue(value)
				res = append(res, _predefined_arg)
			}
		}
	}

	return res
}

func getArgument(a []IArgument, s string) (IArgument, bool) {
	for _, arg := range a {
		if strings.HasPrefix(arg.Text(), s) {
			return arg, true
		}
	}

	return nil, false
}

func splitWithQuotes(s string) []string {
	wordStart, wordEnd := 0, 0
	words := make([]string, 0, 0)
	inWord := false
	inQuote := false
	for i := 0; i < len(s); i++ {
		c := s[i:i+1]
		if c == " " && inWord && !inQuote {
			inWord = false
			wordEnd = i
		}

		if c != " " && !inWord {
			inWord = true
			wordStart = i
		}

		if c== "\"" {
			inQuote = !inQuote
		}

		//cliLogger.Trace("current: %v, wordStart: %v, wordEnd: %v, inQuote: %v, total: %v", c, wordStart, wordEnd, inQuote, len(s))

		if i == len(s) - 1 {
			wordEnd = len(s)
		}

		if wordEnd - wordStart > 0 {
			word := strings.Trim(s[wordStart:wordEnd], " ")
			if word != "" {
				words = append(words, word)
			}
		}
	}

	return words
}

func findLastCommand(words []string, hasEmptySuffix bool) (int, string) {
	cmdIndex := 0
	for i := len(words) - 1; i >= 0; i-- {
		if strings.HasPrefix(words[i], "-") {
			continue
		} else {
			// this is a command
			cmdIndex = i
			cliLogger.Trace("found last command: %v", words[cmdIndex])
			if cmdIndex == len(words) - 1 && cmdIndex > 0 && !hasEmptySuffix {
				cliLogger.Trace("adjust last command to: %v", words[cmdIndex - 1])
				cmdIndex = cmdIndex - 1
			}
			return cmdIndex, words[cmdIndex]
		}
	}

	return -1, ""
}

func findAllCommands(words []string) []string {
	res := make([]string, 0, 0)
	for _, cmd := range words {
		if strings.HasPrefix(cmd, "-") {
			continue
		} else {
			// this is a command
			res = append(res, cmd)
		}
	}

	return res
}
