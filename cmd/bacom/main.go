package main

import (
	"fmt"
	"os"
)

func main() {
	cmd, args := getCommand()

	switch cmd {
	default:
		fmt.Fprintf(os.Stderr, "command %q not implemented yet\n", cmd)
	case testCmdName:
		testCmd(args)
	case importCmdName:
		importCmd(args)
	case listCmdName:
		listCmd(args)
	}
}
