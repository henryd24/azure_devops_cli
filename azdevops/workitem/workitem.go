package workitem

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const patchContentType = "application/json-patch+json"

// DefaultFields son los campos que se piden al listar.
var DefaultFields = []string{
	"System.Id", "System.WorkItemType", "System.Title", "System.State",
	"System.AssignedTo", "System.ChangedDate", "System.IterationPath", "System.Tags",
}

// QuoteWIQL escapa un texto para usarlo entre comillas simples en WIQL.
func QuoteWIQL(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// Query ejecuta una consulta WIQL y devuelve los IDs encontrados.
func Query(client *azdevops.Client, wiql string, top int) ([]int, error) {
	q := url.Values{"api-version": {"7.1"}}
	if top > 0 {
		q.Set("$top", strconv.Itoa(top))
	}
	var result struct {
		WorkItems []struct {
			ID int `json:"id"`
		} `json:"workItems"`
	}
	if err := client.Do("POST", client.ProjectURL("wit/wiql", q), map[string]string{"query": wiql}, &result); err != nil {
		return nil, fmt.Errorf("error en la consulta WIQL: %w", err)
	}
	ids := make([]int, len(result.WorkItems))
	for i, w := range result.WorkItems {
		ids[i] = w.ID
	}
	return ids, nil
}

// GetBatch obtiene varios work items (en lotes de 200) conservando el orden de ids.
func GetBatch(client *azdevops.Client, ids []int, fields []string) ([]models.WorkItem, error) {
	byID := map[int]models.WorkItem{}
	for start := 0; start < len(ids); start += 200 {
		end := min(start+200, len(ids))
		body := map[string]any{"ids": ids[start:end], "fields": fields}
		var result struct {
			Value []models.WorkItem `json:"value"`
		}
		if err := client.Do("POST", client.ProjectURL("wit/workitemsbatch", url.Values{"api-version": {"7.1"}}), body, &result); err != nil {
			return nil, fmt.Errorf("error al obtener los work items: %w", err)
		}
		for _, w := range result.Value {
			byID[w.ID] = w
		}
	}
	out := make([]models.WorkItem, 0, len(ids))
	for _, id := range ids {
		if w, ok := byID[id]; ok {
			out = append(out, w)
		}
	}
	return out, nil
}

// Get obtiene un work item con todos sus campos y relaciones.
func Get(client *azdevops.Client, id int) (*models.WorkItem, error) {
	var w models.WorkItem
	u := client.ProjectURL(fmt.Sprintf("wit/workitems/%d", id), url.Values{"api-version": {"7.1"}, "$expand": {"relations"}})
	if err := client.Do("GET", u, nil, &w); err != nil {
		return nil, fmt.Errorf("error al obtener el work item %d: %w", id, err)
	}
	return &w, nil
}

type patchOp struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value"`
}

func fieldOps(fields map[string]any) []patchOp {
	ops := make([]patchOp, 0, len(fields))
	for k, v := range fields {
		ops = append(ops, patchOp{Op: "add", Path: "/fields/" + k, Value: v})
	}
	return ops
}

func parentOp(client *azdevops.Client, parentID int) patchOp {
	return patchOp{Op: "add", Path: "/relations/-", Value: map[string]any{
		"rel": "System.LinkTypes.Hierarchy-Reverse",
		"url": client.OrgURL(fmt.Sprintf("wit/workItems/%d", parentID), nil),
	}}
}

// Create crea un work item del tipo indicado. parentID (opcional) lo enlaza como hijo.
func Create(client *azdevops.Client, itemType string, fields map[string]any, parentID int) (*models.WorkItem, error) {
	ops := fieldOps(fields)
	if parentID > 0 {
		ops = append(ops, parentOp(client, parentID))
	}
	var w models.WorkItem
	u := client.ProjectURL("wit/workitems/$"+url.PathEscape(itemType), url.Values{"api-version": {"7.1"}})
	if err := client.DoContentType("POST", u, patchContentType, ops, &w); err != nil {
		return nil, fmt.Errorf("error al crear el work item: %w", err)
	}
	return &w, nil
}

// Update modifica campos de un work item. Un comentario se agrega con el campo System.History.
func Update(client *azdevops.Client, id int, fields map[string]any) (*models.WorkItem, error) {
	var w models.WorkItem
	u := client.ProjectURL(fmt.Sprintf("wit/workitems/%d", id), url.Values{"api-version": {"7.1"}})
	if err := client.DoContentType("PATCH", u, patchContentType, fieldOps(fields), &w); err != nil {
		return nil, fmt.Errorf("error al actualizar el work item %d: %w", id, err)
	}
	return &w, nil
}

// Delete envía un work item a la papelera de reciclaje.
func Delete(client *azdevops.Client, id int) error {
	u := client.ProjectURL(fmt.Sprintf("wit/workitems/%d", id), url.Values{"api-version": {"7.1"}})
	if err := client.Do("DELETE", u, nil, nil); err != nil {
		return fmt.Errorf("error al eliminar el work item %d: %w", id, err)
	}
	return nil
}

// Types lista los tipos de work item habilitados del proyecto.
func Types(client *azdevops.Client) ([]models.WorkItemType, error) {
	var result struct {
		Value []models.WorkItemType `json:"value"`
	}
	if err := client.Do("GET", client.ProjectURL("wit/workitemtypes", url.Values{"api-version": {"7.1"}}), nil, &result); err != nil {
		return nil, err
	}
	out := result.Value[:0]
	for _, t := range result.Value {
		if !t.IsDisabled {
			out = append(out, t)
		}
	}
	return out, nil
}

// States lista los estados posibles de un tipo de work item.
func States(client *azdevops.Client, itemType string) ([]string, error) {
	var result struct {
		Value []struct {
			Name string `json:"name"`
		} `json:"value"`
	}
	u := client.ProjectURL("wit/workitemtypes/"+url.PathEscape(itemType)+"/states", url.Values{"api-version": {"7.1"}})
	if err := client.Do("GET", u, nil, &result); err != nil {
		return nil, err
	}
	states := make([]string, len(result.Value))
	for i, s := range result.Value {
		states[i] = s.Name
	}
	return states, nil
}

// WebURL devuelve el enlace del portal a un work item.
func WebURL(client *azdevops.Client, id int) string {
	return client.WebURL(fmt.Sprintf("_workitems/edit/%d", id))
}
