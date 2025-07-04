package main

import (
	"azurevgcli/cmd"
	_ "azurevgcli/cmd/variable_group"
)

func main() {
	cmd.Execute()
}
