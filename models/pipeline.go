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
	ID         string      `json:"id,omitempty"`
	Name       string      `json:"name,omitempty"`
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

type GitRepository struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	DefaultBranch string `json:"defaultBranch"`
	WebURL        string `json:"webUrl"`
	RemoteURL     string `json:"remoteUrl"`
}

type ServiceEndpoint struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Type          string       `json:"type"`
	URL           string       `json:"url"`
	IsReady       bool         `json:"isReady"`
	IsShared      bool         `json:"isShared"`
	Owner         string       `json:"owner,omitempty"`
	Description   string       `json:"description"`
	CreatedBy     *IdentityRef `json:"createdBy,omitempty"`
	Authorization *struct {
		Scheme string `json:"scheme"`
	} `json:"authorization,omitempty"`
	ServiceEndpointProjectReferences []ServiceEndpointProjectReference `json:"serviceEndpointProjectReferences,omitempty"`
}

type ServiceEndpointProjectReference struct {
	Name             string           `json:"name"`
	Description      string           `json:"description,omitempty"`
	ProjectReference ProjectReference `json:"projectReference"`
}

type ServiceEndpointExecution struct {
	ID         int        `json:"id"`
	PlanType   string     `json:"planType"`
	Result     string     `json:"result"`
	StartTime  *time.Time `json:"startTime,omitempty"`
	FinishTime *time.Time `json:"finishTime,omitempty"`
	Definition struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"definition"`
	Owner struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"owner"`
}
