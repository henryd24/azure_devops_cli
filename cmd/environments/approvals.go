package environments

import (
	"fmt"
	"strings"

	"azuredevops/azdevops"
	"azuredevops/azdevops/environment"
	"azuredevops/cmd"
	"azuredevops/internal/ui"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

var approvalsCmd = &cobra.Command{
	Use:     "approvals",
	Aliases: []string{"approval"},
	Short:   "Lista, aprueba o rechaza aprobaciones de despliegue pendientes",
}

var listApprovalsCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "Lista las aprobaciones (pendientes por defecto)",
	Example: `  azdevops approvals list
  azdevops approvals list --state all -o json`,
	RunE: func(c *cobra.Command, args []string) error {
		out, err := cmd.Output("table")
		if err != nil {
			return err
		}
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		state, _ := c.Flags().GetString("state")
		if state == "all" {
			state = ""
		}
		approvals, err := environment.ListApprovals(client, state)
		if err != nil {
			return err
		}
		if out != "table" {
			return cmd.Print(nonNil(approvals))
		}
		if len(approvals) == 0 {
			if state == "pending" {
				ui.Info("No hay aprobaciones pendientes.")
			} else {
				ui.Info("No hay aprobaciones.")
			}
			return nil
		}
		rows := make([][]string, len(approvals))
		for i, a := range approvals {
			rows[i] = []string{a.ID, approvalPipeline(a), ui.Status(a.Status), approvers(a), ui.FormatTime(a.CreatedOn), a.Instructions}
		}
		ui.Table([]string{"id", "pipeline / ejecución", "estado", "aprobadores", "creada", "instrucciones"}, rows)
		return nil
	},
}

var approveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Aprueba una o más aprobaciones pendientes",
	Example: `  azdevops approvals approve --id <id> --comment "OK para producción"
  azdevops approvals approve   # elegir de la lista`,
	RunE: func(c *cobra.Command, args []string) error {
		return decide(c, "approved")
	},
}

var rejectCmd = &cobra.Command{
	Use:   "reject",
	Short: "Rechaza una o más aprobaciones pendientes",
	RunE: func(c *cobra.Command, args []string) error {
		return decide(c, "rejected")
	},
}

func decide(c *cobra.Command, status string) error {
	client, err := cmd.NewClient()
	if err != nil {
		return err
	}
	verb := map[string]string{"approved": "aprobar", "rejected": "rechazar"}[status]
	ids, err := cmd.ResolveSliceFlag(c, "id", func() ([]string, error) {
		return pickPending(client, "Aprobaciones a "+verb)
	})
	if err != nil {
		return err
	}
	comment, _ := c.Flags().GetString("comment")
	if comment == "" && ui.Interactive() {
		if comment, err = ui.Input("Comentario (opcional)", "", "", nil); err != nil {
			return err
		}
	}
	yes, _ := c.Flags().GetBool("yes")
	if status == "rejected" || ui.Interactive() {
		if err := cmd.Confirm(fmt.Sprintf("¿%s %d aprobación(es)?", strings.ToUpper(verb[:1])+verb[1:], len(ids)), yes); err != nil {
			return err
		}
	}

	updated, err := environment.UpdateApprovals(client, ids, status, comment)
	if err != nil {
		return err
	}
	for _, a := range updated {
		ui.Success("%s → %s", a.ID, ui.Status(a.Status))
	}
	if cmd.WantsData() {
		return cmd.Print(updated)
	}
	return nil
}

func pickPending(client *azdevops.Client, title string) ([]string, error) {
	var pending []models.Approval
	if err := ui.Spinner("Cargando aprobaciones pendientes...", func() (err error) {
		pending, err = environment.ListApprovals(client, "pending")
		return err
	}); err != nil {
		return nil, err
	}
	if len(pending) == 0 {
		return nil, fmt.Errorf("no hay aprobaciones pendientes")
	}
	opts := make([]ui.Option[string], len(pending))
	for i, a := range pending {
		label := approvalPipeline(a)
		if a.Instructions != "" {
			label += " · " + a.Instructions
		}
		opts[i] = ui.Option[string]{Label: label, Value: a.ID}
	}
	return ui.MultiSelect(title, opts)
}

func approvalPipeline(a models.Approval) string {
	if a.Pipeline == nil {
		return "-"
	}
	if a.Pipeline.Owner.Name != "" {
		return a.Pipeline.Name + " / " + a.Pipeline.Owner.Name
	}
	return a.Pipeline.Name
}

func approvers(a models.Approval) string {
	names := make([]string, 0, len(a.Steps))
	for _, s := range a.Steps {
		n := s.AssignedApprover.DisplayName
		if s.ActualApprover != nil && s.ActualApprover.DisplayName != "" {
			n = s.ActualApprover.DisplayName
		}
		if n == "" {
			n = "-"
		}
		if s.Status != "" && s.Status != "pending" && s.Status != "undefined" {
			n += " (" + s.Status + ")"
		}
		names = append(names, n)
	}
	return strings.Join(names, ", ")
}

func init() {
	listApprovalsCmd.Flags().String("state", "pending", "Estado: pending, approved, rejected, all…")
	for _, c := range []*cobra.Command{approveCmd, rejectCmd} {
		c.Flags().StringSlice("id", nil, "ID de la aprobación (se puede repetir)")
		c.Flags().StringP("comment", "m", "", "Comentario")
		c.Flags().BoolP("yes", "y", false, "Confirmar sin preguntar")
	}
	approvalsCmd.AddCommand(listApprovalsCmd, approveCmd, rejectCmd)
	cmd.RootCmd.AddCommand(approvalsCmd)
}
