package security

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// ListGroups lista los grupos de seguridad de la organización (o solo los del
// proyecto si scopeDescriptor no está vacío). search filtra por nombre sin
// distinguir mayúsculas.
func ListGroups(client *azdevops.Client, scopeDescriptor, search string) ([]models.GraphGroup, error) {
	var all []models.GraphGroup
	continuation := ""
	for {
		q := url.Values{"api-version": {"7.1-preview.1"}}
		if scopeDescriptor != "" {
			q.Set("scopeDescriptor", scopeDescriptor)
		}
		if continuation != "" {
			q.Set("continuationToken", continuation)
		}
		var page struct {
			Value []models.GraphGroup `json:"value"`
		}
		headers, err := client.GetWithHeaders(client.VSSPSURL("graph/groups", q), &page)
		if err != nil {
			return nil, fmt.Errorf("error al listar los grupos: %w", err)
		}
		all = append(all, page.Value...)
		continuation = headers.Get("X-MS-ContinuationToken")
		if continuation == "" {
			break
		}
	}

	if search != "" {
		needle := strings.ToLower(search)
		filtered := all[:0]
		for _, g := range all {
			if strings.Contains(strings.ToLower(g.DisplayName), needle) || strings.Contains(strings.ToLower(g.PrincipalName), needle) {
				filtered = append(filtered, g)
			}
		}
		all = filtered
	}
	sort.Slice(all, func(i, j int) bool { return all[i].PrincipalName < all[j].PrincipalName })
	return all, nil
}

// GetGroupByPrincipalName busca un grupo del proyecto por nombre (sin el prefijo "[Proyecto]\").
func GetGroupByPrincipalName(client *azdevops.Client, name string) (*models.Identity, error) {
	name = strings.TrimPrefix(name, "["+client.Project+"]\\")
	identities, err := searchIdentities(client, fmt.Sprintf("[%s]\\%s", client.Project, name))
	if err != nil {
		return nil, err
	}
	if len(identities) == 0 {
		return nil, fmt.Errorf("no se encontró el grupo '%s'", name)
	}
	return &identities[0], nil
}

func searchIdentities(client *azdevops.Client, filter string) ([]models.Identity, error) {
	q := url.Values{"searchFilter": {"General"}, "filterValue": {filter}, "api-version": {"7.1-preview.1"}}
	var result struct {
		Value []models.Identity `json:"value"`
	}
	if err := client.Do("GET", client.VSSPSURL("identities", q), nil, &result); err != nil {
		return nil, fmt.Errorf("error al buscar '%s': %w", filter, err)
	}
	return result.Value, nil
}
