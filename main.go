package main

import (
	"azuredevops/cmd"
	_ "azuredevops/cmd/variable_group"
)

func main() {
	cmd.Execute()
}
