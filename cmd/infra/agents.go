package infra

import (
	"fmt"
	"strconv"

	"azuredevops/azdevops"
	"azuredevops/azdevops/agent"
	"azuredevops/cmd"
	"azuredevops/internal/ui"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

var agentsCmd = &cobra.Command{
	Use:     "agents",
	Aliases: []string{"agent", "pools"},
	Short:   "Consulta agent pools y agentes",
}

var poolsCmd = &cobra.Command{
	Use:   "pools",
	Short: "Lista los agent pools de la organización",
	RunE: func(c *cobra.Command, args []string) error {
		out, err := cmd.Output("table")
		if err != nil {
			return err
		}
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		pools, err := agent.ListPools(client)
		if err != nil {
			return err
		}
		if out != "table" {
			return cmd.Print(nonNil(pools))
		}
		rows := make([][]string, len(pools))
		for i, p := range pools {
			kind := "self-hosted"
			if p.IsHosted {
				kind = "Microsoft"
			}
			rows[i] = []string{strconv.Itoa(p.ID), p.Name, kind, strconv.Itoa(p.Size)}
		}
		ui.Table([]string{"id", "nombre", "tipo", "agentes"}, rows)
		return nil
	},
}

var listAgentsCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "Lista los agentes de un pool con su estado y trabajo actual",
	Example: `  azdevops agents list --pool Default`,
	RunE: func(c *cobra.Command, args []string) error {
		out, err := cmd.Output("table")
		if err != nil {
			return err
		}
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		pool, err := resolvePool(c, client)
		if err != nil {
			return err
		}
		agents, err := agent.ListAgents(client, pool.ID)
		if err != nil {
			return err
		}
		if out != "table" {
			return cmd.Print(nonNil(agents))
		}
		if len(agents) == 0 {
			ui.Info("El pool '%s' no tiene agentes.", pool.Name)
			return nil
		}
		rows := make([][]string, len(agents))
		online := 0
		for i, a := range agents {
			if a.Status == "online" {
				online++
			}
			enabled := "sí"
			if !a.Enabled {
				enabled = "no"
			}
			current := "-"
			if r := a.AssignedRequest; r != nil {
				current = r.Definition.Name + " / " + r.Owner.Name
			}
			last := "-"
			if r := a.LastCompletedRequest; r != nil {
				last = fmt.Sprintf("%s %s", ui.Status(r.Result), ui.FormatTime(r.FinishTime))
			}
			rows[i] = []string{strconv.Itoa(a.ID), a.Name, ui.Status(a.Status), enabled, a.Version, current, last}
		}
		ui.Table([]string{"id", "nombre", "estado", "habilitado", "versión", "trabajo actual", "último trabajo"}, rows)
		fmt.Fprintf(ui.Err, "%d de %d agentes en línea\n", online, len(agents))
		return nil
	},
}

var enableAgentCmd = &cobra.Command{
	Use:   "enable",
	Short: "Habilita un agente para que reciba trabajos",
	RunE:  func(c *cobra.Command, args []string) error { return setAgentEnabled(c, true) },
}

var disableAgentCmd = &cobra.Command{
	Use:   "disable",
	Short: "Deshabilita un agente (deja de recibir trabajos, p. ej. para mantenimiento)",
	RunE:  func(c *cobra.Command, args []string) error { return setAgentEnabled(c, false) },
}

func setAgentEnabled(c *cobra.Command, enabled bool) error {
	client, err := cmd.NewClient()
	if err != nil {
		return err
	}
	pool, err := resolvePool(c, client)
	if err != nil {
		return err
	}
	agents, err := agent.ListAgents(client, pool.ID)
	if err != nil {
		return err
	}
	name, _ := c.Flags().GetString("agent")
	var target *models.Agent
	if name == "" {
		if !ui.Interactive() {
			return ui.MissingError("agent")
		}
		opts := make([]ui.Option[int], len(agents))
		for i, a := range agents {
			opts[i] = ui.Option[int]{Label: fmt.Sprintf("%s  (%s, habilitado: %v)", a.Name, a.Status, a.Enabled), Value: i}
		}
		idx, err := ui.Select("Agente", opts)
		if err != nil {
			return err
		}
		target = &agents[idx]
	} else {
		for i, a := range agents {
			if a.Name == name || strconv.Itoa(a.ID) == name {
				target = &agents[i]
			}
		}
		if target == nil {
			return fmt.Errorf("no se encontró el agente '%s' en el pool '%s'", name, pool.Name)
		}
	}
	if !enabled {
		yes, _ := c.Flags().GetBool("yes")
		if err := cmd.Confirm(fmt.Sprintf("¿Deshabilitar el agente '%s'?", target.Name), yes); err != nil {
			return err
		}
	}
	if _, err := agent.SetAgentEnabled(client, pool.ID, target.ID, enabled); err != nil {
		return err
	}
	if enabled {
		ui.Success("Agente '%s' habilitado", target.Name)
	} else {
		ui.Success("Agente '%s' deshabilitado", target.Name)
	}
	return nil
}

func resolvePool(c *cobra.Command, client *azdevops.Client) (*models.AgentPool, error) {
	if name, _ := c.Flags().GetString("pool"); name != "" {
		return agent.FindPool(client, name)
	}
	if !ui.Interactive() {
		return nil, ui.MissingError("pool")
	}
	pools, err := agent.ListPools(client)
	if err != nil {
		return nil, err
	}
	opts := make([]ui.Option[int], len(pools))
	for i, p := range pools {
		opts[i] = ui.Option[int]{Label: fmt.Sprintf("%s  (%d agentes)", p.Name, p.Size), Value: i}
	}
	idx, err := ui.Select("Agent pool", opts)
	if err != nil {
		return nil, err
	}
	return &pools[idx], nil
}

func init() {
	for _, c := range []*cobra.Command{listAgentsCmd, enableAgentCmd, disableAgentCmd} {
		c.Flags().StringP("pool", "p", "", "Nombre o ID del agent pool")
	}
	for _, c := range []*cobra.Command{enableAgentCmd, disableAgentCmd} {
		c.Flags().StringP("agent", "a", "", "Nombre o ID del agente")
	}
	disableAgentCmd.Flags().BoolP("yes", "y", false, "Confirmar sin preguntar")
	agentsCmd.AddCommand(poolsCmd, listAgentsCmd, enableAgentCmd, disableAgentCmd)
	cmd.RootCmd.AddCommand(agentsCmd)
}
