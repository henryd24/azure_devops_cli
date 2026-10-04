package main

import (
	"azuredevops/cmd"
	_ "azuredevops/cmd/environments"
	_ "azuredevops/cmd/infra"
	_ "azuredevops/cmd/interactive"
	_ "azuredevops/cmd/pipelines"
	_ "azuredevops/cmd/securefiles"
	_ "azuredevops/cmd/security"
	_ "azuredevops/cmd/variable_group"
	_ "azuredevops/cmd/workitems"
)

func main() {
	cmd.Execute()
}
