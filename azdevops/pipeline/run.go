package pipeline

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"bytes"

	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func RunPipeline(client *azdevops.Client, pipelineID int, wait bool, params map[string]string, vars map[string]models.BuildVariable) (*models.Build, error) {
	payloadData := models.BuildRunPayload{}

	if len(params) > 0 {
		payloadData.TemplateParameters = params
	}
	if len(vars) > 0 {
		payloadData.Variables = vars
	}

	payload, err := json.Marshal(payloadData)
	if err != nil {
		return nil, fmt.Errorf("error al serializar el payload: %w", err)
	}

	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/pipelines/%d/runs?api-version=7.1", client.Org, client.Project, pipelineID)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", client.AuthHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("la solicitud para iniciar el build falló: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error al iniciar el build: %s", string(body))
	}

	var buildRun models.Build
	if err := json.NewDecoder(resp.Body).Decode(&buildRun); err != nil {
		return nil, fmt.Errorf("error al decodificar la respuesta del build: %w", err)
	}

	fmt.Printf("✔ Pipeline iniciado con éxito (Build ID: %d)\n", buildRun.ID)
	fmt.Printf("🔗 URL de la ejecución: %s\n", buildRun.Links.Web.Href)

	if !wait {
		return &buildRun, nil
	}

	fmt.Print("Esperando a que el pipeline finalice...")
	for {
		time.Sleep(10 * time.Second)

		latestBuild, err := GetBuildByID(client, buildRun.ID)
		if err != nil {
			fmt.Println("\nError al obtener el estado del build:", err)
			break
		}

		if latestBuild.Status == "completed" {
			fmt.Println("\n✔ Pipeline finalizado.")
			fmt.Printf("Resultado: %s\n", latestBuild.Result)
			return latestBuild, nil
		}
		fmt.Print(".")
	}

	return &buildRun, nil
}

func GetBuildByID(client *azdevops.Client, buildID int) (*models.Build, error) {
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/build/builds/%d?api-version=7.1", client.Org, client.Project, buildID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", client.AuthHeader())

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var build models.Build
	if err := json.NewDecoder(resp.Body).Decode(&build); err != nil {
		return nil, fmt.Errorf("error al decodificar la definición: %w", err)
	}
	return &build, nil
}
