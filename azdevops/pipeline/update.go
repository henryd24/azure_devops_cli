package pipeline

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"fmt"
	"net/url"
	"strings"
)

type UpdateOptions struct {
	NewName             string
	YAMLPath            string
	RepoName            string
	Branch              string
	ServiceConnectionID string
}

func (o UpdateOptions) IsEmpty() bool {
	return o == UpdateOptions{}
}

// UpdateBuildDefinition actualiza solo los campos indicados de un pipeline.
func UpdateBuildDefinition(client *azdevops.Client, id int, opts UpdateOptions) (*models.BuildDefinition, error) {
	def, err := GetBuildDefinitionByID(client, id)
	if err != nil {
		return nil, err
	}

	if opts.NewName != "" {
		def.Name = &opts.NewName
	}
	if opts.YAMLPath != "" {
		if def.Process == nil {
			def.Process = &models.Process{}
		}
		def.Process.YAMLFilename = &opts.YAMLPath
	}
	if opts.RepoName != "" || opts.Branch != "" || opts.ServiceConnectionID != "" {
		if def.Repository == nil {
			def.Repository = &models.BuildDefinitionRepository{}
		}
	}
	if opts.RepoName != "" {
		def.Repository.Name = &opts.RepoName
		def.Repository.ID = &opts.RepoName // Para GitHub, el ID y el Nombre son "org/repo"
	}
	if opts.Branch != "" {
		branch := opts.Branch
		if !strings.HasPrefix(branch, "refs/") {
			branch = "refs/heads/" + branch
		}
		def.Repository.DefaultBranch = &branch
	}
	if opts.ServiceConnectionID != "" {
		if def.Repository.Properties == nil {
			def.Repository.Properties = &models.RepositoryProperties{}
		}
		def.Repository.Properties.ConnectedServiceID = &opts.ServiceConnectionID
	}

	var updated models.BuildDefinition
	u := client.ProjectURL(fmt.Sprintf("build/definitions/%d", id), url.Values{"api-version": {"7.1"}})
	if err := client.Do("PUT", u, def, &updated); err != nil {
		return nil, fmt.Errorf("error al actualizar el pipeline %d: %w", id, err)
	}
	return &updated, nil
}
