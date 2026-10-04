package pipeline

import (
	"azuredevops/azdevops"
	"azuredevops/azdevops/git"
	"azuredevops/models"
	"fmt"
	"net/url"
)

type CreateOptions struct {
	Name                string
	RepoType            string // "azureReposGit" o "gitHub"
	RepoName            string // para GitHub: "org/repo"
	Branch              string // rama por defecto (opcional)
	YAMLPath            string
	Folder              string
	ServiceConnectionID string
}

// CreatePipeline crea un pipeline YAML. Si se indica Branch, se fija como rama por defecto.
func CreatePipeline(client *azdevops.Client, opts CreateOptions) (*models.Pipeline, error) {
	repo := models.PipelineRepository{
		FullName: opts.RepoName,
		Type:     opts.RepoType,
	}
	if opts.RepoType == "azureReposGit" {
		// La API espera el ID del repositorio para Azure Repos.
		if r, err := git.GetRepository(client, opts.RepoName); err == nil {
			repo.ID = r.ID
			repo.Name = r.Name
		}
	}
	if opts.ServiceConnectionID != "" {
		repo.Connection = &models.Properties{ID: opts.ServiceConnectionID}
	}

	body := models.PipelineCreate{
		Name:   opts.Name,
		Folder: opts.Folder,
		Configuration: models.Configuration{
			Type:       "yaml",
			Path:       opts.YAMLPath,
			Repository: repo,
		},
	}

	var created models.Pipeline
	u := client.ProjectURL("pipelines", url.Values{"api-version": {"7.1-preview.1"}})
	if err := client.Do("POST", u, body, &created); err != nil {
		return nil, fmt.Errorf("error al crear el pipeline: %w", err)
	}

	if opts.Branch != "" {
		if _, err := UpdateBuildDefinition(client, created.ID, UpdateOptions{Branch: opts.Branch}); err != nil {
			return &created, fmt.Errorf("pipeline creado (ID %d) pero no se pudo fijar la rama por defecto: %w", created.ID, err)
		}
	}
	return &created, nil
}
