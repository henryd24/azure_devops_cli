package main

import (
	"azuredevops/cmd"
	_ "azuredevops/cmd/interactive"
	_ "azuredevops/cmd/pipelines"
	_ "azuredevops/cmd/security"
	_ "azuredevops/cmd/variable_group"
)

func main() {
	cmd.Execute()
}
