package cmd

import "testing"

func TestParseVariables(t *testing.T) {
	vars, err := ParseVariables([]string{"a=1", "secret:token=x=y", " b = con espacio "})
	if err != nil {
		t.Fatal(err)
	}
	if vars["a"].Value != "1" || vars["a"].IsSecret {
		t.Errorf("a = %+v", vars["a"])
	}
	if vars["token"].Value != "x=y" || !vars["token"].IsSecret {
		t.Errorf("token = %+v", vars["token"])
	}
	if vars["b"].Value != "con espacio" {
		t.Errorf("b = %+v", vars["b"])
	}

	for _, bad := range []string{"sinigual", "=valor", "secret:=x"} {
		if _, err := ParseVariables([]string{bad}); err == nil {
			t.Errorf("se esperaba error para %q", bad)
		}
	}
}

func TestParseParams(t *testing.T) {
	p, err := ParseParams([]string{"tag=1.2.3", "empty="})
	if err != nil {
		t.Fatal(err)
	}
	if p["tag"] != "1.2.3" || p["empty"] != "" {
		t.Errorf("params = %v", p)
	}
	if _, err := ParseParams([]string{"x"}); err == nil {
		t.Error("se esperaba error")
	}
}

func TestResolveSettingsPrecedence(t *testing.T) {
	t.Setenv("AZDEVOPS_CONFIG", t.TempDir()+"/c.json")
	t.Setenv("AZURE_ORG", "env-org")
	t.Setenv("AZURE_PROJECT", "env-proj")
	t.Setenv("AZURE_PAT", "env-pat")
	t.Setenv("AZDEVOPS_PROFILE", "")

	globalFlags.project = "flag-proj"
	defer func() { globalFlags.project = "" }()

	s, _, err := ResolveSettings()
	if err != nil {
		t.Fatal(err)
	}
	if s.Org != "env-org" || s.Project != "flag-proj" || s.PAT != "env-pat" {
		t.Errorf("settings = %+v", s)
	}
}
