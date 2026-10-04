package models

import "time"

type Environment struct {
	ID             int        `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	CreatedOn      *time.Time `json:"createdOn,omitempty"`
	LastModifiedOn *time.Time `json:"lastModifiedOn,omitempty"`
}

type EnvironmentDeployment struct {
	ID         int        `json:"id"`
	StageName  string     `json:"stageName"`
	JobName    string     `json:"jobName"`
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

type IdentityRef struct {
	ID          string `json:"id,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	UniqueName  string `json:"uniqueName,omitempty"`
}

type Approval struct {
	ID                   string         `json:"id"`
	Status               string         `json:"status"`
	Instructions         string         `json:"instructions,omitempty"`
	MinRequiredApprovers int            `json:"minRequiredApprovers"`
	CreatedOn            *time.Time     `json:"createdOn,omitempty"`
	LastModifiedOn       *time.Time     `json:"lastModifiedOn,omitempty"`
	Steps                []ApprovalStep `json:"steps,omitempty"`
	Pipeline             *struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Owner struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Links struct {
				Web struct {
					Href string `json:"href"`
				} `json:"web"`
			} `json:"_links"`
		} `json:"owner"`
	} `json:"pipeline,omitempty"`
}

type ApprovalStep struct {
	AssignedApprover IdentityRef  `json:"assignedApprover"`
	ActualApprover   *IdentityRef `json:"actualApprover,omitempty"`
	Status           string       `json:"status"`
	Comment          string       `json:"comment,omitempty"`
}
