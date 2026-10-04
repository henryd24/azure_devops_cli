package git

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"net/url"
	"sort"
)

// ListRepositories lista los repositorios Git (Azure Repos) del proyecto.
func ListRepositories(client *azdevops.Client) ([]models.GitRepository, error) {
	var result struct {
		Value []models.GitRepository `json:"value"`
	}
	if err := client.Do("GET", client.ProjectURL("git/repositories", url.Values{"api-version": {"7.1"}}), nil, &result); err != nil {
		return nil, err
	}
	sort.Slice(result.Value, func(i, j int) bool { return result.Value[i].Name < result.Value[j].Name })
	return result.Value, nil
}

// GetRepository obtiene un repositorio por nombre o ID.
func GetRepository(client *azdevops.Client, nameOrID string) (*models.GitRepository, error) {
	var repo models.GitRepository
	u := client.ProjectURL("git/repositories/"+url.PathEscape(nameOrID), url.Values{"api-version": {"7.1"}})
	if err := client.Do("GET", u, nil, &repo); err != nil {
		return nil, err
	}
	return &repo, nil
}
