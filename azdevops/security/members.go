package security

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"fmt"
	"net/url"
	"sort"
)

// GetUserByPrincipalName busca un usuario por email/UPN.
func GetUserByPrincipalName(client *azdevops.Client, principalName string) (*models.Identity, error) {
	identities, err := searchIdentities(client, principalName)
	if err != nil {
		return nil, err
	}
	if len(identities) == 0 {
		return nil, fmt.Errorf("no se encontró el usuario '%s'", principalName)
	}
	return &identities[0], nil
}

// AddMembership agrega subjectDescriptor (usuario o grupo) como miembro de containerDescriptor.
func AddMembership(client *azdevops.Client, containerDescriptor, subjectDescriptor string) error {
	path := fmt.Sprintf("graph/memberships/%s/%s", subjectDescriptor, containerDescriptor)
	return client.Do("PUT", client.VSSPSURL(path, url.Values{"api-version": {"7.1-preview.1"}}), nil, nil)
}

// RemoveMembership quita subjectDescriptor de containerDescriptor.
func RemoveMembership(client *azdevops.Client, containerDescriptor, subjectDescriptor string) error {
	path := fmt.Sprintf("graph/memberships/%s/%s", subjectDescriptor, containerDescriptor)
	return client.Do("DELETE", client.VSSPSURL(path, url.Values{"api-version": {"7.1-preview.1"}}), nil, nil)
}

// ListMembers devuelve los miembros directos de un grupo.
func ListMembers(client *azdevops.Client, groupDescriptor string) ([]models.GraphSubject, error) {
	q := url.Values{"direction": {"down"}, "api-version": {"7.1-preview.1"}}
	var memberships struct {
		Value []struct {
			MemberDescriptor string `json:"memberDescriptor"`
		} `json:"value"`
	}
	if err := client.Do("GET", client.VSSPSURL("graph/memberships/"+groupDescriptor, q), nil, &memberships); err != nil {
		return nil, fmt.Errorf("error al obtener los miembros: %w", err)
	}
	if len(memberships.Value) == 0 {
		return nil, nil
	}

	type lookupKey struct {
		Descriptor string `json:"descriptor"`
	}
	lookup := struct {
		LookupKeys []lookupKey `json:"lookupKeys"`
	}{}
	for _, m := range memberships.Value {
		lookup.LookupKeys = append(lookup.LookupKeys, lookupKey{m.MemberDescriptor})
	}
	var subjects struct {
		Value map[string]models.GraphSubject `json:"value"`
	}
	if err := client.Do("POST", client.VSSPSURL("graph/subjectlookup", url.Values{"api-version": {"7.1-preview.1"}}), lookup, &subjects); err != nil {
		return nil, fmt.Errorf("error al resolver los miembros: %w", err)
	}
	members := make([]models.GraphSubject, 0, len(subjects.Value))
	for _, s := range subjects.Value {
		members = append(members, s)
	}
	sort.Slice(members, func(i, j int) bool { return members[i].DisplayName < members[j].DisplayName })
	return members, nil
}
