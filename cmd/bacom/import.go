package main

import (
	"fmt"
	"os"
	"strings"
)

func importCmd(args []string) {
	var cmd string

	cmd, args = getImportSubCommand(args)
	switch cmd {
	default:
		_, err := fmt.Fprintf(os.Stderr, "command %q not implemented yet\n", cmd)
		if err != nil {
			panic(err)
		}
		os.Exit(1)
	case harSubCmdName:
		importHarCmd(args)
	case curlSubCmdName:
		importCurlCmd(args)
	case proxySubCmdName:
		importProxyCmd(args)
	}
}

func getImportSubCommand(args []string) (cmd string, cmdArgs []string) {
	if len(args) == 0 {
		printImportUsage()
		os.Exit(2)
	}
	cmd = args[0]
	cmdArgs = args[1:]

	switch strings.ToLower(cmd) {
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown import sub-command %q\n", cmd)
		os.Exit(2)
	case curlSubCmdName, harSubCmdName, proxySubCmdName:
		return strings.ToLower(cmd), cmdArgs
	}

	return "", nil
}

func printImportUsage() {
	bin := getBinaryName()
	fmt.Fprintf(
		os.Stderr,
		`Usage: %s import [SUB-COMMAND] [OPTIONS]

SUB-COMMANDS:
    har    import requests and responses from har files
    curl   save a request/response pair by providing curl-like arguments

Note:
    "%s import SUB-COMMAND -h" to get an overview of each sub-command's flags

`,
		bin, bin,
	)

}
