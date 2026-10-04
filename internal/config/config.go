// Package config gestiona los perfiles guardados en ~/.config/azdevops/config.json.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const DefaultProfile = "default"

type Profile struct {
	Org     string `json:"org"`
	Project string `json:"project"`
	PAT     string `json:"pat,omitempty"`
	// Credential vale "keyring" cuando el PAT está en el llavero del sistema.
	Credential string `json:"credential,omitempty"`
}

type Config struct {
	Current  string             `json:"current"`
	Profiles map[string]Profile `json:"profiles"`
}

// Path devuelve la ruta del archivo de configuración. AZDEVOPS_CONFIG permite sobrescribirla.
func Path() (string, error) {
	if p := os.Getenv("AZDEVOPS_CONFIG"); p != "" {
		return p, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "azdevops", "config.json"), nil
}

// Load lee la configuración. Si el archivo no existe devuelve una configuración vacía.
func Load() (*Config, error) {
	cfg := &Config{Profiles: map[string]Profile{}}
	path, err := Path()
	if err != nil {
		return cfg, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return cfg, fmt.Errorf("el archivo de configuración %s es inválido: %w", path, err)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]Profile{}
	}
	return cfg, nil
}

// Save escribe la configuración con permisos 0600 (contiene el PAT).
func (c *Config) Save() error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Profile devuelve el perfil indicado, o el actual si name está vacío.
func (c *Config) Profile(name string) (Profile, string, bool) {
	if name == "" {
		name = c.Current
	}
	if name == "" {
		name = DefaultProfile
	}
	p, ok := c.Profiles[name]
	return p, name, ok
}

// Names devuelve los nombres de perfiles ordenados.
func (c *Config) Names() []string {
	names := make([]string, 0, len(c.Profiles))
	for n := range c.Profiles {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
