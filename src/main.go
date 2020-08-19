package main

import (
	"os"

	"github.com/kr/pretty"
)

func main() {

	var command string
	var args []string

	if len(os.Args) >= 2 {
		command = os.Args[1]
		args = os.Args[2:]
	} else {
		command = "help"
		args = []string{}
	}

	if ImpatienceCommands[command] != nil {
		ImpatienceCommands[command](args)
	} else {
		knownCommands := []string{}

		for command := range ImpatienceCommands {
			knownCommands = append(knownCommands, command)
		}

		pretty.Println("[Impatience] Error!\nUnkown sub command:", command, "!\nPlease use on of the following commands: ", knownCommands)

		ImpatienceCommands["help"]([]string{})
		return
	}

}

// ImpatienceCLICommand add a new CLI command
type ImpatienceCLICommand = func(args []string)

// ImpatienceCommands hold all the callable impatience commands
var ImpatienceCommands map[string]ImpatienceCLICommand = map[string]ImpatienceCLICommand{
	"launch": Launch,
	"help":   Help,
	"init": func(args []string) {
		pretty.Println(
			"[Impatience - Init]:\n",
			"			Generating config file...",
		)

	},
}
