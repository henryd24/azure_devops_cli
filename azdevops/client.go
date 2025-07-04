package azdevops

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"time"
)

type Client struct {
	Org     string
	Project string
	PAT     string
	HTTP    *http.Client
}

func NewClient(org, project, pat string) *Client {
	return &Client{
		Org:     org,
		Project: project,
		PAT:     pat,
		HTTP:    &http.Client{Timeout: 10 * time.Second},
	}
}

func GetClientFromEnv() *Client {
	org := os.Getenv("AZURE_ORG")
	project := os.Getenv("AZURE_PROJECT")
	pat := os.Getenv("AZURE_PAT")

	if org == "" || project == "" || pat == "" {
		log.Fatal("AZURE_ORG, AZURE_PROJECT y AZURE_PAT deben estar definidos en las variables de entorno")
	}

	return NewClient(org, project, pat)
}

func (c *Client) authHeader() string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(":"+c.PAT))
}
