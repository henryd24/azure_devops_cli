package models

type VariableGroup struct {
	Id                             int                              `json:"id,omitempty"`
	Name                           string                           `json:"name"`
	Type                           string                           `json:"type,omitempty"`
	Description                    string                           `json:"description,omitempty"`
	IsShared                       bool                             `json:"isShared,omitempty"`
	Variables                      map[string]VariableVal           `json:"variables"`
	VariableGroupProjectReferences []VariableGroupProjectReferences `json:"variableGroupProjectReferences,omitempty"`
}

type VariableVal struct {
	Value    string `json:"value,omitempty"`
	IsSecret bool   `json:"isSecret,omitempty"`
}

type VariableGroupById struct {
	ID                             int                              `json:"id"`
	Name                           string                           `json:"name"`
	Type                           string                           `json:"type"`
	Project                        string                           `json:"project"`
	Variables                      map[string]VariableVal           `json:"variables"`
	VariableGroupProjectReferences []VariableGroupProjectReferences `json:"variableGroupProjectReferences,omitempty"`
}

type VariableGroupProjectReferences struct {
	Description      string           `json:"description"`
	Name             string           `json:"name"`
	ProjectReference ProjectReference `json:"projectReference"`
}

type ProjectReference struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func ConstructVariableGroupProjectReferences(project string, variableName string, description string) []VariableGroupProjectReferences {
	if description == "" {
		description = "Default description"
	}
	return []VariableGroupProjectReferences{
		{
			Description: description,
			Name:        variableName,
			ProjectReference: ProjectReference{
				ID:   "",
				Name: project,
			},
		},
	}
}
