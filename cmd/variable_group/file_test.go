package variable_group

import (
	"testing"

	"azuredevops/models"
)

func TestParseEnvFile(t *testing.T) {
	data := []byte(`
# comentario
export A=1
B = "con espacios"
C='literal $x'
D=a=b
EMPTY=
`)
	vars, err := parseEnvFile(data)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"A": "1", "B": "con espacios", "C": "literal $x", "D": "a=b", "EMPTY": ""}
	if len(vars) != len(want) {
		t.Fatalf("got %d vars, want %d: %v", len(vars), len(want), vars)
	}
	for k, v := range want {
		if vars[k].Value != v {
			t.Errorf("%s = %q, want %q", k, vars[k].Value, v)
		}
	}

	if _, err := parseEnvFile([]byte("SIN_IGUAL")); err == nil {
		t.Error("se esperaba error para línea sin '='")
	}
}

func TestParseJSONFile(t *testing.T) {
	vars, err := parseJSONFile([]byte(`{"a":"1","b":{"value":"2","isSecret":true}}`))
	if err != nil {
		t.Fatal(err)
	}
	if vars["a"].Value != "1" || vars["a"].IsSecret {
		t.Errorf("a incorrecta: %+v", vars["a"])
	}
	if vars["b"].Value != "2" || !vars["b"].IsSecret {
		t.Errorf("b incorrecta: %+v", vars["b"])
	}
}

func TestFormatEnvRoundTrip(t *testing.T) {
	in := map[string]models.VariableVal{
		"PLAIN":  {Value: "x"},
		"SPACED": {Value: "hola mundo"},
		"TOKEN":  {IsSecret: true},
	}
	out := formatEnv(in)
	parsed, err := parseEnvFile(out)
	if err != nil {
		t.Fatalf("no se pudo releer:\n%s\n%v", out, err)
	}
	if parsed["PLAIN"].Value != "x" || parsed["SPACED"].Value != "hola mundo" {
		t.Errorf("round trip incorrecto: %v", parsed)
	}
	if _, ok := parsed["TOKEN"]; ok {
		t.Error("las variables secretas no deben exportarse con valor")
	}
}
