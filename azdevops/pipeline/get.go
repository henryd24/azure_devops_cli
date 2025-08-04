package pipeline

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"encoding/json"
	"fmt"
	"net/http"
)

func GetBuildDefinitionByName(client *azdevops.Client, name string) ([]models.BuildDefinition, error) {
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/build/definitions?name=%s&api-version=6.0", client.Org, client.Project, name)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", client.AuthHeader())

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Value []models.BuildDefinition `json:"value"`
		Count int                      `json:"count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error al decodificar la lista de definiciones: %w", err)
	}

	if result.Count > 0 {
		return result.Value, nil
	}

	return nil, nil
}

func GetBuildDefinitionByID(client *azdevops.Client, id int) (*models.BuildDefinition, error) {
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/build/definitions/%d?api-version=6.0", client.Org, client.Project, id)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", client.AuthHeader())

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var definition models.BuildDefinition
	if err := json.NewDecoder(resp.Body).Decode(&definition); err != nil {
		return nil, fmt.Errorf("error al decodificar la definición: %w", err)
	}
	return &definition, nil
}
