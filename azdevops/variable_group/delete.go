package variable_group

import (
	"azuredevops/azdevops"
	org "azuredevops/azdevops/organization"
	"azuredevops/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func DeleteVariableGroupByName(client *azdevops.Client, names []string) (bool, error) {
	if len(names) == 0 {
		return false, fmt.Errorf("debes proporcionar al menos un nombre de Variable Group")
	}
	orgInfo, err := org.GetOrganizationInfo(client)
	if err != nil {
		return false, fmt.Errorf("error al obtener la información del projecto: %w", err)
	}
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
			url := fmt.Sprintf("https://dev.azure.com/%s/_apis/distributedtask/variablegroups/%d?api-version=7.1-preview.2&projectIds=%s", client.Org, group.Id, orgInfo.ID)
			req, _ := http.NewRequest("DELETE", url, nil)
			req.Header.Add("Authorization", client.AuthHeader())

			resp, err := client.HTTP.Do(req)
			if err != nil {
				fmt.Printf("error al eliminar el grupo de variables: %s\n", err)
				return false, err
			}

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				fmt.Printf("DeleteVariableGroupByName error: %s", string(body))
				continue
			}

			resp.Body.Close()
			fmt.Printf("✔ Variable Group '%s' eliminado (ID: %d)\n", name, group.Id)
		}
	}
	return true, nil
}

func DeleteVariableInGroupById(client *azdevops.Client, names []string, variables []string) (bool, error) {
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
			for _, variable := range variables {
				if _, exists := group.Variables[variable]; !exists {
					fmt.Printf("Variable '%s' no encontrada en el Variable Group '%s'\n", variable, group.Name)
					continue
				}
				delete(group.Variables, variable)
			}
			groupPayload := models.VariableGroupById{
				ID:                             group.Id,
				Name:                           group.Name,
				Type:                           group.Type,
				Project:                        client.Project,
				Variables:                      group.Variables,
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

			fmt.Printf("✔ Variables eliminadas satisfactoriamente al grupo: %d\n", group.Id)
			fmt.Printf("🔗 https://dev.azure.com/%s/%s/_library?itemType=variableGroups&view=VariableGroupView&variableGroupId=%d\n", client.Org, client.Project, group.Id)
		}
	}
	return true, nil
}
