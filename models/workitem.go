package models

import (
	"fmt"
	"strings"
)

type WorkItem struct {
	ID        int                `json:"id"`
	Rev       int                `json:"rev,omitempty"`
	Fields    map[string]any     `json:"fields"`
	Relations []WorkItemRelation `json:"relations,omitempty"`
	URL       string             `json:"url,omitempty"`
}

type WorkItemRelation struct {
	Rel        string         `json:"rel"`
	URL        string         `json:"url"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

type WorkItemType struct {
	Name          string `json:"name"`
	ReferenceName string `json:"referenceName"`
	Description   string `json:"description"`
	IsDisabled    bool   `json:"isDisabled"`
}

// Field devuelve un campo como texto. Para identidades devuelve el nombre visible.
func (w WorkItem) Field(name string) string {
	v, ok := w.Fields[name]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case map[string]any:
		if dn, ok := t["displayName"].(string); ok {
			return dn
		}
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprint(int64(t))
		}
	}
	return strings.TrimSpace(fmt.Sprint(v))
}
