package cli

import (
)

type ISuggest interface {
	Text()	string
	Description()	string
}

type ICommand interface {
	ISuggest
	Subcommands()	[]string
	FollowedBy()	[]string
	SupportedArguments()	[]string
	Execute(previousCmds []string, args []IArgument)	
}

type IArgument interface {
	ISuggest
	GetValue() string
	SetValue(v string)
}

type BaseArgument struct { _value string }
func (a *BaseArgument) GetValue() string { 
	cliLogger.Trace("Getting value: %v from %v", a._value, &a)
	return a._value 
}
func (a *BaseArgument) SetValue(v string) {
	cliLogger.Trace("Setting value: %v to %v", v, &a)
	a._value = v 
}

