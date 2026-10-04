package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "config.json")
	t.Setenv("AZDEVOPS_CONFIG", path)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load sin archivo: %v", err)
	}
	if len(cfg.Profiles) != 0 {
		t.Fatalf("se esperaba config vacía")
	}

	cfg.Profiles["work"] = Profile{Org: "o", Project: "p", PAT: "secret"}
	cfg.Current = "work"
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("permisos = %v, se esperaba 0600", info.Mode().Perm())
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	p, name, ok := loaded.Profile("")
	if !ok || name != "work" || p.Org != "o" || p.PAT != "secret" {
		t.Errorf("perfil cargado incorrecto: %+v %s %v", p, name, ok)
	}
}
