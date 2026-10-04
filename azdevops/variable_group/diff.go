package variable_group

import (
	"azuredevops/models"
	"sort"
)

const (
	DiffOnlyA     = "solo-a"
	DiffOnlyB     = "solo-b"
	DiffChanged   = "distinto"
	DiffEqual     = "igual"
	DiffSecret    = "secreta"
	secretDisplay = "********"
)

// VariableDiff describe la comparación de una variable entre dos grupos.
type VariableDiff struct {
	Variable string `json:"variable"`
	A        string `json:"a"`
	B        string `json:"b"`
	Status   string `json:"status"`
}

// Diff compara las variables de dos grupos. Los valores secretos no se pueden
// leer, así que cuando ambas son secretas el estado es "secreta" (no comparable).
func Diff(a, b map[string]models.VariableVal) []VariableDiff {
	keys := map[string]bool{}
	for k := range a {
		keys[k] = true
	}
	for k := range b {
		keys[k] = true
	}
	sorted := make([]string, 0, len(keys))
	for k := range keys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	out := make([]VariableDiff, 0, len(sorted))
	for _, k := range sorted {
		va, inA := a[k]
		vb, inB := b[k]
		d := VariableDiff{Variable: k, A: display(va, inA), B: display(vb, inB)}
		switch {
		case !inB:
			d.Status = DiffOnlyA
		case !inA:
			d.Status = DiffOnlyB
		case va.IsSecret && vb.IsSecret:
			d.Status = DiffSecret
		case va.IsSecret != vb.IsSecret || va.Value != vb.Value:
			d.Status = DiffChanged
		default:
			d.Status = DiffEqual
		}
		out = append(out, d)
	}
	return out
}

// HasDifferences indica si alguna variable difiere (las secretas no cuentan).
func HasDifferences(diffs []VariableDiff) bool {
	for _, d := range diffs {
		if d.Status != DiffEqual && d.Status != DiffSecret {
			return true
		}
	}
	return false
}

func display(v models.VariableVal, ok bool) string {
	switch {
	case !ok:
		return ""
	case v.IsSecret:
		return secretDisplay
	}
	return v.Value
}
