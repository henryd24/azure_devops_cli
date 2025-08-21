package security

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func GetUserByPrincipalName(client *azdevops.Client, principalName string) (*models.Identity, error) {
	encodedFilterValue := url.QueryEscape(principalName)
	url := fmt.Sprintf("https://vssps.dev.azure.com/%s/_apis/identities?searchFilter=General&filterValue=%s&api-version=7.1-preview.1", client.Org, encodedFilterValue)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", client.AuthHeader())

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Value []models.Identity `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Value) == 0 {
		return nil, fmt.Errorf("no se encontró al usuario '%s'", principalName)
	}
	fmt.Printf("✔ Usuario encontrado: %s (ID: %s)\n", result.Value[0].DisplayName, result.Value[0].ID)

	return &result.Value[0], nil
}

func AddUserToGroup(client *azdevops.Client, groupDescriptor, userPrincipalName string) (bool, error) {
	userIdentity, err := GetUserByPrincipalName(client, userPrincipalName)
	if err != nil {
		return false, err
	}

	url := fmt.Sprintf("https://vssps.dev.azure.com/%s/_apis/graph/memberships/%s/%s?api-version=7.1-preview.1", client.Org, userIdentity.SubjectDescriptor, groupDescriptor)

	req, _ := http.NewRequest("PUT", url, nil)
	req.Header.Set("Authorization", client.AuthHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("error al agregar el miembro: %s", string(body))
	}

	return true, nil
}

func AddUserToGroupOptimized(client *azdevops.Client, groupDescriptor, userDescriptor string) (bool, error) {
	url := fmt.Sprintf("https://vssps.dev.azure.com/%s/_apis/graph/memberships/%s/%s?api-version=7.1-preview.1",
		client.Org, userDescriptor, groupDescriptor)

	req, _ := http.NewRequest("PUT", url, nil)
	req.Header.Set("Authorization", client.AuthHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("error al agregar el miembro (código %d): %s", resp.StatusCode, string(body))
	}

	return true, nil
}
