package variable_group

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"fmt"
	"net/url"
	"sort"
)

const apiVersion = "7.1-preview.2"

// ListVariableGroups lista los Variable Groups del proyecto. filter admite comodines
// (p. ej. "app-*"); vacío devuelve todos.
func ListVariableGroups(client *azdevops.Client, filter string) ([]models.VariableGroup, error) {
	q := url.Values{"api-version": {apiVersion}}
	if filter != "" {
		q.Set("groupName", filter)
	}
	var result struct {
		Value []models.VariableGroup `json:"value"`
	}
	if err := client.Do("GET", client.ProjectURL("distributedtask/variablegroups", q), nil, &result); err != nil {
		return nil, fmt.Errorf("error al listar los Variable Groups: %w", err)
	}
	sort.Slice(result.Value, func(i, j int) bool { return result.Value[i].Name < result.Value[j].Name })
	return result.Value, nil
}

// GetVariableGroupByName busca Variable Groups por nombre (admite comodines).
func GetVariableGroupByName(client *azdevops.Client, name string) ([]models.VariableGroup, error) {
	if name == "" {
		return nil, fmt.Errorf("el nombre del Variable Group no puede estar vacío")
	}
	return ListVariableGroups(client, name)
}

// GetVariableGroupById obtiene un Variable Group por su ID.
func GetVariableGroupById(client *azdevops.Client, id int) (*models.VariableGroup, error) {
	var result models.VariableGroup
	u := client.ProjectURL(fmt.Sprintf("distributedtask/variablegroups/%d", id), url.Values{"api-version": {apiVersion}})
	if err := client.Do("GET", u, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
