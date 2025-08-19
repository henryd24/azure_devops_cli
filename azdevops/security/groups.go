package security

import (
	"azuredevops/azdevops"
	"azuredevops/azdevops/organization"
	"azuredevops/models"
	"bytes"
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

func SearchGroupByName(client *azdevops.Client, groupName string) (*models.GraphGroup, error) {
	scopeDescriptor, err := organization.GetProjectDescriptor(client)
	if err != nil {
		return nil, fmt.Errorf("no se pudo obtener la información del proyecto: %w", err)
	}

	payloadData := models.SubjectQueryPayload{
		Query:           groupName,
		ScopeDescriptor: scopeDescriptor,
		SubjectKind:     []string{"Group"},
	}

	payload, err := json.Marshal(payloadData)
	if err != nil {
		return nil, fmt.Errorf("error al serializar el payload: %w", err)
	}

	url := fmt.Sprintf("https://vssps.dev.azure.com/%s/_apis/graph/subjectquery?api-version=7.1-preview.1", client.Org)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", client.AuthHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result models.SubjectQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error al decodificar la respuesta de la búsqueda: %w", err)
	}

	if len(result.Value) == 0 {
		return nil, fmt.Errorf("no se encontró ningún grupo con el nombre '%s'", groupName)
	}
	if len(result.Value) > 1 {
		return nil, fmt.Errorf("se encontró más de un grupo con el nombre '%s', por favor sé más específico", groupName)
	}

	return &result.Value[0], nil
}
