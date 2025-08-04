package models

import "time"

type Pipeline struct {
	ID       int           `json:"id"`
	Name     string        `json:"name"`
	Folder   string        `json:"folder"`
	Revision int           `json:"revision"`
	URL      string        `json:"url"`
	Links    PipelineLinks `json:"_links"`
}

type PipelineLinks struct {
	Web Link `json:"web"`
}

type PipelineCreate struct {
	Name          string        `json:"name"`
	Folder        string        `json:"folder,omitempty"`
	Configuration Configuration `json:"configuration"`
}

type Configuration struct {
	Type       string             `json:"type"`
	Path       string             `json:"path"`
	Repository PipelineRepository `json:"repository"`
}

type PipelineRepository struct {
	FullName   string      `json:"FullName"`
	Type       string      `json:"type"` // "azureReposGit" o "gitHub"
	Connection *Properties `json:"connection,omitempty"`
}

type Properties struct {
	ID string `json:"id"`
}

type RetentionLease struct {
	LeaseID    int       `json:"leaseId"`
	OwnerID    string    `json:"ownerId"`
	CreatedOn  time.Time `json:"createdOn"`
	PipelineID int       `json:"runId"`
}
