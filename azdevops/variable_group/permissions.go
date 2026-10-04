package variable_group

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"fmt"
	"net/url"
	"strings"
)

// ValidRoles son los roles que se pueden asignar en un Variable Group.
var ValidRoles = []string{"Reader", "User", "Administrator"}

// NormalizeRole valida un rol sin distinguir mayúsculas y lo devuelve con el formato de la API.
func NormalizeRole(role string) (string, error) {
	for _, r := range ValidRoles {
		if strings.EqualFold(r, role) {
			return r, nil
		}
	}
	return "", fmt.Errorf("el rol '%s' no es válido. Roles válidos: %s", role, strings.Join(ValidRoles, ", "))
}

// SetPermissions asigna un rol a las identidades indicadas en un Variable Group.
func SetPermissions(client *azdevops.Client, projectID string, groupID int, identityIDs []string, role string) error {
	role, err := NormalizeRole(role)
	if err != nil {
		return err
	}
	if len(identityIDs) == 0 {
		return fmt.Errorf("no se proporcionaron identidades para asignar permisos")
	}

	assignments := make([]models.SecurityRoleAssignment, len(identityIDs))
	for i, id := range identityIDs {
		assignments[i] = models.SecurityRoleAssignment{RoleName: role, UserID: id}
	}

	path := fmt.Sprintf("securityroles/scopes/distributedtask.variablegroup/roleassignments/resources/%s$%d", projectID, groupID)
	if err := client.Do("PUT", client.OrgURL(path, url.Values{"api-version": {"7.1-preview.1"}}), assignments, nil); err != nil {
		return fmt.Errorf("error al asignar permisos: %w", err)
	}
	return nil
}
