package variable_group

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"fmt"
	"net/url"
)

// CreateVariableGroup crea un Variable Group nuevo en el proyecto del cliente.
func CreateVariableGroup(client *azdevops.Client, name string, variables map[string]models.VariableVal, description string) (*models.VariableGroup, error) {
	if variables == nil {
		variables = map[string]models.VariableVal{}
	}
	group := models.VariableGroupById{
		Name:                           name,
		Type:                           "Vsts",
		Project:                        client.Project,
		Variables:                      variables,
		VariableGroupProjectReferences: models.ConstructVariableGroupProjectReferences(client.Project, name, description),
	}

	var created models.VariableGroup
	u := client.ProjectURL("distributedtask/variablegroups", url.Values{"api-version": {apiVersion}})
	if err := client.Do("POST", u, group, &created); err != nil {
		return nil, fmt.Errorf("error al crear el Variable Group '%s': %w", name, err)
	}
	return &created, nil
}
