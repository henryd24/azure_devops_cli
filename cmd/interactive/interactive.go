// Package interactive implementa el menú navegable de la CLI. El menú se construye
// a partir del árbol de comandos de cobra, por lo que cualquier comando nuevo
// aparece automáticamente. Al elegir un comando se ejecuta sin flags y este
// pregunta lo que necesite.
package interactive

import (
	"errors"
	"fmt"
	"strings"

	"azuredevops/cmd"
	"azuredevops/internal/ui"

	"github.com/spf13/cobra"
)

var interactiveCmd = &cobra.Command{
	Use:     "interactive",
	Aliases: []string{"i", "menu"},
	Short:   "Abre el modo interactivo (menú navegable)",
	RunE: func(c *cobra.Command, args []string) error {
		if !ui.Interactive() {
			return fmt.Errorf("el modo interactivo requiere una terminal (y no usar --no-input)")
		}
		return Run(cmd.RootCmd)
	},
}

const (
	actionBack = "__back"
	actionExit = "__exit"
)

// Run muestra el menú principal hasta que el usuario sale.
func Run(root *cobra.Command) error {
	printHeader()
	current := root
	for {
		choice, err := ui.Select(title(current), menuOptions(current, root))
		if errors.Is(err, ui.ErrAborted) {
			if current == root {
				return nil
			}
			current = current.Parent()
			continue
		}
		if err != nil {
			return err
		}

		switch choice {
		case actionExit:
			return nil
		case actionBack:
			current = current.Parent()
			continue
		}

		next := findChild(current, choice)
		if next == nil {
			continue
		}
		if visibleChildren(next) != nil && next.RunE == nil && next.Run == nil {
			current = next
			continue
		}
		execute(next)
		if next.Name() == "login" || next.Parent().Name() == "config" {
			printHeader()
		}
	}
}

func execute(c *cobra.Command) {
	fmt.Fprintln(ui.Err)
	ui.Info("%s", ui.Bold(c.CommandPath()))
	var err error
	switch {
	case c.RunE != nil:
		err = c.RunE(c, nil)
	case c.Run != nil:
		c.Run(c, nil)
	}
	switch {
	case errors.Is(err, ui.ErrAborted):
		ui.Warn("Cancelado")
	case err != nil:
		ui.Errorf("%v", err)
	}
	fmt.Fprintln(ui.Err)
}

func printHeader() {
	settings, profile, err := cmd.ResolveSettings()
	if err != nil {
		ui.Warn("%v", err)
		return
	}
	if settings.Org == "" {
		ui.Info("Sin credenciales configuradas: elige 'login' para empezar.")
		return
	}
	ui.Info("Organización %s · Proyecto %s · Perfil %s", ui.Bold(settings.Org), ui.Bold(settings.Project), profile)
}

func title(c *cobra.Command) string {
	if !c.HasParent() {
		return "¿Qué quieres hacer?"
	}
	return strings.TrimPrefix(c.CommandPath(), c.Root().Name()+" ")
}

func menuOptions(c, root *cobra.Command) []ui.Option[string] {
	var opts []ui.Option[string]
	for _, child := range visibleChildren(c) {
		label := fmt.Sprintf("%-16s %s", child.Name(), child.Short)
		if child.HasAvailableSubCommands() {
			label = fmt.Sprintf("%-16s %s ›", child.Name(), child.Short)
		}
		opts = append(opts, ui.Option[string]{Label: label, Value: child.Name()})
	}
	if c != root {
		opts = append(opts, ui.Option[string]{Label: "← Volver", Value: actionBack})
	}
	opts = append(opts, ui.Option[string]{Label: "Salir", Value: actionExit})
	return opts
}

func visibleChildren(c *cobra.Command) []*cobra.Command {
	var out []*cobra.Command
	for _, child := range c.Commands() {
		if !child.IsAvailableCommand() || child.Deprecated != "" {
			continue
		}
		switch child.Name() {
		case "help", "completion", "interactive":
			continue
		}
		out = append(out, child)
	}
	return out
}

func findChild(c *cobra.Command, name string) *cobra.Command {
	for _, child := range c.Commands() {
		if child.Name() == name {
			return child
		}
	}
	return nil
}

func init() {
	cmd.RootCmd.AddCommand(interactiveCmd)
	cmd.InteractiveRun = func(*cobra.Command) error { return Run(cmd.RootCmd) }
}
