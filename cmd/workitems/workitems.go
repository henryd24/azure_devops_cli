// Package workitems contiene los comandos de Azure Boards.
package workitems

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"azuredevops/azdevops"
	"azuredevops/azdevops/organization"
	"azuredevops/azdevops/workitem"
	"azuredevops/cmd"
	"azuredevops/internal/ui"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

var workItemsCmd = &cobra.Command{
	Use:     "workitems",
	Aliases: []string{"wi", "boards"},
	Short:   "Consulta, crea y actualiza work items (tareas, bugs, historias…)",
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "Lista work items (por defecto: los tuyos que siguen abiertos)",
	Example: `  azdevops wi list
  azdevops wi list --type Bug --state Active --assigned-to any
  azdevops wi list --search "login" --all-states
  azdevops wi list --wiql "SELECT [System.Id] FROM WorkItems WHERE [System.Tags] CONTAINS 'urgente'"`,
	RunE: func(c *cobra.Command, args []string) error {
		out, err := cmd.Output("table")
		if err != nil {
			return err
		}
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		f := c.Flags()
		wiql, _ := f.GetString("wiql")
		if wiql == "" {
			o := listOptions{}
			o.AssignedTo, _ = f.GetString("assigned-to")
			if strings.EqualFold(o.AssignedTo, "any") {
				o.AssignedTo = ""
			}
			o.Types, _ = f.GetStringSlice("type")
			o.States, _ = f.GetStringSlice("state")
			o.Search, _ = f.GetString("search")
			o.Area, _ = f.GetString("area")
			allStates, _ := f.GetBool("all-states")
			o.OpenOnly = !allStates
			wiql = buildWIQL(o)
		}
		top, _ := f.GetInt("top")

		var items []models.WorkItem
		if err := ui.Spinner("Consultando work items...", func() error {
			ids, err := workitem.Query(client, wiql, top)
			if err != nil {
				return err
			}
			items, err = workitem.GetBatch(client, ids, workitem.DefaultFields)
			return err
		}); err != nil {
			return err
		}
		if out != "table" {
			if items == nil {
				items = []models.WorkItem{}
			}
			return cmd.Print(items)
		}
		if len(items) == 0 {
			ui.Info("No se encontraron work items.")
			return nil
		}
		printList(items)
		return nil
	},
}

func printList(items []models.WorkItem) {
	rows := make([][]string, len(items))
	for i, w := range items {
		rows[i] = []string{
			strconv.Itoa(w.ID), w.Field("System.WorkItemType"), ui.Status(w.Field("System.State")),
			truncate(w.Field("System.Title"), 60), w.Field("System.AssignedTo"), formatDate(w.Field("System.ChangedDate")),
		}
	}
	ui.Table([]string{"id", "tipo", "estado", "título", "asignado a", "modificado"}, rows)
}

var getCmd = &cobra.Command{
	Use:     "get [id]",
	Aliases: []string{"show"},
	Short:   "Muestra un work item",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		out, err := cmd.Output("table")
		if err != nil {
			return err
		}
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		id, err := resolveID(c, args, client)
		if err != nil {
			return err
		}
		w, err := workitem.Get(client, id)
		if err != nil {
			return err
		}
		if out != "table" {
			return cmd.Print(w)
		}
		printDetail(client, w)
		return nil
	},
}

func printDetail(client *azdevops.Client, w *models.WorkItem) {
	fmt.Fprintf(ui.Out, "%s %s\n", ui.Bold(fmt.Sprintf("%s %d:", w.Field("System.WorkItemType"), w.ID)), w.Field("System.Title"))
	rows := [][]string{
		{"Estado", ui.Status(w.Field("System.State")) + reason(w)},
		{"Asignado a", orDash(w.Field("System.AssignedTo"))},
		{"Área", w.Field("System.AreaPath")},
		{"Iteración", w.Field("System.IterationPath")},
		{"Etiquetas", orDash(w.Field("System.Tags"))},
		{"Creado", fmt.Sprintf("%s por %s", formatDate(w.Field("System.CreatedDate")), w.Field("System.CreatedBy"))},
		{"Modificado", fmt.Sprintf("%s por %s", formatDate(w.Field("System.ChangedDate")), w.Field("System.ChangedBy"))},
	}
	if p := w.Field("Microsoft.VSTS.Common.Priority"); p != "" {
		rows = append(rows, []string{"Prioridad", p})
	}
	for _, r := range w.Relations {
		name, _ := r.Attributes["name"].(string)
		if name == "Parent" || name == "Child" {
			parts := strings.Split(r.URL, "/")
			rows = append(rows, []string{map[string]string{"Parent": "Padre", "Child": "Hijo"}[name], "#" + parts[len(parts)-1]})
		}
	}
	for _, r := range rows {
		fmt.Fprintf(ui.Out, "  %-12s %s\n", r[0]+":", r[1])
	}
	desc := w.Field("System.Description")
	if desc == "" {
		desc = w.Field("Microsoft.VSTS.TCM.ReproSteps")
	}
	if desc != "" {
		fmt.Fprintf(ui.Out, "\n%s\n", htmlToText(desc))
	}
	ui.Link(workitem.WebURL(client, w.ID))
}

func reason(w *models.WorkItem) string {
	if r := w.Field("System.Reason"); r != "" {
		return " (" + r + ")"
	}
	return ""
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Crea un work item",
	Example: `  azdevops wi create --type Bug --title "Falla el login" --assigned-to @me --tags "frontend; urgente"
  azdevops wi create --type Task --title "Escribir tests" --parent 123
  azdevops wi create   # modo guiado`,
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		f := c.Flags()
		itemType, _ := f.GetString("type")
		title, _ := f.GetString("title")
		description, _ := f.GetString("description")
		assignedTo, _ := f.GetString("assigned-to")
		parent, _ := f.GetInt("parent")
		fieldEntries, _ := f.GetStringSlice("field")
		fields, err := parseFields(fieldEntries)
		if err != nil {
			return err
		}

		if title == "" {
			if !ui.Interactive() {
				return ui.MissingError("title")
			}
			if !f.Changed("type") {
				if itemType, err = pickType(client); err != nil {
					return err
				}
			}
			if title, err = ui.Input("Título", "", "", ui.Required); err != nil {
				return err
			}
			if description, err = ui.Input("Descripción (opcional)", "", description, nil); err != nil {
				return err
			}
			if assignedTo == "" {
				mine, err := ui.Confirm("¿Asignármelo?", true)
				if err != nil {
					return err
				}
				if mine {
					assignedTo = "@me"
				}
			}
		}

		fields["System.Title"] = title
		if description != "" {
			fields["System.Description"] = textToHTML(description)
		}
		if assignedTo != "" {
			if fields["System.AssignedTo"], err = resolveAssignee(client, assignedTo); err != nil {
				return err
			}
		}
		for flag, field := range map[string]string{"area": "System.AreaPath", "iteration": "System.IterationPath", "tags": "System.Tags"} {
			if v, _ := f.GetString(flag); v != "" {
				fields[field] = v
			}
		}

		w, err := workitem.Create(client, itemType, fields, parent)
		if err != nil {
			return err
		}
		ui.Success("%s %d creado: %s", w.Field("System.WorkItemType"), w.ID, w.Field("System.Title"))
		ui.Link(workitem.WebURL(client, w.ID))
		if cmd.WantsData() {
			return cmd.Print(w)
		}
		return nil
	},
}

var updateCmd = &cobra.Command{
	Use:   "update [id]",
	Short: "Actualiza estado, título, asignación o etiquetas, o agrega un comentario",
	Example: `  azdevops wi update 123 --state Active --comment "Empiezo con esto"
  azdevops wi update 123 --assigned-to ana@empresa.com
  azdevops wi update   # elegir uno de tus work items y qué cambiar`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		id, err := resolveID(c, args, client)
		if err != nil {
			return err
		}
		f := c.Flags()
		fieldEntries, _ := f.GetStringSlice("field")
		fields, err := parseFields(fieldEntries)
		if err != nil {
			return err
		}
		for flag, field := range map[string]string{"state": "System.State", "title": "System.Title", "tags": "System.Tags", "comment": "System.History"} {
			if v, _ := f.GetString(flag); v != "" {
				fields[field] = v
			}
		}
		if v, _ := f.GetString("assigned-to"); v != "" {
			if fields["System.AssignedTo"], err = resolveAssignee(client, v); err != nil {
				return err
			}
		}

		if len(fields) == 0 {
			if !ui.Interactive() {
				return fmt.Errorf("no hay nada que actualizar: usa --state, --title, --assigned-to, --tags, --comment o --field")
			}
			if fields, err = promptUpdate(client, id); err != nil {
				return err
			}
			if len(fields) == 0 {
				ui.Info("No hay cambios.")
				return nil
			}
		}

		w, err := workitem.Update(client, id, fields)
		if err != nil {
			return err
		}
		ui.Success("Work item %d actualizado (estado: %s)", w.ID, w.Field("System.State"))
		ui.Link(workitem.WebURL(client, w.ID))
		if cmd.WantsData() {
			return cmd.Print(w)
		}
		return nil
	},
}

func promptUpdate(client *azdevops.Client, id int) (map[string]any, error) {
	w, err := workitem.Get(client, id)
	if err != nil {
		return nil, err
	}
	ui.Info("%s %d: %s (%s)", w.Field("System.WorkItemType"), w.ID, w.Field("System.Title"), w.Field("System.State"))
	changes, err := ui.MultiSelect("¿Qué quieres cambiar?", []ui.Option[string]{
		{Label: "Estado", Value: "state"},
		{Label: "Agregar comentario", Value: "comment"},
		{Label: "Asignación", Value: "assign"},
		{Label: "Título", Value: "title"},
		{Label: "Etiquetas", Value: "tags"},
	})
	if err != nil {
		return nil, err
	}
	fields := map[string]any{}
	for _, ch := range changes {
		switch ch {
		case "state":
			states, err := workitem.States(client, w.Field("System.WorkItemType"))
			if err != nil {
				return nil, err
			}
			opts := make([]ui.Option[string], 0, len(states))
			for _, s := range states {
				if s != w.Field("System.State") {
					opts = append(opts, ui.Option[string]{Label: s, Value: s})
				}
			}
			if fields["System.State"], err = ui.Select("Nuevo estado", opts); err != nil {
				return nil, err
			}
		case "comment":
			v, err := ui.Input("Comentario", "", "", ui.Required)
			if err != nil {
				return nil, err
			}
			fields["System.History"] = textToHTML(v)
		case "assign":
			v, err := ui.Input("Asignar a", "Email, @me, o vacío para quitar la asignación", "@me", nil)
			if err != nil {
				return nil, err
			}
			if v == "" {
				fields["System.AssignedTo"] = ""
			} else if fields["System.AssignedTo"], err = resolveAssignee(client, v); err != nil {
				return nil, err
			}
		case "title":
			if fields["System.Title"], err = ui.Input("Título", "", w.Field("System.Title"), ui.Required); err != nil {
				return nil, err
			}
		case "tags":
			if fields["System.Tags"], err = ui.Input("Etiquetas", "Separadas por ';'", w.Field("System.Tags"), nil); err != nil {
				return nil, err
			}
		}
	}
	return fields, nil
}

var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Envía un work item a la papelera de reciclaje",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		id, err := resolveID(c, args, client)
		if err != nil {
			return err
		}
		w, err := workitem.Get(client, id)
		if err != nil {
			return err
		}
		yes, _ := c.Flags().GetBool("yes")
		if err := cmd.Confirm(fmt.Sprintf("¿Enviar a la papelera '%s %d: %s'?", w.Field("System.WorkItemType"), id, w.Field("System.Title")), yes); err != nil {
			return err
		}
		if err := workitem.Delete(client, id); err != nil {
			return err
		}
		ui.Success("Work item %d enviado a la papelera (se puede restaurar desde Boards)", id)
		return nil
	},
}

// resolveID toma el ID del argumento, de --id o de un selector con tus work items abiertos.
func resolveID(c *cobra.Command, args []string, client *azdevops.Client) (int, error) {
	if len(args) == 1 {
		id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
		if err != nil {
			return 0, fmt.Errorf("ID de work item inválido: %s", args[0])
		}
		return id, nil
	}
	if id, _ := c.Flags().GetInt("id"); id != 0 {
		return id, nil
	}
	if !ui.Interactive() {
		return 0, fmt.Errorf("indica el ID del work item")
	}
	var items []models.WorkItem
	if err := ui.Spinner("Cargando tus work items...", func() error {
		ids, err := workitem.Query(client, buildWIQL(listOptions{AssignedTo: "@me", OpenOnly: true}), 100)
		if err != nil {
			return err
		}
		items, err = workitem.GetBatch(client, ids, workitem.DefaultFields)
		return err
	}); err != nil {
		return 0, err
	}
	if len(items) == 0 {
		v, err := ui.Input("ID del work item", "No tienes work items abiertos asignados", "", cmd.PositiveInt)
		if err != nil {
			return 0, err
		}
		return strconv.Atoi(v)
	}
	opts := make([]ui.Option[int], len(items))
	for i, w := range items {
		opts[i] = ui.Option[int]{Label: fmt.Sprintf("#%d  [%s] %s  · %s", w.ID, w.Field("System.WorkItemType"), w.Field("System.Title"), w.Field("System.State")), Value: w.ID}
	}
	return ui.Select("Work item", opts)
}

// resolveAssignee traduce "@me" al email del dueño del PAT.
func resolveAssignee(client *azdevops.Client, v string) (string, error) {
	if !strings.EqualFold(v, "@me") {
		return v, nil
	}
	u, err := organization.CurrentUser(client)
	if err != nil {
		return "", err
	}
	if u.Email != "" {
		return u.Email, nil
	}
	return u.DisplayName, nil
}

func pickType(client *azdevops.Client) (string, error) {
	types, err := workitem.Types(client)
	if err != nil {
		return "", err
	}
	common := map[string]int{"Task": 0, "Bug": 1, "User Story": 2, "Product Backlog Item": 2, "Issue": 2, "Feature": 3, "Epic": 4}
	var first, rest []ui.Option[string]
	for _, t := range types {
		o := ui.Option[string]{Label: t.Name, Value: t.Name}
		if _, ok := common[t.Name]; ok {
			first = append(first, o)
		} else {
			rest = append(rest, o)
		}
	}
	return ui.Select("Tipo de work item", append(first, rest...))
}

func formatDate(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return orDash(s)
	}
	return ui.FormatTime(&t)
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func init() {
	lf := listCmd.Flags()
	lf.String("assigned-to", "@me", "Asignado a: @me, un email/nombre o 'any'")
	lf.StringSliceP("type", "t", nil, "Tipo (Bug, Task, User Story…; se puede repetir)")
	lf.StringSliceP("state", "s", nil, "Estado (se puede repetir)")
	lf.Bool("all-states", false, "Incluir también los cerrados/terminados")
	lf.String("search", "", "Texto contenido en el título")
	lf.String("area", "", "Ruta de área (incluye sub-áreas)")
	lf.String("wiql", "", "Consulta WIQL completa (ignora los demás filtros)")
	lf.Int("top", 50, "Número máximo de resultados")

	cf := createCmd.Flags()
	cf.StringP("type", "t", "Task", "Tipo de work item")
	cf.String("title", "", "Título")
	cf.StringP("description", "d", "", "Descripción (texto plano)")
	cf.String("assigned-to", "", "Asignar a: @me o email")
	cf.String("area", "", "Ruta de área")
	cf.String("iteration", "", "Ruta de iteración")
	cf.String("tags", "", "Etiquetas separadas por ';'")
	cf.Int("parent", 0, "ID del work item padre")
	cf.StringSlice("field", nil, "Otro campo: Referencia=Valor (p. ej. Microsoft.VSTS.Common.Priority=1)")

	uf := updateCmd.Flags()
	uf.String("state", "", "Nuevo estado")
	uf.String("title", "", "Nuevo título")
	uf.String("assigned-to", "", "Asignar a: @me o email")
	uf.String("tags", "", "Etiquetas (reemplaza las actuales)")
	uf.StringP("comment", "m", "", "Agregar un comentario")
	uf.StringSlice("field", nil, "Otro campo: Referencia=Valor")

	for _, c := range []*cobra.Command{getCmd, updateCmd, deleteCmd} {
		c.Flags().IntP("id", "i", 0, "ID del work item (también se acepta como argumento)")
	}
	deleteCmd.Flags().BoolP("yes", "y", false, "Confirmar sin preguntar")

	workItemsCmd.AddCommand(listCmd, getCmd, createCmd, updateCmd, deleteCmd)
	cmd.RootCmd.AddCommand(workItemsCmd)
}
