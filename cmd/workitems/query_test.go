package workitems

import (
	"strings"
	"testing"
)

func TestBuildWIQL(t *testing.T) {
	got := buildWIQL(listOptions{AssignedTo: "@me", Types: []string{"Bug", "Task"}, OpenOnly: true, Search: "it's"})
	for _, want := range []string{
		"[System.AssignedTo] = @me",
		"[System.WorkItemType] IN ('Bug', 'Task')",
		"[System.State] NOT IN ('Closed', 'Done', 'Removed', 'Resolved')",
		"[System.Title] CONTAINS 'it''s'",
		"ORDER BY [System.ChangedDate] DESC",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("falta %q en:\n%s", want, got)
		}
	}
	got = buildWIQL(listOptions{States: []string{"Active"}, OpenOnly: true})
	if strings.Contains(got, "NOT IN") || strings.Contains(got, "AssignedTo") {
		t.Errorf("con --state no debe filtrar abiertos ni asignado: %s", got)
	}
}

func TestHTMLToText(t *testing.T) {
	got := htmlToText("<div>Hola&nbsp;<b>mundo</b></div><div>línea 2<br/>x &amp; y</div>")
	want := "Hola mundo\nlínea 2\nx & y"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseFields(t *testing.T) {
	f, err := parseFields([]string{"Microsoft.VSTS.Common.Priority=1", "Tags=a; b"})
	if err != nil {
		t.Fatal(err)
	}
	if f["Microsoft.VSTS.Common.Priority"] != "1" || f["System.Tags"] != "a; b" {
		t.Errorf("fields = %v", f)
	}
	if _, err := parseFields([]string{"x"}); err == nil {
		t.Error("se esperaba error")
	}
}
