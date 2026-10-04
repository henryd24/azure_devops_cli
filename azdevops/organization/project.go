package organization

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"fmt"
	"net/url"
	"strings"
)

// GetProject devuelve la información del proyecto configurado en el cliente.
func GetProject(client *azdevops.Client) (*models.Project, error) {
	var result models.Project
	u := client.OrgURL("projects/"+url.PathEscape(client.Project), url.Values{"api-version": {"7.1"}})
	if err := client.Do("GET", u, nil, &result); err != nil {
		return nil, fmt.Errorf("no se pudo obtener el proyecto '%s': %w", client.Project, err)
	}
	return &result, nil
}

// ListProjects devuelve los proyectos de la organización.
func ListProjects(client *azdevops.Client) ([]models.Project, error) {
	var result struct {
		Value []models.Project `json:"value"`
	}
	u := client.OrgURL("projects", url.Values{"api-version": {"7.1"}, "$top": {"500"}})
	if err := client.Do("GET", u, nil, &result); err != nil {
		return nil, err
	}
	return result.Value, nil
}

// GetProjectDescriptor devuelve el descriptor de Graph del proyecto.
func GetProjectDescriptor(client *azdevops.Client) (string, error) {
	project, err := GetProject(client)
	if err != nil {
		return "", err
	}
	var result models.GraphDescriptor
	if err := client.Do("GET", client.VSSPSURL("graph/descriptors/"+project.ID, nil), nil, &result); err != nil {
		return "", fmt.Errorf("error al obtener el descriptor del proyecto: %w", err)
	}
	return result.ID, nil
}

// User es el usuario autenticado con el PAT.
type User struct {
	ID          string
	DisplayName string
	Email       string
}

// CurrentUser devuelve el usuario dueño del PAT.
func CurrentUser(client *azdevops.Client) (*User, error) {
	var data struct {
		AuthenticatedUser struct {
			ID                  string `json:"id"`
			Descriptor          string `json:"descriptor"`
			ProviderDisplayName string `json:"providerDisplayName"`
			Properties          struct {
				Account struct {
					Value string `json:"$value"`
				} `json:"Account"`
			} `json:"properties"`
		} `json:"authenticatedUser"`
	}
	if err := client.Do("GET", client.OrgURL("connectionData", nil), nil, &data); err != nil {
		return nil, fmt.Errorf("no se pudo obtener el usuario actual: %w", err)
	}
	u := data.AuthenticatedUser
	// En organizaciones con proyectos públicos un PAT inválido no da 401: la sesión es anónima.
	if strings.HasPrefix(u.Descriptor, "System:PublicAccess") {
		return nil, &azdevops.APIError{StatusCode: 401, Message: "autenticación fallida: el PAT es inválido o expiró (acceso anónimo)"}
	}
	return &User{ID: u.ID, DisplayName: u.ProviderDisplayName, Email: u.Properties.Account.Value}, nil
}
