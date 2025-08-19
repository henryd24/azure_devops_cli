package variable_group

import (
	"azuredevops/azdevops"
	"azuredevops/azdevops/organization"
	"azuredevops/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// SetPermissionsOptimized asigna un rol a un Variable Group para un conjunto de identidades ya resueltas.
func SetPermissionsOptimized(client *azdevops.Client, variableGroup *models.VariableGroup, identityIDs []string, role string) error {
	// 1. Validar y normalizar el rol
	validRoles := map[string]bool{"reader": true, "user": true, "administrator": true}
	normalizedRole := strings.ToLower(role)
	if !validRoles[normalizedRole] {
		return fmt.Errorf("el rol '%s' no es válido. Roles válidos: Reader, User, Administrator", role)
	}
	role = strings.Title(normalizedRole)

	// 2. Construir el payload a partir de los IDs pre-buscados
	assignments := make([]models.SecurityRoleAssignment, len(identityIDs))
	for i, id := range identityIDs {
		assignments[i] = models.SecurityRoleAssignment{
			RoleName: role,
			UserID:   id,
		}
	}

	if len(assignments) == 0 {
		return fmt.Errorf("no se proporcionaron identidades para asignar permisos")
	}

	// 3. Obtener el ID del proyecto
	project, err := organization.GetOrganizationInfo(client)
	if err != nil {
		return fmt.Errorf("no se pudo obtener la información del proyecto: %w", err)
	}

	// 4. Preparar y enviar la solicitud a la API
	payload, err := json.Marshal(assignments)
	if err != nil {
		return fmt.Errorf("error al serializar el payload: %w", err)
	}

	// URL corregida con "_" en lugar de "$" como separador del recurso
	url := fmt.Sprintf("https://dev.azure.com/%s/_apis/securityroles/scopes/distributedtask.variablegroup/roleassignments/resources/%s$%d?api-version=7.1-preview.1",
		client.Org, project.ID, variableGroup.Id)

	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", client.AuthHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error al asignar permisos (código %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
