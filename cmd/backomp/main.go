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
	case "test":
		testCmd(args)
	case "import":
		importCmd(args)
	}
}
