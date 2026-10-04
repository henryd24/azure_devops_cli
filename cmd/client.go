package cmd

import (
	"fmt"
	"os"
	"strings"

	"azuredevops/azdevops"
	"azuredevops/internal/config"
	"azuredevops/internal/ui"
)

var cachedClient *azdevops.Client

// ResolveSettings combina flags, variables de entorno y el perfil guardado.
// Precedencia: flags > (--profile explícito > env) > perfil actual.
func ResolveSettings() (config.Profile, string, error) {
	cfg, err := config.Load()
	if err != nil {
		return config.Profile{}, "", err
	}
	profileName := globalFlags.profile
	if profileName == "" {
		profileName = os.Getenv("AZDEVOPS_PROFILE")
	}
	profile, name, found := cfg.Profile(profileName)
	if profileName != "" && !found {
		return config.Profile{}, "", fmt.Errorf("el perfil '%s' no existe (perfiles: %s)", profileName, strings.Join(cfg.Names(), ", "))
	}

	pick := func(flag, env, prof string) string {
		if flag != "" {
			return flag
		}
		if profileName != "" && prof != "" {
			return prof
		}
		if v := os.Getenv(env); v != "" {
			return v
		}
		return prof
	}
	resolved := config.Profile{
		Org:     pick(globalFlags.org, "AZURE_ORG", profile.Org),
		Project: pick(globalFlags.project, "AZURE_PROJECT", profile.Project),
		PAT:     pick(globalFlags.pat, "AZURE_PAT", profile.PAT),
	}
	return resolved, name, nil
}

// NewClient construye el cliente de Azure DevOps. Si falta configuración y la
// sesión es interactiva, ofrece ejecutar el asistente de login.
func NewClient() (*azdevops.Client, error) {
	if cachedClient != nil {
		return cachedClient, nil
	}
	settings, _, err := ResolveSettings()
	if err != nil {
		return nil, err
	}
	if settings.Org == "" || settings.Project == "" || settings.PAT == "" {
		if !ui.Interactive() {
			return nil, fmt.Errorf("falta configuración: define AZURE_ORG, AZURE_PROJECT y AZURE_PAT, usa --org/--project/--pat o ejecuta 'azdevops login'")
		}
		ui.Warn("No hay credenciales configuradas. Vamos a configurarlas.")
		if settings, err = runLoginWizard(config.DefaultProfile, settings); err != nil {
			return nil, err
		}
	}
	client := azdevops.NewClient(settings.Org, settings.Project, settings.PAT)
	client.Debug = globalFlags.debug || os.Getenv("AZDEVOPS_DEBUG") != ""
	// Permite apuntar a otro host (pruebas o Azure DevOps Server).
	if u := os.Getenv("AZDEVOPS_BASE_URL"); u != "" {
		client.BaseURL = strings.TrimRight(u, "/")
	}
	if u := os.Getenv("AZDEVOPS_VSSPS_URL"); u != "" {
		client.VSSPSBaseURL = strings.TrimRight(u, "/")
	}
	cachedClient = client
	return client, nil
}

// ResetClient descarta el cliente en caché (p. ej. al cambiar de perfil en el modo interactivo).
func ResetClient() { cachedClient = nil }
