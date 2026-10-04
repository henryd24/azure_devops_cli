package environments

import (
	"fmt"
	"strconv"

	"azuredevops/azdevops"
	"azuredevops/azdevops/environment"
	"azuredevops/cmd"
	"azuredevops/internal/ui"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

var environmentsCmd = &cobra.Command{
	Use:     "environments",
	Aliases: []string{"env", "envs"},
	Short:   "Consulta environments y sus despliegues",
}

var listEnvironmentsCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "Lista los environments del proyecto",
	RunE: func(c *cobra.Command, args []string) error {
		out, err := cmd.Output("table")
		if err != nil {
			return err
		}
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		envs, err := environment.ListEnvironments(client)
		if err != nil {
			return err
		}
		if out != "table" {
			return cmd.Print(nonNil(envs))
		}
		if len(envs) == 0 {
			ui.Info("El proyecto no tiene environments.")
			return nil
		}
		rows := make([][]string, len(envs))
		for i, e := range envs {
			rows[i] = []string{strconv.Itoa(e.ID), e.Name, ui.FormatTime(e.LastModifiedOn), e.Description}
		}
		ui.Table([]string{"id", "nombre", "modificado", "descripción"}, rows)
		return nil
	},
}

var deploymentsCmd = &cobra.Command{
	Use:     "deployments",
	Short:   "Muestra los despliegues recientes de un environment",
	Example: `  azdevops environments deployments --name produccion --top 10`,
	RunE: func(c *cobra.Command, args []string) error {
		out, err := cmd.Output("table")
		if err != nil {
			return err
		}
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		env, err := resolveEnvironment(c, client)
		if err != nil {
			return err
		}
		top, _ := c.Flags().GetInt("top")
		deps, err := environment.ListDeployments(client, env.ID, top)
		if err != nil {
			return err
		}
		if out != "table" {
			return cmd.Print(nonNil(deps))
		}
		if len(deps) == 0 {
			ui.Info("'%s' no tiene despliegues.", env.Name)
			return nil
		}
		rows := make([][]string, len(deps))
		for i, d := range deps {
			rows[i] = []string{
				d.Definition.Name, d.Owner.Name, d.StageName, d.JobName, ui.Status(d.Result),
				ui.FormatTime(d.StartTime), ui.FormatDuration(d.StartTime, d.FinishTime),
			}
		}
		ui.Table([]string{"pipeline", "ejecución", "etapa", "job", "resultado", "inicio", "duración"}, rows)
		return nil
	},
}

func resolveEnvironment(c *cobra.Command, client *azdevops.Client) (*models.Environment, error) {
	envs, err := environment.ListEnvironments(client)
	if err != nil {
		return nil, err
	}
	id, _ := c.Flags().GetInt("id")
	name, _ := c.Flags().GetString("name")
	if id == 0 && name == "" {
		if !ui.Interactive() {
			return nil, fmt.Errorf("debes indicar el environment con --id o --name")
		}
		opts := make([]ui.Option[int], len(envs))
		for i, e := range envs {
			opts[i] = ui.Option[int]{Label: e.Name, Value: e.ID}
		}
		if id, err = ui.Select("Environment", opts); err != nil {
			return nil, err
		}
	}
	for i, e := range envs {
		if (id != 0 && e.ID == id) || (id == 0 && e.Name == name) {
			return &envs[i], nil
		}
	}
	return nil, fmt.Errorf("no se encontró el environment")
}

func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

func init() {
	deploymentsCmd.Flags().IntP("id", "i", 0, "ID del environment")
	deploymentsCmd.Flags().StringP("name", "n", "", "Nombre del environment")
	deploymentsCmd.Flags().Int("top", 20, "Número máximo de despliegues")
	environmentsCmd.AddCommand(listEnvironmentsCmd, deploymentsCmd)
	cmd.RootCmd.AddCommand(environmentsCmd)
}
