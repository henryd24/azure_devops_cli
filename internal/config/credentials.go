package config

import (
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	keyringService = "azdevops-cli"

	StoreAuto    = "auto"
	StoreKeyring = "keyring"
	StoreFile    = "file"

	credentialKeyring = "keyring"
)

// SetPAT guarda el PAT del perfil según store:
//   - keyring: en el llavero del sistema (Keychain, Credential Manager, Secret Service); falla si no está disponible.
//   - file: en el archivo de configuración.
//   - auto: intenta el llavero y, si no está disponible, usa el archivo.
//
// Devuelve dónde quedó guardado ("keyring" o "file").
func SetPAT(name string, p *Profile, pat, store string) (string, error) {
	switch store {
	case StoreFile:
		_ = keyring.Delete(keyringService, name)
		p.PAT, p.Credential = pat, ""
		return StoreFile, nil
	case StoreKeyring, StoreAuto, "":
		err := keyring.Set(keyringService, name, pat)
		if err == nil {
			p.PAT, p.Credential = "", credentialKeyring
			return StoreKeyring, nil
		}
		if store == StoreKeyring {
			return "", fmt.Errorf("no se pudo guardar el PAT en el llavero del sistema: %w", err)
		}
		p.PAT, p.Credential = pat, ""
		return StoreFile, nil
	}
	return "", fmt.Errorf("almacenamiento '%s' no válido (usa auto, keyring o file)", store)
}

// GetPAT devuelve el PAT del perfil, leyéndolo del llavero si corresponde.
func (p Profile) GetPAT(name string) (string, error) {
	if p.Credential != credentialKeyring {
		return p.PAT, nil
	}
	pat, err := keyring.Get(keyringService, name)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", fmt.Errorf("el PAT del perfil '%s' no está en el llavero; ejecuta 'azdevops login --profile %s'", name, name)
	}
	if err != nil {
		return "", fmt.Errorf("no se pudo leer el PAT del llavero: %w", err)
	}
	return pat, nil
}

// InKeyring indica si el PAT del perfil está en el llavero del sistema.
func (p Profile) InKeyring() bool { return p.Credential == credentialKeyring }

// DeletePAT elimina el PAT del perfil del llavero (si existe).
func DeletePAT(name string) {
	_ = keyring.Delete(keyringService, name)
}
