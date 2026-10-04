package security

import (
	"fmt"
	"strings"

	"azuredevops/azdevops"
	"azuredevops/azdevops/organization"
	"azuredevops/azdevops/security"
	"azuredevops/internal/ui"
	"azuredevops/models"
)

type subject struct{ Name, Descriptor string }

// projectGroups lista los grupos de seguridad del proyecto actual.
func projectGroups(client *azdevops.Client) ([]models.GraphGroup, error) {
	var groups []models.GraphGroup
	err := ui.Spinner("Cargando grupos del proyecto...", func() error {
		scope, err := organization.GetProjectDescriptor(client)
		if err != nil {
			return err
		}
		groups, err = security.ListGroups(client, scope, "")
		return err
	})
	return groups, err
}

// pickProjectGroups permite elegir grupos del proyecto; devuelve sus nombres sin prefijo.
func pickProjectGroups(client *azdevops.Client, title string, multi bool) ([]string, error) {
	groups, err := projectGroups(client)
	if err != nil {
		return nil, err
	}
	prefix := "[" + client.Project + "]\\"
	opts := make([]ui.Option[string], len(groups))
	for i, g := range groups {
		name := strings.TrimPrefix(g.PrincipalName, prefix)
		opts[i] = ui.Option[string]{Label: name, Value: name}
	}
	if multi {
		return ui.MultiSelect(title, opts)
	}
	v, err := ui.Select(title, opts)
	return []string{v}, err
}

// resolveGroups busca los descriptores de grupos del proyecto. Avisa de los que no existen.
func resolveGroups(client *azdevops.Client, names []string) []subject {
	var out []subject
	for _, n := range names {
		g, err := security.GetGroupByPrincipalName(client, n)
		if err != nil {
			ui.Warn("%v", err)
			continue
		}
		out = append(out, subject{n, g.SubjectDescriptor})
	}
	return out
}

// resolveUsers busca los descriptores de usuarios. Avisa de los que no existen.
func resolveUsers(client *azdevops.Client, upns []string) []subject {
	var out []subject
	for _, u := range upns {
		id, err := security.GetUserByPrincipalName(client, u)
		if err != nil {
			ui.Warn("%v", err)
			continue
		}
		out = append(out, subject{u, id.SubjectDescriptor})
	}
	return out
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// promptMembers pide usuarios y grupos a agregar/quitar.
func promptMembers(client *azdevops.Client) (users, groups []string, err error) {
	u, err := ui.Input("Usuarios (emails separados por coma)", "Opcional", "", nil)
	if err != nil {
		return nil, nil, err
	}
	users = splitList(u)
	addGroups, err := ui.Confirm("¿Incluir grupos como miembros?", false)
	if err != nil {
		return nil, nil, err
	}
	if addGroups {
		if groups, err = pickProjectGroups(client, "Grupos miembro", true); err != nil {
			return nil, nil, err
		}
	}
	if len(users) == 0 && len(groups) == 0 {
		return nil, nil, fmt.Errorf("debes indicar al menos un usuario o grupo")
	}
	return users, groups, nil
}
