package variable_group

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
)

func AddVariablesToGroup(client *azdevops.Client, names []string, variables map[string]models.VariableVal, description string) (bool, error) {
	for _, name := range names {
		groups, err := GetVariableGroupByName(client, name)
		if err != nil {
			return false, fmt.Errorf("error al obtener el grupo de variables: %w", err)
		}
		if len(groups) == 0 {
			fmt.Printf("No se encontró el Variable Group con nombre '%s'\n", name)
			continue
		}
		for _, group := range groups {
			url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/distributedtask/variablegroups/%d?api-version=7.1-preview.2", client.Org, client.Project, group.Id)
			if description != "" {
				group.Description = description
			}
			mergedVars := make(map[string]models.VariableVal)
			maps.Copy(mergedVars, group.Variables)
			maps.Copy(mergedVars, variables)
			groupPayload := models.VariableGroupById{
				ID:                             group.Id,
				Name:                           group.Name,
				Type:                           group.Type,
				Project:                        client.Project,
				Variables:                      mergedVars,
				VariableGroupProjectReferences: models.ConstructVariableGroupProjectReferences(client.Project, group.Name, group.Description),
			}
			payload, _ := json.Marshal(groupPayload)

			req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
			req.Header.Add("Authorization", client.AuthHeader())
			req.Header.Add("Content-Type", "application/json")

			resp, err := client.HTTP.Do(req)

			if err != nil {
				return false, err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return false, fmt.Errorf("AddVariablesToGroup error: %s", string(body))
			}

			fmt.Printf("✔ Variables agregadas satisfactoriamente al grupo: %d\n", group.Id)
			fmt.Printf("🔗 https://dev.azure.com/%s/%s/_library?itemType=variableGroups&view=VariableGroupView&variableGroupId=%d\n", client.Org, client.Project, group.Id)
		}
	}
	return true, nil
}
