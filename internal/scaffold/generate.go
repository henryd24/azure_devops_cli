package scaffold

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"
)

//go:embed all:templates
var templatesFS embed.FS

// Categorías de task.json y del Marketplace.
var (
	TaskCategories      = []string{"Utility", "Build", "Deploy", "Test", "Package", "Tool", "Azure Pipelines"}
	ExtensionCategories = []string{"Azure Pipelines", "Azure Repos", "Azure Boards", "Azure Test Plans", "Azure Artifacts"}
)

// Project describe la extensión.
type Project struct {
	ID           string // id del manifiesto (release); dev usa ID + "-dev"
	Name         string
	Description  string
	Publisher    string
	DevPublisher string
	Author       string
	Repository   string
	Category     string
	Public       bool
	Version      string
	LicenseID    string // "MIT" o "UNLICENSED"
}

// Endpoint describe un tipo de conexión de servicio personalizado.
type Endpoint struct {
	ID          string // id de la contribución
	Name        string // nombre del tipo, usado en connectedService:<Name>
	DisplayName string
	Description string
	Icon        string // archivo dentro de images/
	URL         string
}

// Task describe una tarea de pipeline.
type Task struct {
	Slug         string // carpeta en src/tasks y tasks/
	Name         string // name en task.json (único en la organización)
	FriendlyName string
	Description  string
	Category     string
	UUID         string
	Major        int
	Minor        int
	Patch        int
	Entry        string // archivo principal sin extensión
	Endpoint     *Endpoint
}

// NewTask completa los valores derivados de una tarea.
func NewTask(publisher, slug, friendly, description, category string, endpoint *Endpoint) Task {
	if friendly == "" {
		friendly = Title(slug)
	}
	if description == "" {
		description = friendly
	}
	if category == "" {
		category = "Utility"
	}
	return Task{
		Slug: slug, Name: publisher + "-" + slug, FriendlyName: friendly, Description: description,
		Category: category, UUID: NewUUID(), Major: 0, Minor: 1, Patch: 0, Entry: Snake(slug), Endpoint: endpoint,
	}
}

// NewEndpoint crea la definición de un tipo de conexión de servicio para la extensión.
func NewEndpoint(projectID, displayName, url string) *Endpoint {
	slug := Slugify(projectID)
	if displayName == "" {
		displayName = Title(slug)
	}
	if url == "" {
		url = "https://"
	}
	return &Endpoint{
		ID: "service-endpoint", Name: Camel(slug) + "Service", DisplayName: displayName,
		Description: "Service connection for " + displayName, Icon: "service-endpoint.png", URL: url,
	}
}

type templateData struct {
	Project   Project
	Task      Task
	Tasks     []Task
	Endpoints []Endpoint
	Year      int
}

var funcs = template.FuncMap{
	// j escapa un texto para usarlo dentro de una cadena JSON.
	"j": func(s string) string {
		b, _ := json.Marshal(s)
		return string(b[1 : len(b)-1])
	},
}

func render(name string, data templateData) ([]byte, error) {
	src, err := templatesFS.ReadFile("templates/" + name)
	if err != nil {
		return nil, err
	}
	tpl, err := template.New(name).Funcs(funcs).Option("missingkey=error").Parse(string(src))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("plantilla %s: %w", name, err)
	}
	return buf.Bytes(), nil
}

// fileWriter escribe archivos y registra los creados. Nunca sobrescribe.
type fileWriter struct {
	root    string
	created []string
}

func (w *fileWriter) write(rel string, data []byte, mode os.FileMode) error {
	path := filepath.Join(w.root, rel)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s ya existe", rel)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, data, mode); err != nil {
		return err
	}
	w.created = append(w.created, rel)
	return nil
}

func (w *fileWriter) renderTo(tpl, rel string, data templateData) error {
	out, err := render(tpl, data)
	if err != nil {
		return err
	}
	if strings.HasSuffix(rel, ".json") {
		var v any
		if err := json.Unmarshal(out, &v); err != nil {
			return fmt.Errorf("%s generado no es JSON válido: %w", rel, err)
		}
	}
	return w.write(rel, out, 0o644)
}

var projectFiles = map[string]string{
	"project/package.json.tmpl":        "package.json",
	"project/vss-extension.json.tmpl":  "vss-extension.json",
	"project/config/dev.json.tmpl":     "config/dev.json",
	"project/config/release.json.tmpl": "config/release.json",
	"project/README.md.tmpl":           "README.md",
	"project/gitignore.tmpl":           ".gitignore",
	"project/scripts/tasks.js.tmpl":    "scripts/tasks.js",
}

var taskFiles = map[string]string{
	"task/package.json.tmpl":                 "package.json",
	"task/tsconfig.json.tmpl":                "tsconfig.json",
	"task/tsconfig.build.json.tmpl":          "tsconfig.build.json",
	"task/jest.config.js.tmpl":               "jest.config.js",
	"task/testSetup.js.tmpl":                 "testSetup.js",
	"task/env-test.tmpl":                     ".env-test",
	"task/task.json.tmpl":                    "task.json",
	"task/src/utils/inputs.ts.tmpl":          "src/utils/inputs.ts",
	"task/src/__tests__/inputs.test.ts.tmpl": "src/__tests__/inputs.test.ts",
}

// Generate crea un proyecto de extensión en dir (que debe estar vacío o no existir).
// Devuelve las rutas relativas de los archivos creados.
func Generate(dir string, p Project, tasks []Task, endpoints []Endpoint) ([]string, error) {
	if entries, err := os.ReadDir(dir); err == nil && len(entries) > 0 {
		return nil, fmt.Errorf("el directorio %s no está vacío", dir)
	}
	if len(tasks) == 0 {
		return nil, errors.New("la extensión necesita al menos una tarea")
	}
	data := templateData{Project: p, Tasks: tasks, Endpoints: endpoints, Year: time.Now().Year()}
	w := &fileWriter{root: dir}
	for tpl, rel := range projectFiles {
		if err := w.renderTo(tpl, rel, data); err != nil {
			return w.created, err
		}
	}
	if p.LicenseID == "MIT" {
		if err := w.renderTo("project/LICENSE.tmpl", "LICENSE", data); err != nil {
			return w.created, err
		}
	}
	if err := w.write("images/extension-icon.png", PlaceholderIcon(128, p.ID), 0o644); err != nil {
		return w.created, err
	}
	for _, e := range endpoints {
		if err := w.write("images/"+e.Icon, PlaceholderIcon(64, e.Name), 0o644); err != nil {
			return w.created, err
		}
	}
	for _, t := range tasks {
		if err := writeTask(w, p, t); err != nil {
			return w.created, err
		}
	}
	sort.Strings(w.created)
	return w.created, nil
}

func writeTask(w *fileWriter, p Project, t Task) error {
	data := templateData{Project: p, Task: t, Year: time.Now().Year()}
	base := filepath.Join("src", "tasks", t.Slug)
	for tpl, rel := range taskFiles {
		if err := w.renderTo(tpl, filepath.Join(base, rel), data); err != nil {
			return err
		}
	}
	if err := w.renderTo("task/src/entry.ts.tmpl", filepath.Join(base, "src", t.Entry+".ts"), data); err != nil {
		return err
	}
	return w.write(filepath.Join(base, "icon.png"), PlaceholderIcon(128, t.Name), 0o644)
}

// FindRoot busca hacia arriba desde dir la carpeta que contiene vss-extension.json.
func FindRoot(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for d := abs; ; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "vss-extension.json")); err == nil {
			return d, nil
		}
		if filepath.Dir(d) == d {
			return "", fmt.Errorf("no se encontró vss-extension.json en %s ni en sus carpetas superiores", abs)
		}
	}
}

// TaskDirs devuelve las carpetas de src/tasks que contienen task.json.
func TaskDirs(root string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(root, "src", "tasks"))
	if err != nil {
		return nil, err
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			if _, err := os.Stat(filepath.Join(root, "src", "tasks", e.Name(), "task.json")); err == nil {
				dirs = append(dirs, e.Name())
			}
		}
	}
	return dirs, nil
}

// ProjectInfo lee del proyecto existente los datos necesarios para agregar tareas.
func ProjectInfo(root string) (Project, []Endpoint, error) {
	var p Project
	manifest, err := ReadObject(filepath.Join(root, "vss-extension.json"))
	if err != nil {
		return p, nil, err
	}
	_, _ = manifest.Get("id", &p.ID)
	_, _ = manifest.Get("name", &p.Name)
	_, _ = manifest.Get("publisher", &p.Publisher)
	if release, err := ReadObject(filepath.Join(root, "config", "release.json")); err == nil {
		if p.Publisher == "" {
			_, _ = release.Get("publisher", &p.Publisher)
		}
	}
	if p.Publisher == "" {
		p.Publisher = Slugify(p.ID)
	}
	var pkg struct {
		Author  string `json:"author"`
		License string `json:"license"`
	}
	if data, err := os.ReadFile(filepath.Join(root, "package.json")); err == nil {
		_ = json.Unmarshal(data, &pkg)
	}
	p.Author, p.LicenseID = pkg.Author, pkg.License
	if p.LicenseID == "" {
		p.LicenseID = "UNLICENSED"
	}

	var contributions []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Properties struct {
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
			URL         struct {
				Value string `json:"value"`
			} `json:"url"`
		} `json:"properties"`
	}
	_, _ = manifest.Get("contributions", &contributions)
	var endpoints []Endpoint
	for _, c := range contributions {
		if c.Type == "ms.vss-endpoint.service-endpoint-type" {
			endpoints = append(endpoints, Endpoint{ID: c.ID, Name: c.Properties.Name, DisplayName: c.Properties.DisplayName, URL: c.Properties.URL.Value})
		}
	}
	return p, endpoints, nil
}

// AddTask crea una tarea nueva en un proyecto existente y la registra en vss-extension.json.
func AddTask(root string, p Project, t Task) ([]string, error) {
	if _, err := os.Stat(filepath.Join(root, "src", "tasks", t.Slug)); err == nil {
		return nil, fmt.Errorf("ya existe la tarea src/tasks/%s", t.Slug)
	}
	manifestPath := filepath.Join(root, "vss-extension.json")
	manifest, err := ReadObject(manifestPath)
	if err != nil {
		return nil, err
	}
	var contributions []json.RawMessage
	if _, err := manifest.Get("contributions", &contributions); err != nil {
		return nil, fmt.Errorf("vss-extension.json: contributions inválido: %w", err)
	}
	// Insertar después de la última tarea para mantenerlas agrupadas.
	contribution := map[string]any{
		"id":         t.Slug + "-task",
		"type":       "ms.vss-distributed-task.task",
		"targets":    []string{"ms.vss-distributed-task.tasks"},
		"properties": map[string]string{"name": "tasks/" + t.Slug},
	}
	ordered := &Object{vals: map[string]json.RawMessage{}}
	for _, k := range []string{"id", "type", "targets", "properties"} {
		_ = ordered.Set(k, contribution[k])
	}
	raw, err := ordered.Bytes()
	if err != nil {
		return nil, err
	}
	insertAt := 0
	for i, c := range contributions {
		var head struct{ Type string }
		if json.Unmarshal(c, &head) == nil && head.Type == "ms.vss-distributed-task.task" {
			insertAt = i + 1
		}
	}
	contributions = append(contributions[:insertAt], append([]json.RawMessage{json.RawMessage(bytes.TrimSpace(raw))}, contributions[insertAt:]...)...)

	w := &fileWriter{root: root}
	if err := writeTask(w, p, t); err != nil {
		_ = os.RemoveAll(filepath.Join(root, "src", "tasks", t.Slug))
		return nil, err
	}
	if err := manifest.Set("contributions", contributions); err != nil {
		return nil, err
	}
	if err := manifest.Write(manifestPath); err != nil {
		return nil, err
	}
	sort.Strings(w.created)
	return append(w.created, "vss-extension.json (contribución "+t.Slug+"-task)"), nil
}

// listTemplates se usa en los tests para comprobar que todas las plantillas están mapeadas.
func listTemplates() ([]string, error) {
	var out []string
	err := fs.WalkDir(templatesFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			out = append(out, strings.TrimPrefix(path, "templates/"))
		}
		return err
	})
	return out, err
}
