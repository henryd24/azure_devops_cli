package variable_group

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"fmt"
	"maps"
	"net/url"
)

// UpdateVariableGroup reemplaza el contenido de un Variable Group con el recibido.
// Las variables secretas sin valor conservan el valor que ya tenían en Azure DevOps.
func UpdateVariableGroup(client *azdevops.Client, group models.VariableGroup) (*models.VariableGroup, error) {
	payload := models.VariableGroupById{
		ID:                             group.Id,
		Name:                           group.Name,
		Type:                           group.Type,
		Project:                        client.Project,
		Variables:                      group.Variables,
		VariableGroupProjectReferences: models.ConstructVariableGroupProjectReferences(client.Project, group.Name, group.Description),
	}
	var updated models.VariableGroup
	u := client.ProjectURL(fmt.Sprintf("distributedtask/variablegroups/%d", group.Id), url.Values{"api-version": {apiVersion}})
	if err := client.Do("PUT", u, payload, &updated); err != nil {
		return nil, fmt.Errorf("error al actualizar el Variable Group '%s': %w", group.Name, err)
	}
	return &updated, nil
}

// AddVariablesToGroup agrega o sobrescribe variables (y opcionalmente la descripción) de un grupo.
func AddVariablesToGroup(client *azdevops.Client, group models.VariableGroup, variables map[string]models.VariableVal, description string) (*models.VariableGroup, error) {
	if description != "" {
		group.Description = description
	}
	merged := make(map[string]models.VariableVal, len(group.Variables)+len(variables))
	maps.Copy(merged, group.Variables)
	maps.Copy(merged, variables)
	group.Variables = merged
	return UpdateVariableGroup(client, group)
}

// RemoveVariablesFromGroup elimina las variables indicadas de un grupo. Devuelve las
// variables que no existían en el grupo.
func RemoveVariablesFromGroup(client *azdevops.Client, group models.VariableGroup, keys []string) (missing []string, err error) {
	remaining := maps.Clone(group.Variables)
	removed := 0
	for _, key := range keys {
		if _, ok := remaining[key]; !ok {
			missing = append(missing, key)
			continue
		}
		delete(remaining, key)
		removed++
	}
	if removed == 0 {
		return missing, nil
	}
	group.Variables = remaining
	_, err = UpdateVariableGroup(client, group)
	return missing, err
}
