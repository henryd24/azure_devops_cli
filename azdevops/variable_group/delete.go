package variable_group

import (
	"azuredevops/azdevops"
	org "azuredevops/azdevops/organization"
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
			fmt.Println(url)
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
