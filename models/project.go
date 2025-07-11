package models

import "time"

type Project struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	URL            string       `json:"url"`
	State          string       `json:"state"`
	Revision       int          `json:"revision"`
	Links          ProjectLinks `json:"_links"`
	Visibility     string       `json:"visibility"`
	DefaultTeam    Team         `json:"defaultTeam"`
	LastUpdateTime time.Time    `json:"lastUpdateTime"`
}

type ProjectLinks struct {
	Self       Link `json:"self"`
	Collection Link `json:"collection"`
	Web        Link `json:"web"`
}

type Link struct {
	Href string `json:"href"`
}

type Team struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}
