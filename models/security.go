package models

type GraphGroup struct {
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Descriptor  string `json:"descriptor"`
	URL         string `json:"url"`
}

type GraphUser struct {
	DisplayName   string `json:"displayName"`
	PrincipalName string `json:"principalName"`
	Descriptor    string `json:"descriptor"`
}

type Identity struct {
	ID                string `json:"id"`
	SubjectDescriptor string `json:"subjectDescriptor"`
	DisplayName       string `json:"providerDisplayName"`
	CustomDisplayName string `json:"customDisplayName"`
	Descriptor        string `json:"descriptor"`
}

type SubjectQueryPayload struct {
	Query           string   `json:"query"`
	ScopeDescriptor string   `json:"scopeDescriptor"`
	SubjectKind     []string `json:"subjectKind"`
}

type SubjectQueryResponse struct {
	Value []GraphGroup `json:"value"`
}
