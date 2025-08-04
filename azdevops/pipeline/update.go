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

func UpdateBuildDefinition(client *azdevops.Client, id int, newName, yamlPath, repoName, serviceConnectionID string) (*models.BuildDefinition, error) {
	currentDef, err := GetBuildDefinitionByID(client, id)
	if err != nil {
		return nil, fmt.Errorf("no se pudo obtener la definición actual para actualizar: %w", err)
	}

	if newName != "" {
		currentDef.Name = &newName
	}
	if yamlPath != "" {
		currentDef.Process.YAMLFilename = &yamlPath
	}
	if repoName != "" {
		currentDef.Repository.Name = &repoName
		currentDef.Repository.ID = &repoName // Para GitHub, el ID y el Nombre son "org/repo"
	}
	if serviceConnectionID != "" {
		if currentDef.Repository.Properties == nil {
			currentDef.Repository.Properties = &models.RepositoryProperties{}
		}
		currentDef.Repository.Properties.ConnectedServiceID = &serviceConnectionID
	}

	payload, err := json.Marshal(currentDef)
	if err != nil {
		return nil, fmt.Errorf("error al serializar el payload de actualización: %w", err)
	}

	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/build/definitions/%d?api-version=6.0", client.Org, client.Project, id)
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", client.AuthHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("la solicitud de actualización falló: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error al actualizar la definición: %s", string(body))
	}

	var updatedDef models.BuildDefinition
	if err := json.Unmarshal(body, &updatedDef); err != nil {
		return nil, fmt.Errorf("error al decodificar la respuesta de actualización: %w", err)
	}

	name := ""

	if updatedDef.Name != nil {
		name = *updatedDef.Name
	}
	fmt.Printf("✔ Pipeline '%s' (ID: %d) actualizado con éxito.\n", name, id)
	return &updatedDef, nil
}
