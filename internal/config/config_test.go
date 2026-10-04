package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/zalando/go-keyring"
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

func TestKeyringPAT(t *testing.T) {
	keyring.MockInit()
	var p Profile
	where, err := SetPAT("work", &p, "secret", StoreAuto)
	if err != nil || where != StoreKeyring {
		t.Fatalf("SetPAT = %s, %v", where, err)
	}
	if p.PAT != "" || !p.InKeyring() {
		t.Errorf("el PAT no debe quedar en el archivo: %+v", p)
	}
	got, err := p.GetPAT("work")
	if err != nil || got != "secret" {
		t.Errorf("GetPAT = %q, %v", got, err)
	}

	if where, _ := SetPAT("work", &p, "otro", StoreFile); where != StoreFile || p.PAT != "otro" || p.InKeyring() {
		t.Errorf("store=file incorrecto: %s %+v", where, p)
	}
	if _, err := (Profile{Credential: "keyring"}).GetPAT("work"); err == nil {
		t.Error("se esperaba error: el PAT se eliminó del llavero al pasar a archivo")
	}
}

func TestKeyringUnavailableFallsBack(t *testing.T) {
	keyring.MockInitWithError(errors.New("sin dbus"))
	var p Profile
	where, err := SetPAT("work", &p, "secret", StoreAuto)
	if err != nil || where != StoreFile || p.PAT != "secret" {
		t.Errorf("auto debería caer a archivo: %s %v %+v", where, err, p)
	}
	if _, err := SetPAT("work", &p, "secret", StoreKeyring); err == nil {
		t.Error("store=keyring debería fallar si no hay llavero")
	}
}
