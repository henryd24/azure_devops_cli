package pipeline

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func DeletePipeline(client *azdevops.Client, pipelineID int, autoconfirmation bool) error {
	leases, err := getRetentionLeases(client, pipelineID)
	if err != nil {
		return fmt.Errorf("no se pudieron obtener las retenciones: %w", err)
	}

	if len(leases) > 0 {
		fmt.Printf("Se encontraron %d retenciones en el pipeline %d. Eliminando...\n", len(leases), pipelineID)
		leaseIds := ""
		for i, lease := range leases {
			leaseIds += strconv.Itoa(lease.LeaseID)
			if i < len(leases)-1 {
				leaseIds += ","
			}
		}

		url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/build/retention/leases?ids=%s&api-version=6.0-preview.1", client.Org, client.Project, leaseIds)
		req, _ := http.NewRequest("DELETE", url, nil)
		req.Header.Add("Authorization", client.AuthHeader())
		resp, err := client.HTTP.Do(req)
		if err != nil {
			return fmt.Errorf("error al eliminar las retenciones: %w", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusNoContent {
			return fmt.Errorf("código de estado inesperado al eliminar retenciones: %d", resp.StatusCode)
		}
		fmt.Println("✔ Retenciones eliminadas con éxito.")
	}
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/build/definitions/%d?api-version=7.1", client.Org, client.Project, pipelineID)
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Add("Authorization", client.AuthHeader())

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("error al eliminar el pipeline: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("código de estado inesperado al eliminar el pipeline: %d", resp.StatusCode)
	}

	fmt.Printf("✔ Pipeline con ID %d eliminado correctamente.\n", pipelineID)
	return nil
}

func getRetentionLeases(client *azdevops.Client, pipelineID int) ([]models.RetentionLease, error) {
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/build/retention/leases?definitionId=%d&api-version=6.0-preview.1", client.Org, client.Project, pipelineID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", client.AuthHeader())

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Value []models.RetentionLease `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Value, nil
}
