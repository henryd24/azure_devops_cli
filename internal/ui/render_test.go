package ui

import (
	"bytes"
	"strings"
	"testing"
)

type item struct {
	ID   int      `json:"id"`
	Name string   `json:"name"`
	Tags []string `json:"tags,omitempty"`
}

var items = []item{{ID: 1, Name: "a", Tags: []string{"x"}}, {ID: 22, Name: "b"}}

func render(t *testing.T, format, query string) string {
	t.Helper()
	var b bytes.Buffer
	if err := Render(&b, items, format, query); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func TestRenderQueryTSV(t *testing.T) {
	got := render(t, "tsv", "[].[id, name]")
	if got != "1\ta\n22\tb\n" {
		t.Errorf("tsv = %q", got)
	}
	if got := render(t, "tsv", "[?id > `5`].name | [0]"); got != "b\n" {
		t.Errorf("tsv escalar = %q", got)
	}
}

func TestRenderYAML(t *testing.T) {
	got := render(t, "yaml", "")
	if !strings.Contains(got, "- id: 1\n  name: a\n") || !strings.Contains(got, "id: 22") {
		t.Errorf("yaml = %s", got)
	}
}

func TestRenderJSONQuery(t *testing.T) {
	got := render(t, "json", "[].name")
	if strings.ReplaceAll(strings.ReplaceAll(got, " ", ""), "\n", "") != `["a","b"]` {
		t.Errorf("json = %s", got)
	}
}

func TestRenderBadQuery(t *testing.T) {
	var b bytes.Buffer
	if err := Render(&b, items, "json", "[?"); err == nil {
		t.Error("se esperaba error de consulta")
	}
}
