package security

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"encoding/json"
	"fmt"
	"net/http"
)

func ListGroups(client *azdevops.Client) ([]models.GraphGroup, error) {
	var allGroups []models.GraphGroup
	var continuationToken string

	for {
		url := fmt.Sprintf("https://vssps.dev.azure.com/%s/_apis/graph/groups?api-version=7.1-preview.1", client.Org)
		if continuationToken != "" {
			url = fmt.Sprintf("https://vssps.dev.azure.com/%s/_apis/graph/groups?continuationToken=%s&api-version=7.1-preview.1", client.Org, continuationToken)
		}

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("Authorization", client.AuthHeader())

		resp, err := client.HTTP.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var result struct {
			Value []models.GraphGroup `json:"value"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("error al decodificar una página de grupos: %w", err)
		}

		allGroups = append(allGroups, result.Value...)

		token := resp.Header.Get("X-MS-ContinuationToken")
		if token == "" {
			break
		}
		continuationToken = token
	}

	return allGroups, nil
}

func GetGroupByPrincipalName(client *azdevops.Client, principalName string) (*models.Identity, error) {
	url := fmt.Sprintf("https://vssps.dev.azure.com/%s/_apis/identities?searchFilter=General&filterValue=[%s]\\%s&api-version=7.1-preview.1", client.Org, client.Project, principalName)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", client.AuthHeader())

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Value []models.Identity `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Value) == 0 {
		return nil, fmt.Errorf("no se encontró al grupo '%s'", principalName)
	}
	fmt.Printf("✔ Grupo encontrado: %s (ID: %s)\n", result.Value[0].DisplayName, result.Value[0].ID)

	return &result.Value[0], nil
}
