package variable_group

import (
	"testing"

	"azuredevops/models"
)

func TestDiff(t *testing.T) {
	a := map[string]models.VariableVal{
		"same": {Value: "1"}, "changed": {Value: "1"}, "onlyA": {Value: "x"},
		"secret": {IsSecret: true}, "toSecret": {Value: "v"},
	}
	b := map[string]models.VariableVal{
		"same": {Value: "1"}, "changed": {Value: "2"}, "onlyB": {Value: "y"},
		"secret": {IsSecret: true}, "toSecret": {IsSecret: true},
	}
	want := map[string]string{
		"same": DiffEqual, "changed": DiffChanged, "onlyA": DiffOnlyA, "onlyB": DiffOnlyB,
		"secret": DiffSecret, "toSecret": DiffChanged,
	}
	diffs := Diff(a, b)
	if len(diffs) != len(want) {
		t.Fatalf("got %d diffs: %+v", len(diffs), diffs)
	}
	for _, d := range diffs {
		if want[d.Variable] != d.Status {
			t.Errorf("%s: status %s, want %s", d.Variable, d.Status, want[d.Variable])
		}
	}
	if diffs[0].Variable != "changed" {
		t.Errorf("no está ordenado: %+v", diffs)
	}
	if !HasDifferences(diffs) {
		t.Error("HasDifferences debería ser true")
	}
	if HasDifferences(Diff(map[string]models.VariableVal{"s": {IsSecret: true}}, map[string]models.VariableVal{"s": {IsSecret: true}})) {
		t.Error("dos secretas no cuentan como diferencia")
	}
}
