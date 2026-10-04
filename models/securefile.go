package models

import (
	"encoding/json"
	"time"
)

type SecureFile struct {
	ID         string            `json:"id,omitempty"`
	Name       string            `json:"name"`
	Properties map[string]string `json:"properties,omitempty"`
	CreatedBy  *IdentityRef      `json:"createdBy,omitempty"`
	CreatedOn  *time.Time        `json:"createdOn,omitempty"`
	ModifiedBy *IdentityRef      `json:"modifiedBy,omitempty"`
	ModifiedOn *time.Time        `json:"modifiedOn,omitempty"`
}

type PipelineAuthorization struct {
	ID           int          `json:"id"`
	Authorized   bool         `json:"authorized"`
	AuthorizedBy *IdentityRef `json:"authorizedBy,omitempty"`
	AuthorizedOn *time.Time   `json:"authorizedOn,omitempty"`
}

type ResourcePipelinePermissions struct {
	Resource struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	} `json:"resource"`
	Pipelines    []PipelineAuthorization `json:"pipelines"`
	AllPipelines *struct {
		Authorized   bool         `json:"authorized"`
		AuthorizedBy *IdentityRef `json:"authorizedBy,omitempty"`
		AuthorizedOn *time.Time   `json:"authorizedOn,omitempty"`
	} `json:"allPipelines,omitempty"`
}

// AuthorizedPipelineIDs devuelve los IDs de pipelines autorizados.
func (p ResourcePipelinePermissions) AuthorizedPipelineIDs() []int {
	var ids []int
	for _, pl := range p.Pipelines {
		if pl.Authorized {
			ids = append(ids, pl.ID)
		}
	}
	return ids
}

// OpenToAllPipelines indica si todos los pipelines pueden usar el recurso.
func (p ResourcePipelinePermissions) OpenToAllPipelines() bool {
	return p.AllPipelines != nil && p.AllPipelines.Authorized
}

type ResourceRoleAssignment struct {
	Identity IdentityRef `json:"identity"`
	Role     struct {
		Name        string `json:"name"`
		DisplayName string `json:"displayName,omitempty"`
	} `json:"role"`
	// Access es "assigned" (asignado directamente) o "inherited" (heredado del proyecto).
	Access string `json:"access"`
}

type CheckConfiguration struct {
	ID   int `json:"id,omitempty"`
	Type struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"type"`
	Settings json.RawMessage `json:"settings,omitempty"`
	Timeout  int             `json:"timeout,omitempty"`
	Resource struct {
		Type string `json:"type"`
		ID   string `json:"id"`
		Name string `json:"name,omitempty"`
	} `json:"resource"`
}
