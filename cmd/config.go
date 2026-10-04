package cmd

import (
	"fmt"
	"os"

	"azuredevops/internal/config"
	"azuredevops/internal/ui"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Gestiona los perfiles guardados (organización/proyecto/PAT)",
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "Muestra los perfiles y la configuración efectiva",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		out, err := Output("table")
		if err != nil {
			return err
		}
		settings, active, err := ResolveSettings()
		if err != nil {
			return err
		}
		if out != "table" {
			type view struct {
				Name    string `json:"name"`
				Current bool   `json:"current"`
				Org     string `json:"org"`
				Project string `json:"project"`
				PAT     string `json:"pat"`
			}
			var views []view
			for _, n := range cfg.Names() {
				p := cfg.Profiles[n]
				views = append(views, view{n, n == cfg.Current, p.Org, p.Project, profilePAT(p)})
			}
			return Print(views)
		}

		var rows [][]string
		for _, n := range cfg.Names() {
			p := cfg.Profiles[n]
			marker := ""
			if n == cfg.Current {
				marker = "*"
			}
			rows = append(rows, []string{marker, n, p.Org, p.Project, profilePAT(p)})
		}
		if len(rows) > 0 {
			ui.Table([]string{"", "perfil", "organización", "proyecto", "pat"}, rows)
		} else {
			ui.Info("No hay perfiles guardados. Ejecuta 'azdevops login'.")
		}
		fmt.Fprintf(ui.Out, "\nConfiguración efectiva (perfil '%s' + variables de entorno + flags):\n  org=%s project=%s pat=%s\n",
			active, settings.Org, settings.Project, maskPAT(settings.PAT))
		return nil
	},
}

var configUseCmd = &cobra.Command{
	Use:   "use [perfil]",
	Short: "Cambia el perfil por defecto",
	Args:  cobra.MaximumNArgs(1),
	ValidArgsFunction: func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		cfg, _ := config.Load()
		return cfg.Names(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		var name string
		if len(args) == 1 {
			name = args[0]
		} else if ui.Interactive() {
			opts := []ui.Option[string]{}
			for _, n := range cfg.Names() {
				p := cfg.Profiles[n]
				opts = append(opts, ui.Option[string]{Label: fmt.Sprintf("%s (%s/%s)", n, p.Org, p.Project), Value: n})
			}
			if name, err = ui.Select("Perfil a usar", opts); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("indica el nombre del perfil")
		}
		if _, ok := cfg.Profiles[name]; !ok {
			return fmt.Errorf("el perfil '%s' no existe", name)
		}
		cfg.Current = name
		if err := cfg.Save(); err != nil {
			return err
		}
		ResetClient()
		ui.Success("Ahora se usa el perfil '%s'", name)
		return nil
	},
}

var configDeleteCmd = &cobra.Command{
	Use:   "delete [perfil]",
	Short: "Elimina un perfil guardado",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		var name string
		if len(args) == 1 {
			name = args[0]
		} else if ui.Interactive() {
			opts := []ui.Option[string]{}
			for _, n := range cfg.Names() {
				opts = append(opts, ui.Option[string]{Label: n, Value: n})
			}
			if name, err = ui.Select("Perfil a eliminar", opts); err != nil {
				return err
			}
			if err := Confirm(fmt.Sprintf("¿Eliminar el perfil '%s'?", name), false); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("indica el nombre del perfil")
		}
		if _, ok := cfg.Profiles[name]; !ok {
			return fmt.Errorf("el perfil '%s' no existe", name)
		}
		delete(cfg.Profiles, name)
		config.DeletePAT(name)
		if cfg.Current == name {
			cfg.Current = ""
		}
		if err := cfg.Save(); err != nil {
			return err
		}
		ResetClient()
		ui.Success("Perfil '%s' eliminado", name)
		return nil
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Muestra la ruta del archivo de configuración",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.Path()
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, path)
		return nil
	},
}

func profilePAT(p config.Profile) string {
	if p.InKeyring() {
		return "(llavero del sistema)"
	}
	return maskPAT(p.PAT)
}

func maskPAT(pat string) string {
	if pat == "" {
		return "(no definido)"
	}
	if len(pat) <= 4 {
		return "****"
	}
	return "****" + pat[len(pat)-4:]
}

func init() {
	configCmd.AddCommand(configViewCmd, configUseCmd, configDeleteCmd, configPathCmd)
	RootCmd.AddCommand(configCmd)
}
