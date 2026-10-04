// Package securefile gestiona los archivos seguros (Library > Secure files).
package securefile

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

const (
	apiVersion   = "7.1-preview.1"
	resourceType = "securefile"
	roleScope    = "distributedtask.securefile"
)

// ValidRoles son los roles asignables en un secure file.
var ValidRoles = []string{"Reader", "User", "Administrator"}

// List lista los secure files del proyecto.
func List(client *azdevops.Client) ([]models.SecureFile, error) {
	var result struct {
		Value []models.SecureFile `json:"value"`
	}
	if err := client.Do("GET", client.ProjectURL("distributedtask/securefiles", url.Values{"api-version": {apiVersion}}), nil, &result); err != nil {
		return nil, fmt.Errorf("error al listar los secure files: %w", err)
	}
	sort.Slice(result.Value, func(i, j int) bool {
		return strings.ToLower(result.Value[i].Name) < strings.ToLower(result.Value[j].Name)
	})
	return result.Value, nil
}

// Find busca un secure file por nombre exacto o ID.
func Find(client *azdevops.Client, nameOrID string) (*models.SecureFile, error) {
	files, err := List(client)
	if err != nil {
		return nil, err
	}
	for i, f := range files {
		if f.Name == nameOrID || f.ID == nameOrID {
			return &files[i], nil
		}
	}
	return nil, fmt.Errorf("no se encontró el secure file '%s'", nameOrID)
}

// Upload sube un archivo nuevo. authorizeAll lo deja disponible para todos los pipelines.
func Upload(client *azdevops.Client, name string, data []byte, authorizeAll bool) (*models.SecureFile, error) {
	q := url.Values{"api-version": {apiVersion}, "name": {name}}
	if authorizeAll {
		q.Set("authorizePipelines", "true")
	}
	var created models.SecureFile
	if err := client.DoContentType("POST", client.ProjectURL("distributedtask/securefiles", q), "application/octet-stream", data, &created); err != nil {
		return nil, fmt.Errorf("error al subir '%s': %w", name, err)
	}
	return &created, nil
}

// Update cambia el nombre y las propiedades de un secure file (no su contenido).
func Update(client *azdevops.Client, file models.SecureFile) (*models.SecureFile, error) {
	body := models.SecureFile{ID: file.ID, Name: file.Name, Properties: file.Properties}
	var updated models.SecureFile
	u := client.ProjectURL("distributedtask/securefiles/"+url.PathEscape(file.ID), url.Values{"api-version": {apiVersion}})
	if err := client.Do("PATCH", u, body, &updated); err != nil {
		return nil, fmt.Errorf("error al actualizar '%s': %w", file.Name, err)
	}
	return &updated, nil
}

// Delete elimina un secure file.
func Delete(client *azdevops.Client, id string) error {
	u := client.ProjectURL("distributedtask/securefiles/"+url.PathEscape(id), url.Values{"api-version": {apiVersion}})
	if err := client.Do("DELETE", u, nil, nil); err != nil {
		return fmt.Errorf("error al eliminar el secure file %s: %w", id, err)
	}
	return nil
}

func permissionsURL(client *azdevops.Client, id string) string {
	return client.ProjectURL("pipelines/pipelinepermissions/"+resourceType+"/"+url.PathEscape(id), url.Values{"api-version": {apiVersion}})
}

// GetPipelinePermissions devuelve qué pipelines pueden usar el archivo.
func GetPipelinePermissions(client *azdevops.Client, id string) (*models.ResourcePipelinePermissions, error) {
	var perms models.ResourcePipelinePermissions
	if err := client.Do("GET", permissionsURL(client, id), nil, &perms); err != nil {
		return nil, fmt.Errorf("error al leer los pipelines autorizados: %w", err)
	}
	return &perms, nil
}

// SetPipelinePermissions autoriza (o revoca) pipelines concretos y, si all no es nil,
// el acceso abierto para todos los pipelines.
func SetPipelinePermissions(client *azdevops.Client, id string, pipelineIDs []int, authorized bool, all *bool) error {
	type pipelineRef struct {
		ID         int  `json:"id"`
		Authorized bool `json:"authorized"`
	}
	body := map[string]any{}
	refs := make([]pipelineRef, len(pipelineIDs))
	for i, pid := range pipelineIDs {
		refs[i] = pipelineRef{ID: pid, Authorized: authorized}
	}
	body["pipelines"] = refs
	if all != nil {
		body["allPipelines"] = map[string]bool{"authorized": *all}
	}
	if err := client.Do("PATCH", permissionsURL(client, id), body, nil); err != nil {
		return fmt.Errorf("error al actualizar los pipelines autorizados: %w", err)
	}
	return nil
}

func roleResource(projectID, fileID string) string {
	return projectID + "$" + fileID
}

func rolesURL(client *azdevops.Client, projectID, fileID string, q url.Values) string {
	if q == nil {
		q = url.Values{}
	}
	q.Set("api-version", apiVersion)
	return client.OrgURL("securityroles/scopes/"+roleScope+"/roleassignments/resources/"+roleResource(projectID, fileID), q)
}

// GetRoles devuelve los roles (asignados y heredados) del archivo.
func GetRoles(client *azdevops.Client, projectID, fileID string) ([]models.ResourceRoleAssignment, error) {
	var result struct {
		Value []models.ResourceRoleAssignment `json:"value"`
	}
	if err := client.Do("GET", rolesURL(client, projectID, fileID, nil), nil, &result); err != nil {
		return nil, fmt.Errorf("error al leer los roles: %w", err)
	}
	return result.Value, nil
}

// SetRoles asigna roles a identidades (IDs de usuario o grupo).
func SetRoles(client *azdevops.Client, projectID, fileID string, assignments []models.SecurityRoleAssignment) error {
	if len(assignments) == 0 {
		return nil
	}
	if err := client.Do("PUT", rolesURL(client, projectID, fileID, nil), assignments, nil); err != nil {
		return fmt.Errorf("error al asignar roles: %w", err)
	}
	return nil
}

// SetInheritance activa o desactiva la herencia de permisos del proyecto.
func SetInheritance(client *azdevops.Client, projectID, fileID string, inherit bool) error {
	u := rolesURL(client, projectID, fileID, url.Values{"inheritPermissions": {strconv.FormatBool(inherit)}})
	if err := client.Do("PATCH", u, nil, nil); err != nil {
		return fmt.Errorf("error al cambiar la herencia de permisos: %w", err)
	}
	return nil
}

func checksURL(client *azdevops.Client, q url.Values) string {
	q.Set("api-version", apiVersion)
	return client.ProjectURL("pipelines/checks/configurations", q)
}

// GetChecks devuelve las aprobaciones y checks configurados sobre el archivo.
func GetChecks(client *azdevops.Client, fileID string) ([]models.CheckConfiguration, error) {
	var result struct {
		Value []models.CheckConfiguration `json:"value"`
	}
	q := url.Values{"resourceType": {resourceType}, "resourceId": {fileID}, "$expand": {"settings"}}
	if err := client.Do("GET", checksURL(client, q), nil, &result); err != nil {
		return nil, fmt.Errorf("error al leer los checks: %w", err)
	}
	return result.Value, nil
}

// CopyCheck crea en fileID un check igual a cfg.
func CopyCheck(client *azdevops.Client, cfg models.CheckConfiguration, fileID, fileName string) error {
	var body models.CheckConfiguration
	body.Type = cfg.Type
	body.Settings = cfg.Settings
	body.Timeout = cfg.Timeout
	body.Resource.Type = resourceType
	body.Resource.ID = fileID
	body.Resource.Name = fileName
	if err := client.Do("POST", checksURL(client, url.Values{}), body, nil); err != nil {
		return fmt.Errorf("error al copiar el check '%s': %w", cfg.Type.Name, err)
	}
	return nil
}

// DeleteCheck elimina un check.
func DeleteCheck(client *azdevops.Client, id int) error {
	return client.Do("DELETE", client.ProjectURL(fmt.Sprintf("pipelines/checks/configurations/%d", id), url.Values{"api-version": {apiVersion}}), nil, nil)
}
