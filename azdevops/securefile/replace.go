package securefile

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Snapshot es la configuración de acceso de un secure file que se conserva al reemplazarlo.
type Snapshot struct {
	File        models.SecureFile                  `json:"file"`
	Permissions models.ResourcePipelinePermissions `json:"pipelinePermissions"`
	Roles       []models.ResourceRoleAssignment    `json:"roles"`
	Checks      []models.CheckConfiguration        `json:"checks"`
}

// TakeSnapshot lee pipelines autorizados, roles y checks de un archivo.
func TakeSnapshot(client *azdevops.Client, projectID string, file models.SecureFile) (*Snapshot, error) {
	perms, err := GetPipelinePermissions(client, file.ID)
	if err != nil {
		return nil, err
	}
	roles, err := GetRoles(client, projectID, file.ID)
	if err != nil {
		return nil, err
	}
	checks, err := GetChecks(client, file.ID)
	if err != nil {
		return nil, err
	}
	return &Snapshot{File: file, Permissions: *perms, Roles: roles, Checks: checks}, nil
}

// AssignedRoles devuelve los roles asignados directamente (no heredados).
func (s Snapshot) AssignedRoles() []models.ResourceRoleAssignment {
	var out []models.ResourceRoleAssignment
	for _, r := range s.Roles {
		if r.Access == "assigned" {
			out = append(out, r)
		}
	}
	return out
}

// Inherits indica si el archivo hereda permisos del proyecto. Si la herencia está
// desactivada la API no devuelve roles "inherited".
func (s Snapshot) Inherits() bool {
	for _, r := range s.Roles {
		if r.Access == "inherited" {
			return true
		}
	}
	return false
}

// ReplaceOptions controla el reemplazo.
type ReplaceOptions struct {
	// KeepOld conserva el archivo anterior (renombrado) en lugar de eliminarlo.
	KeepOld bool
	// SelfID es el ID de la identidad que ejecuta el reemplazo. Azure DevOps no permite
	// cambiar el rol propio y quien sube el archivo queda como Administrator, así que
	// su rol no se copia.
	SelfID string
	// Progress recibe una descripción de cada paso (opcional).
	Progress func(step string)
}

// ReplaceResult describe el resultado del reemplazo.
type ReplaceResult struct {
	New            models.SecureFile `json:"new"`
	OldID          string            `json:"oldId"`
	OldRenamedTo   string            `json:"oldRenamedTo,omitempty"`
	OldDeleted     bool              `json:"oldDeleted"`
	Pipelines      []int             `json:"pipelines"`
	AllPipelines   bool              `json:"allPipelines"`
	Roles          int               `json:"roles"`
	Inherits       bool              `json:"inheritsPermissions"`
	Checks         int               `json:"checks"`
	DeleteOldError string            `json:"deleteOldError,omitempty"`
	// Warnings son diferencias que no se pudieron conservar.
	Warnings []string `json:"warnings,omitempty"`
}

// Replace sustituye el contenido de un secure file conservando su nombre, sus
// propiedades, los pipelines autorizados, los roles asignados, la herencia de
// permisos y los checks/aprobaciones.
//
// Azure DevOps no permite cambiar el contenido de un secure file ni tener dos con el
// mismo nombre, así que el proceso es: tomar una foto de la configuración, renombrar
// el archivo actual, subir el nuevo con el nombre original, copiar la configuración y
// eliminar el anterior. Si algo falla antes de eliminar el anterior, se revierte
// (se borra el nuevo y el anterior recupera su nombre).
func Replace(client *azdevops.Client, projectID string, snap *Snapshot, data []byte, opts ReplaceOptions) (result *ReplaceResult, err error) {
	progress := func(format string, a ...any) {
		if opts.Progress != nil {
			opts.Progress(fmt.Sprintf(format, a...))
		}
	}
	old := snap.File
	result = &ReplaceResult{OldID: old.ID, AllPipelines: snap.Permissions.OpenToAllPipelines(), Inherits: snap.Inherits()}

	// 1. Renombrar el archivo actual para liberar el nombre.
	tmpName := fmt.Sprintf("%s.replaced-%s", old.Name, time.Now().UTC().Format("20060102T150405"))
	progress("Renombrando '%s' a '%s'", old.Name, tmpName)
	renamed := old
	renamed.Name = tmpName
	if _, err := Update(client, renamed); err != nil {
		return nil, err
	}
	result.OldRenamedTo = tmpName

	var created *models.SecureFile
	rollback := func(cause error) error {
		errs := []error{cause}
		if created != nil {
			progress("Revirtiendo: eliminando el archivo nuevo")
			if err := Delete(client, created.ID); err != nil {
				errs = append(errs, fmt.Errorf("no se pudo eliminar el archivo nuevo %s: %w", created.ID, err))
			}
		}
		progress("Revirtiendo: restaurando el nombre '%s'", old.Name)
		if _, err := Update(client, old); err != nil {
			errs = append(errs, fmt.Errorf("no se pudo restaurar el nombre del archivo original (ahora '%s'): %w", tmpName, err))
		}
		return fmt.Errorf("reemplazo revertido: %w", errors.Join(errs...))
	}

	// 2. Subir el nuevo contenido con el nombre original.
	progress("Subiendo el nuevo contenido como '%s'", old.Name)
	if created, err = Upload(client, old.Name, data, result.AllPipelines); err != nil {
		return nil, rollback(err)
	}
	result.New = *created

	// 3. Propiedades.
	if len(old.Properties) > 0 {
		progress("Copiando %d propiedad(es)", len(old.Properties))
		withProps := *created
		withProps.Properties = old.Properties
		updated, err := Update(client, withProps)
		if err != nil {
			return nil, rollback(err)
		}
		result.New = *updated
	}

	// 4. Pipelines autorizados.
	if ids := snap.Permissions.AuthorizedPipelineIDs(); len(ids) > 0 {
		progress("Autorizando %d pipeline(s)", len(ids))
		if err := SetPipelinePermissions(client, created.ID, ids, true, nil); err != nil {
			return nil, rollback(err)
		}
		result.Pipelines = ids
	}

	// 5. Roles asignados y herencia.
	var roles []models.SecurityRoleAssignment
	for _, r := range snap.AssignedRoles() {
		if opts.SelfID != "" && strings.EqualFold(r.Identity.ID, opts.SelfID) {
			if r.Role.Name != "Administrator" {
				result.Warnings = append(result.Warnings, fmt.Sprintf(
					"tu identidad tenía el rol %s; en el archivo nuevo eres Administrator (Azure DevOps no permite cambiar el rol propio)", r.Role.Name))
			}
			result.Roles++
			continue
		}
		roles = append(roles, models.SecurityRoleAssignment{RoleName: r.Role.Name, UserID: r.Identity.ID})
	}
	if len(roles) > 0 {
		progress("Copiando %d rol(es) asignado(s)", len(roles))
		if err := SetRoles(client, projectID, created.ID, roles); err != nil {
			return nil, rollback(err)
		}
		result.Roles += len(roles)
	}
	if !result.Inherits {
		progress("Desactivando la herencia de permisos")
		if err := SetInheritance(client, projectID, created.ID, false); err != nil {
			return nil, rollback(err)
		}
	}

	// 6. Checks y aprobaciones.
	for _, check := range snap.Checks {
		progress("Copiando el check '%s'", check.Type.Name)
		if err := CopyCheck(client, check, created.ID, old.Name); err != nil {
			return nil, rollback(err)
		}
		result.Checks++
	}

	// 7. Eliminar el archivo anterior (sus checks se eliminan con él).
	if opts.KeepOld {
		return result, nil
	}
	progress("Eliminando el archivo anterior")
	if err := Delete(client, old.ID); err != nil {
		// El nuevo ya está completo: no se revierte, solo se informa.
		result.DeleteOldError = err.Error()
		return result, nil
	}
	result.OldDeleted = true
	result.OldRenamedTo = ""
	return result, nil
}
