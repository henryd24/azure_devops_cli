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

func AddVariablesToGroup(client *azdevops.Client, variableGroup models.VariableGroup, variables map[string]models.VariableVal, description string) (bool, error) {
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/distributedtask/variablegroups/%d?api-version=7.1-preview.2", client.Org, client.Project, variableGroup.Id)
	mergedVars := make(map[string]models.VariableVal)
	maps.Copy(mergedVars, variableGroup.Variables)
	maps.Copy(mergedVars, variables)
	group := models.VariableGroupById{
		ID:                             variableGroup.Id,
		Name:                           variableGroup.Name,
		Type:                           variableGroup.Type,
		Project:                        client.Project,
		Variables:                      mergedVars,
		VariableGroupProjectReferences: models.ConstructVariableGroupProjectReferences(client.Project, variableGroup.Name, description),
	}
	payload, _ := json.Marshal(group)

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

	fmt.Printf("✔ Variables agregadas satisfactoriamente al grupo: %d\n", variableGroup.Id)
	fmt.Printf("🔗 https://dev.azure.com/%s/%s/_library?itemType=variableGroups&view=VariableGroupView&variableGroupId=%d\n", client.Org, client.Project, variableGroup.Id)
	return true, nil
}
