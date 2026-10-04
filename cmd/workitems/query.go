package workitems

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"azuredevops/azdevops/workitem"
)

type listOptions struct {
	AssignedTo string // "@me", un email/nombre o "" para cualquiera
	Types      []string
	States     []string
	Search     string
	Area       string
	OpenOnly   bool
}

// closedStates son los estados finales de los procesos estándar (Agile, Scrum, CMMI, Basic).
var closedStates = []string{"Closed", "Done", "Removed", "Resolved"}

func buildWIQL(o listOptions) string {
	conds := []string{"[System.TeamProject] = @project"}
	switch {
	case strings.EqualFold(o.AssignedTo, "@me"):
		conds = append(conds, "[System.AssignedTo] = @me")
	case o.AssignedTo != "":
		conds = append(conds, "[System.AssignedTo] = "+workitem.QuoteWIQL(o.AssignedTo))
	}
	if len(o.Types) > 0 {
		conds = append(conds, "[System.WorkItemType] IN ("+quoteAll(o.Types)+")")
	}
	if len(o.States) > 0 {
		conds = append(conds, "[System.State] IN ("+quoteAll(o.States)+")")
	} else if o.OpenOnly {
		conds = append(conds, "[System.State] NOT IN ("+quoteAll(closedStates)+")")
	}
	if o.Search != "" {
		conds = append(conds, "[System.Title] CONTAINS "+workitem.QuoteWIQL(o.Search))
	}
	if o.Area != "" {
		conds = append(conds, "[System.AreaPath] UNDER "+workitem.QuoteWIQL(o.Area))
	}
	return "SELECT [System.Id] FROM WorkItems WHERE " + strings.Join(conds, " AND ") + " ORDER BY [System.ChangedDate] DESC"
}

func quoteAll(values []string) string {
	q := make([]string, len(values))
	for i, v := range values {
		q[i] = workitem.QuoteWIQL(v)
	}
	return strings.Join(q, ", ")
}

var (
	reBreaks = regexp.MustCompile(`(?i)<br\s*/?>|</p>|</div>|</li>`)
	reTags   = regexp.MustCompile(`<[^>]+>`)
	reBlank  = regexp.MustCompile(`\n{3,}`)
)

// htmlToText convierte el HTML de descripciones y comentarios en texto plano legible.
func htmlToText(s string) string {
	s = reBreaks.ReplaceAllString(s, "\n")
	s = reTags.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	return strings.TrimSpace(reBlank.ReplaceAllString(s, "\n\n"))
}

// textToHTML convierte texto plano a HTML simple para la descripción.
func textToHTML(s string) string {
	lines := strings.Split(html.EscapeString(s), "\n")
	return strings.Join(lines, "<br>")
}

// parseFields interpreta --field Nombre=Valor. Los nombres sin punto se asumen System.*.
func parseFields(entries []string) (map[string]any, error) {
	fields := map[string]any{}
	for _, e := range entries {
		k, v, ok := strings.Cut(e, "=")
		k = strings.TrimSpace(k)
		if !ok || k == "" {
			return nil, fmt.Errorf("campo inválido '%s': usa Referencia=Valor (p. ej. Microsoft.VSTS.Common.Priority=1)", e)
		}
		if !strings.Contains(k, ".") {
			k = "System." + k
		}
		fields[k] = v
	}
	return fields, nil
}
