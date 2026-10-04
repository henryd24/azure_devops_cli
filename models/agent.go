package models

import "time"

type AgentPool struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	IsHosted      bool   `json:"isHosted"`
	PoolType      string `json:"poolType"`
	Size          int    `json:"size"`
	AutoProvision bool   `json:"autoProvision"`
}

type AgentRequest struct {
	RequestID  int        `json:"requestId"`
	QueueTime  *time.Time `json:"queueTime,omitempty"`
	AssignTime *time.Time `json:"assignTime,omitempty"`
	FinishTime *time.Time `json:"finishTime,omitempty"`
	Result     string     `json:"result,omitempty"`
	Definition struct {
		Name string `json:"name"`
	} `json:"definition"`
	Owner struct {
		Name string `json:"name"`
	} `json:"owner"`
}

type Agent struct {
	ID                   int           `json:"id"`
	Name                 string        `json:"name"`
	Version              string        `json:"version"`
	Status               string        `json:"status"`
	Enabled              bool          `json:"enabled"`
	OSDescription        string        `json:"osDescription"`
	CreatedOn            *time.Time    `json:"createdOn,omitempty"`
	AssignedRequest      *AgentRequest `json:"assignedRequest,omitempty"`
	LastCompletedRequest *AgentRequest `json:"lastCompletedRequest,omitempty"`
}
