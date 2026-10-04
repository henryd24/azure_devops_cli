// Package infra contiene los comandos de conexiones de servicio y agentes.
package infra

import (
	"fmt"
	"strconv"
	"strings"

	"azuredevops/azdevops"
	"azuredevops/azdevops/organization"
	"azuredevops/azdevops/serviceendpoint"
	"azuredevops/cmd"
	"azuredevops/internal/ui"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

var serviceConnectionsCmd = &cobra.Command{
	Use:     "service-connections",
	Aliases: []string{"sc", "endpoints"},
	Short:   "Consulta y comparte conexiones de servicio",
}

var listSCCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "Lista las conexiones de servicio del proyecto",
	Example: `  azdevops service-connections list
  azdevops sc list --type azurerm -o json`,
	RunE: func(c *cobra.Command, args []string) error {
		out, err := cmd.Output("table")
		if err != nil {
			return err
		}
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		typ, _ := c.Flags().GetString("type")
		eps, err := serviceendpoint.ListEndpoints(client, typ)
		if err != nil {
			return err
		}
		if out != "table" {
			return cmd.Print(nonNil(eps))
		}
		if len(eps) == 0 {
			ui.Info("No hay conexiones de servicio.")
			return nil
		}
		rows := make([][]string, len(eps))
		for i, e := range eps {
			ready := "sí"
			if !e.IsReady {
				ready = ui.Status("failed")
			}
			scheme := ""
			if e.Authorization != nil {
				scheme = e.Authorization.Scheme
			}
			rows[i] = []string{e.Name, e.Type, scheme, ready, strconv.Itoa(len(e.ServiceEndpointProjectReferences)), e.ID}
		}
		ui.Table([]string{"nombre", "tipo", "autenticación", "lista", "proyectos", "id"}, rows)
		return nil
	},
}

var getSCCmd = &cobra.Command{
	Use:   "get",
	Short: "Muestra el detalle de una conexión de servicio",
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		ep, err := resolveEndpoint(c, client)
		if err != nil {
			return err
		}
		return cmd.Print(ep)
	},
}

var historySCCmd = &cobra.Command{
	Use:     "history",
	Short:   "Muestra qué pipelines usaron una conexión de servicio",
	Example: `  azdevops sc history --name DockerHub --top 10`,
	RunE: func(c *cobra.Command, args []string) error {
		out, err := cmd.Output("table")
		if err != nil {
			return err
		}
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		ep, err := resolveEndpoint(c, client)
		if err != nil {
			return err
		}
		top, _ := c.Flags().GetInt("top")
		history, err := serviceendpoint.ExecutionHistory(client, ep.ID, top)
		if err != nil {
			return err
		}
		if out != "table" {
			return cmd.Print(nonNil(history))
		}
		if len(history) == 0 {
			ui.Info("'%s' no se ha usado en ninguna ejecución.", ep.Name)
			return nil
		}
		rows := make([][]string, len(history))
		for i, h := range history {
			rows[i] = []string{h.Definition.Name, h.Owner.Name, h.PlanType, ui.Status(h.Result), ui.FormatTime(h.StartTime), ui.FormatDuration(h.StartTime, h.FinishTime)}
		}
		ui.Table([]string{"pipeline", "ejecución", "tipo", "resultado", "inicio", "duración"}, rows)
		return nil
	},
}

var shareSCCmd = &cobra.Command{
	Use:     "share",
	Short:   "Comparte una conexión de servicio con otros proyectos",
	Example: `  azdevops sc share --name DockerHub --with-project OtroProyecto --with-project Tercero`,
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		ep, err := resolveEndpoint(c, client)
		if err != nil {
			return err
		}
		already := map[string]bool{}
		for _, r := range ep.ServiceEndpointProjectReferences {
			already[strings.ToLower(r.ProjectReference.Name)] = true
		}

		projects, err := organization.ListProjects(client)
		if err != nil {
			return err
		}
		targets, err := cmd.ResolveSliceFlag(c, "with-project", func() ([]string, error) {
			var opts []ui.Option[string]
			for _, p := range projects {
				if !already[strings.ToLower(p.Name)] {
					opts = append(opts, ui.Option[string]{Label: p.Name, Value: p.Name})
				}
			}
			return ui.MultiSelect("Proyectos con los que compartir '"+ep.Name+"'", opts)
		})
		if err != nil {
			return err
		}

		var refs []models.ServiceEndpointProjectReference
		for _, t := range targets {
			if already[strings.ToLower(t)] {
				ui.Info("'%s' ya está compartida con %s", ep.Name, t)
				continue
			}
			var found *models.Project
			for i := range projects {
				if strings.EqualFold(projects[i].Name, t) {
					found = &projects[i]
				}
			}
			if found == nil {
				return fmt.Errorf("no existe el proyecto '%s'", t)
			}
			refs = append(refs, models.ServiceEndpointProjectReference{
				Name: ep.Name, Description: ep.Description,
				ProjectReference: models.ProjectReference{ID: found.ID, Name: found.Name},
			})
		}
		if len(refs) == 0 {
			return nil
		}
		yes, _ := c.Flags().GetBool("yes")
		if err := cmd.Confirm(fmt.Sprintf("¿Compartir '%s' con %d proyecto(s)? Podrán usarla sus pipelines", ep.Name, len(refs)), yes); err != nil {
			return err
		}
		if err := serviceendpoint.ShareEndpoint(client, ep.ID, refs); err != nil {
			return err
		}
		for _, r := range refs {
			ui.Success("'%s' compartida con %s", ep.Name, r.ProjectReference.Name)
		}
		return nil
	},
}

func resolveEndpoint(c *cobra.Command, client *azdevops.Client) (*models.ServiceEndpoint, error) {
	if id, _ := c.Flags().GetString("id"); id != "" {
		return serviceendpoint.GetEndpoint(client, id)
	}
	if name, _ := c.Flags().GetString("name"); name != "" {
		return serviceendpoint.GetEndpointByName(client, name)
	}
	if !ui.Interactive() {
		return nil, fmt.Errorf("debes indicar la conexión con --name o --id")
	}
	eps, err := serviceendpoint.ListEndpoints(client, "")
	if err != nil {
		return nil, err
	}
	opts := make([]ui.Option[int], len(eps))
	for i, e := range eps {
		opts[i] = ui.Option[int]{Label: fmt.Sprintf("%s  (%s)", e.Name, e.Type), Value: i}
	}
	idx, err := ui.Select("Conexión de servicio", opts)
	if err != nil {
		return nil, err
	}
	return &eps[idx], nil
}

func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

func init() {
	listSCCmd.Flags().StringP("type", "t", "", "Filtrar por tipo (azurerm, github, dockerregistry, …)")
	for _, c := range []*cobra.Command{getSCCmd, historySCCmd, shareSCCmd} {
		c.Flags().StringP("name", "n", "", "Nombre de la conexión")
		c.Flags().String("id", "", "ID de la conexión")
	}
	historySCCmd.Flags().Int("top", 20, "Número máximo de ejecuciones")
	shareSCCmd.Flags().StringSlice("with-project", nil, "Proyecto con el que compartir (se puede repetir)")
	shareSCCmd.Flags().BoolP("yes", "y", false, "Confirmar sin preguntar")
	serviceConnectionsCmd.AddCommand(listSCCmd, getSCCmd, historySCCmd, shareSCCmd)
	cmd.RootCmd.AddCommand(serviceConnectionsCmd)
}
