package variable_group

import (
	"fmt"
	"strings"

	"azuredevops/azdevops"
	vg "azuredevops/azdevops/variable_group"
	"azuredevops/cmd"
	"azuredevops/internal/ui"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

// findExact busca un Variable Group por su nombre exacto.
func findExact(client *azdevops.Client, name string) (*models.VariableGroup, error) {
	groups, err := vg.GetVariableGroupByName(client, name)
	if err != nil {
		return nil, err
	}
	for i := range groups {
		if groups[i].Name == name {
			return &groups[i], nil
		}
	}
	return nil, nil
}

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copia un Variable Group (también a otro proyecto)",
	Long: `Copia las variables y la descripción de un Variable Group a uno nuevo o existente.

La API no devuelve el valor de las variables secretas: en una terminal se te pedirá
cada valor (puedes dejarlo vacío) y en scripts puedes pasarlos con --secret-value.
Si el destino ya existe, usa --merge para fusionar: las variables comunes se
sobrescriben y las secretas sin valor conservan el que ya tenían en el destino.`,
	Example: `  azdevops variables copy --from app-dev --to app-qa
  azdevops variables copy --from app-dev --to app-dev --to-project OtroProyecto
  azdevops variables copy --from app-dev --to app-prod --merge --secret-value DB_PASS=xxx`,
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		from, err := cmd.ResolveStringFlag(c, "from", func() (string, error) {
			return cmd.PickVariableGroupName(client, "Variable Group origen")
		})
		if err != nil {
			return err
		}
		toProject, _ := c.Flags().GetString("to-project")
		to, err := cmd.ResolveStringFlag(c, "to", func() (string, error) {
			def := from
			if toProject == "" {
				def = from + "-copia"
			}
			return ui.Input("Nombre del grupo destino", "", def, ui.Required)
		})
		if err != nil {
			return err
		}
		target := client
		if toProject != "" {
			target = client.WithProject(toProject)
		}
		if toProject == "" && to == from {
			return fmt.Errorf("el origen y el destino son el mismo grupo; usa otro --to o --to-project")
		}

		source, err := findExact(client, from)
		if err != nil {
			return err
		}
		if source == nil {
			return fmt.Errorf("no se encontró el Variable Group '%s'", from)
		}

		secretValues, _ := c.Flags().GetStringSlice("secret-value")
		provided, err := cmd.ParseParams(secretValues)
		if err != nil {
			return err
		}
		vars := make(map[string]models.VariableVal, len(source.Variables))
		var missingSecrets []string
		for k, v := range source.Variables {
			if v.IsSecret {
				v.Value = ""
				if val, ok := provided[k]; ok {
					v.Value = val
				} else if ui.Interactive() {
					if v.Value, err = ui.Input("Valor de la variable secreta "+k, "Vacío para copiarla sin valor", "", nil); err != nil {
						return err
					}
				}
				if v.Value == "" {
					missingSecrets = append(missingSecrets, k)
				}
			}
			vars[k] = v
		}

		existing, err := findExact(target, to)
		if err != nil {
			return err
		}
		dest := to
		if toProject != "" {
			dest = toProject + "/" + to
		}
		var result *models.VariableGroup
		if existing == nil {
			if result, err = vg.CreateVariableGroup(target, to, vars, source.Description); err != nil {
				return err
			}
			ui.Success("'%s' copiado a '%s' (ID %d, %d variables)", from, dest, result.Id, len(vars))
		} else {
			merge, _ := c.Flags().GetBool("merge")
			if !merge {
				if !ui.Interactive() {
					return fmt.Errorf("el grupo destino '%s' ya existe; usa --merge para fusionar", dest)
				}
				if err := cmd.Confirm(fmt.Sprintf("'%s' ya existe. ¿Fusionar las variables (las comunes se sobrescriben)?", dest), false); err != nil {
					return err
				}
			}
			if result, err = vg.AddVariablesToGroup(target, *existing, vars, ""); err != nil {
				return err
			}
			ui.Success("'%s' fusionado en '%s' (ID %d)", from, dest, result.Id)
		}
		if len(missingSecrets) > 0 {
			if existing == nil {
				ui.Warn("Variables secretas copiadas sin valor: %s", strings.Join(missingSecrets, ", "))
			} else {
				ui.Warn("Variables secretas sin valor nuevo (conservan el del destino si existían): %s", strings.Join(missingSecrets, ", "))
			}
		}
		ui.Link(target.VariableGroupWebURL(result.Id))
		if cmd.WantsData() {
			return cmd.Print(result)
		}
		return nil
	},
}

var diffCmd = &cobra.Command{
	Use:   "diff [grupo-a] [grupo-b]",
	Short: "Compara las variables de dos Variable Groups",
	Long: `Compara dos Variable Groups (por ejemplo dev contra prod), incluso de proyectos
distintos con --project-b. Los valores secretos no se pueden leer, así que solo se
compara si existen y si son secretas en ambos lados.`,
	Example: `  azdevops variables diff app-dev app-prod
  azdevops variables diff app-prod app-prod --project-b OtroProyecto --all
  azdevops variables diff app-dev app-prod --exit-code   # código 1 si hay diferencias`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(c *cobra.Command, args []string) error {
		out, err := cmd.Output("table")
		if err != nil {
			return err
		}
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		projectB, _ := c.Flags().GetString("project-b")
		clientB := client
		if projectB != "" {
			clientB = client.WithProject(projectB)
		}
		names := append([]string{}, args...)
		if len(names) < 2 {
			if !ui.Interactive() {
				return fmt.Errorf("indica los dos grupos a comparar")
			}
			if len(names) == 0 {
				n, err := cmd.PickVariableGroupName(client, "Grupo A")
				if err != nil {
					return err
				}
				names = append(names, n)
			}
			n, err := cmd.PickVariableGroupName(clientB, "Grupo B")
			if err != nil {
				return err
			}
			names = append(names, n)
		}

		a, err := findExact(client, names[0])
		if err != nil {
			return err
		}
		b, err := findExact(clientB, names[1])
		if err != nil {
			return err
		}
		if a == nil || b == nil {
			missing := names[0]
			if a != nil {
				missing = names[1]
			}
			return fmt.Errorf("no se encontró el Variable Group '%s'", missing)
		}

		diffs := vg.Diff(a.Variables, b.Variables)
		showAll, _ := c.Flags().GetBool("all")
		if !showAll {
			filtered := diffs[:0:0]
			for _, d := range diffs {
				if d.Status != vg.DiffEqual {
					filtered = append(filtered, d)
				}
			}
			diffs = filtered
		}

		if out != "table" {
			if err := cmd.Print(diffs); err != nil {
				return err
			}
		} else if len(diffs) == 0 {
			ui.Success("'%s' y '%s' tienen las mismas variables y valores", names[0], names[1])
		} else {
			labels := map[string]string{
				vg.DiffOnlyA: "solo en A", vg.DiffOnlyB: "solo en B", vg.DiffChanged: "distinto",
				vg.DiffEqual: "igual", vg.DiffSecret: "secreta (no comparable)",
			}
			rows := make([][]string, len(diffs))
			for i, d := range diffs {
				rows[i] = []string{d.Variable, truncate(d.A, 40), truncate(d.B, 40), labels[d.Status]}
			}
			fmt.Fprintf(ui.Err, "A = %s · B = %s\n", ui.Bold(names[0]), ui.Bold(names[1]))
			ui.Table([]string{"variable", "A", "B", "estado"}, rows)
		}

		if exit, _ := c.Flags().GetBool("exit-code"); exit && vg.HasDifferences(diffs) {
			return cmd.ExitError{Code: 1}
		}
		return nil
	},
}

func init() {
	copyCmd.Flags().String("from", "", "Variable Group origen")
	copyCmd.Flags().String("to", "", "Nombre del Variable Group destino")
	copyCmd.Flags().String("to-project", "", "Proyecto destino (por defecto el mismo)")
	copyCmd.Flags().Bool("merge", false, "Si el destino existe, fusionar las variables")
	copyCmd.Flags().StringSlice("secret-value", nil, "Valor de una variable secreta: CLAVE=VALOR (se puede repetir)")
	_ = copyCmd.RegisterFlagCompletionFunc("from", cmd.CompleteVariableGroups)

	diffCmd.Flags().String("project-b", "", "Proyecto del grupo B (por defecto el mismo)")
	diffCmd.Flags().Bool("all", false, "Mostrar también las variables iguales")
	diffCmd.Flags().Bool("exit-code", false, "Terminar con código 1 si hay diferencias")
	diffCmd.ValidArgsFunction = cmd.CompleteVariableGroups

	cmd.Variables.AddCommand(copyCmd, diffCmd)
}
