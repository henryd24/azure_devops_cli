package variable_group

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	vg "azuredevops/azdevops/variable_group"
	"azuredevops/cmd"
	"azuredevops/internal/ui"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Exporta las variables de un grupo a .env o JSON",
	Long: `Exporta las variables de un Variable Group. Los valores secretos no se pueden
leer desde la API, por lo que se exportan sin valor.`,
	Example: `  azdevops variables export --name MiGrupo > .env
  azdevops variables export --name MiGrupo --format json --file vars.json`,
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		name, err := cmd.ResolveStringFlag(c, "name", func() (string, error) {
			return cmd.PickVariableGroupName(client, "Variable Group a exportar")
		})
		if err != nil {
			return err
		}
		format, _ := c.Flags().GetString("format")
		file, _ := c.Flags().GetString("file")
		if format == "" {
			format = "env"
			if strings.EqualFold(filepath.Ext(file), ".json") {
				format = "json"
			}
		}

		groups, err := vg.GetVariableGroupByName(client, name)
		if err != nil {
			return err
		}
		var group *models.VariableGroup
		for i := range groups {
			if groups[i].Name == name {
				group = &groups[i]
			}
		}
		if group == nil {
			return fmt.Errorf("no se encontró el Variable Group '%s'", name)
		}

		var data []byte
		switch format {
		case "env":
			data = formatEnv(group.Variables)
		case "json":
			if data, err = json.MarshalIndent(group.Variables, "", "  "); err != nil {
				return err
			}
			data = append(data, '\n')
		default:
			return fmt.Errorf("formato '%s' no soportado (usa env o json)", format)
		}

		secrets := 0
		for _, v := range group.Variables {
			if v.IsSecret {
				secrets++
			}
		}
		if secrets > 0 {
			ui.Warn("%d variable(s) secreta(s) se exportaron sin valor", secrets)
		}
		if file == "" {
			_, err = ui.Out.Write(data)
			return err
		}
		if err := os.WriteFile(file, data, 0o600); err != nil {
			return err
		}
		ui.Success("%d variables exportadas a %s", len(group.Variables), file)
		return nil
	},
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Importa variables desde un archivo .env o JSON a un grupo",
	Example: `  azdevops variables import --name MiGrupo --file .env --secret DB_PASSWORD --secret API_KEY
  azdevops variables import --name NuevoGrupo --file vars.json --create`,
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		file, err := cmd.ResolveStringFlag(c, "file", func() (string, error) {
			return ui.Input("Archivo a importar (.env o .json)", "", ".env", ui.Required)
		})
		if err != nil {
			return err
		}
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		var vars map[string]models.VariableVal
		if strings.EqualFold(filepath.Ext(file), ".json") {
			vars, err = parseJSONFile(data)
		} else {
			vars, err = parseEnvFile(data)
		}
		if err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		if len(vars) == 0 {
			return fmt.Errorf("%s no contiene variables", file)
		}

		secretKeys, _ := c.Flags().GetStringSlice("secret")
		if len(secretKeys) == 0 && ui.Interactive() {
			opts := make([]ui.Option[string], 0, len(vars))
			for _, k := range sortedKeys(vars) {
				opts = append(opts, ui.Option[string]{Label: k, Value: k})
			}
			mark, err := ui.Confirm("¿Marcar alguna variable como secreta?", false)
			if err != nil {
				return err
			}
			if mark {
				if secretKeys, err = ui.MultiSelect("Variables secretas", opts); err != nil {
					return err
				}
			}
		}
		for _, k := range secretKeys {
			v, ok := vars[k]
			if !ok {
				return fmt.Errorf("--secret %s: la variable no está en el archivo", k)
			}
			v.IsSecret = true
			vars[k] = v
		}

		name, err := cmd.ResolveStringFlag(c, "name", func() (string, error) {
			return cmd.PickVariableGroupName(client, "Variable Group destino")
		})
		if err != nil {
			return err
		}
		groups, err := vg.GetVariableGroupByName(client, name)
		if err != nil {
			return err
		}
		var target *models.VariableGroup
		for i := range groups {
			if groups[i].Name == name {
				target = &groups[i]
			}
		}

		if target == nil {
			create, _ := c.Flags().GetBool("create")
			if !create {
				if !ui.Interactive() {
					return fmt.Errorf("el Variable Group '%s' no existe (usa --create para crearlo)", name)
				}
				if create, err = ui.Confirm(fmt.Sprintf("'%s' no existe. ¿Crearlo?", name), true); err != nil || !create {
					return ui.ErrAborted
				}
			}
			created, err := vg.CreateVariableGroup(client, name, vars, "")
			if err != nil {
				return err
			}
			ui.Success("Variable Group '%s' creado con %d variables (ID %d)", name, len(vars), created.Id)
			ui.Link(client.VariableGroupWebURL(created.Id))
			return nil
		}

		overwritten := 0
		for k := range vars {
			if _, ok := target.Variables[k]; ok {
				overwritten++
			}
		}
		if overwritten > 0 && ui.Interactive() {
			yes, _ := c.Flags().GetBool("yes")
			if err := cmd.Confirm(fmt.Sprintf("Se sobrescribirán %d variable(s) existentes en '%s'. ¿Continuar?", overwritten, name), yes); err != nil {
				return err
			}
		}
		updated, err := vg.AddVariablesToGroup(client, *target, vars, "")
		if err != nil {
			return err
		}
		ui.Success("%d variables importadas en '%s' (%d nuevas, %d sobrescritas)", len(vars), updated.Name, len(vars)-overwritten, overwritten)
		ui.Link(client.VariableGroupWebURL(updated.Id))
		return nil
	},
}

func init() {
	exportCmd.Flags().StringP("name", "n", "", "Nombre del Variable Group")
	exportCmd.Flags().StringP("format", "f", "", "Formato: env o json (por defecto según la extensión de --file, o env)")
	exportCmd.Flags().String("file", "", "Archivo de salida (por defecto stdout)")
	_ = exportCmd.RegisterFlagCompletionFunc("name", cmd.CompleteVariableGroups)

	importCmd.Flags().StringP("name", "n", "", "Nombre del Variable Group destino")
	importCmd.Flags().String("file", "", "Archivo .env o .json a importar")
	importCmd.Flags().StringSlice("secret", nil, "Claves a marcar como secretas (se puede repetir)")
	importCmd.Flags().Bool("create", false, "Crear el Variable Group si no existe")
	importCmd.Flags().BoolP("yes", "y", false, "No pedir confirmación al sobrescribir")
	_ = importCmd.RegisterFlagCompletionFunc("name", cmd.CompleteVariableGroups)

	cmd.Variables.AddCommand(exportCmd, importCmd)
}
