package pipeline

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func CreatePipeline(client *azdevops.Client, name, repoType, repoName, branch, yamlPath, folder, serviceConnectionID string) (*models.Pipeline, error) {
	repo := models.PipelineRepository{
		FullName: repoName,
		Type:     repoType,
	}

	if serviceConnectionID != "" {
		repo.Connection = &models.Properties{
			ID: serviceConnectionID,
		}
	}

	config := models.Configuration{
		Type:       "yaml",
		Path:       yamlPath,
		Repository: repo,
	}

	pipeline := models.PipelineCreate{
		Name:          name,
		Folder:        folder,
		Configuration: config,
	}

	payload, err := json.Marshal(pipeline)
	if err != nil {
		return nil, fmt.Errorf("error al serializar el payload: %w", err)
	}

	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/pipelines?api-version=7.1-preview.1", client.Org, client.Project)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", client.AuthHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("la solicitud falló: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error al crear el pipeline: %s", string(body))
	}

	var newPipeline models.Pipeline
	if err := json.NewDecoder(resp.Body).Decode(&newPipeline); err != nil {
		return nil, fmt.Errorf("error al decodificar la respuesta: %w", err)
	}

	fmt.Printf("✔ Pipeline '%s' creado con éxito (ID: %d)\n", newPipeline.Name, newPipeline.ID)
	fmt.Printf("🔗 %s\n", newPipeline.Links.Web.Href)

	return &newPipeline, nil
}
