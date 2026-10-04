package scaffold

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Niveles de incremento de versión.
const (
	LevelMajor = "major"
	LevelMinor = "minor"
	LevelPatch = "patch"
)

// Change describe una modificación hecha por Bump o ResetIDs.
type Change struct {
	File string `json:"file"`
	From string `json:"from"`
	To   string `json:"to"`
}

type taskVersion struct {
	Major int `json:"Major"`
	Minor int `json:"Minor"`
	Patch int `json:"Patch"`
}

func (v taskVersion) String() string { return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch) }

func bumpParts(major, minor, patch int, level string) (int, int, int) {
	switch level {
	case LevelMajor:
		return major + 1, 0, 0
	case LevelMinor:
		return major, minor + 1, 0
	}
	return major, minor, patch + 1
}

// BumpVersion incrementa una versión "X.Y.Z".
func BumpVersion(v, level string) (string, error) {
	if !IsSemver(v) {
		return "", fmt.Errorf("versión inválida '%s' (se esperaba X.Y.Z)", v)
	}
	p := strings.Split(v, ".")
	a, _ := strconv.Atoi(p[0])
	b, _ := strconv.Atoi(p[1])
	c, _ := strconv.Atoi(p[2])
	a, b, c = bumpParts(a, b, c, level)
	return fmt.Sprintf("%d.%d.%d", a, b, c), nil
}

// BumpTasks incrementa la versión de las tareas indicadas (todas si only está vacío).
func BumpTasks(root, level string, only []string, dryRun bool) ([]Change, error) {
	dirs, err := TaskDirs(root)
	if err != nil {
		return nil, err
	}
	selected := map[string]bool{}
	for _, o := range only {
		selected[o] = true
	}
	var changes []Change
	for _, d := range dirs {
		if len(only) > 0 && !selected[d] {
			continue
		}
		delete(selected, d)
		path := filepath.Join(root, "src", "tasks", d, "task.json")
		obj, err := ReadObject(path)
		if err != nil {
			return changes, err
		}
		var v taskVersion
		if ok, err := obj.Get("version", &v); !ok || err != nil {
			return changes, fmt.Errorf("%s: falta 'version' o es inválida", path)
		}
		nv := v
		nv.Major, nv.Minor, nv.Patch = bumpParts(v.Major, v.Minor, v.Patch, level)
		if !dryRun {
			if err := obj.Set("version", nv); err != nil {
				return changes, err
			}
			if err := obj.Write(path); err != nil {
				return changes, err
			}
		}
		changes = append(changes, Change{File: filepath.Join("src", "tasks", d, "task.json"), From: v.String(), To: nv.String()})
	}
	for missing := range selected {
		return changes, fmt.Errorf("no existe la tarea '%s'", missing)
	}
	return changes, nil
}

// BumpConfigs incrementa "version" en los archivos de config indicados (p. ej. dev, release).
func BumpConfigs(root, level string, configs []string, dryRun bool) ([]Change, error) {
	var changes []Change
	for _, name := range configs {
		rel := filepath.Join("config", name+".json")
		path := filepath.Join(root, rel)
		obj, err := ReadObject(path)
		if err != nil {
			return changes, err
		}
		var v string
		if ok, err := obj.Get("version", &v); !ok || err != nil {
			return changes, fmt.Errorf("%s: falta 'version'", rel)
		}
		nv, err := BumpVersion(v, level)
		if err != nil {
			return changes, fmt.Errorf("%s: %w", rel, err)
		}
		if !dryRun {
			if err := obj.Set("version", nv); err != nil {
				return changes, err
			}
			if err := obj.Write(path); err != nil {
				return changes, err
			}
		}
		changes = append(changes, Change{File: rel, From: v, To: nv})
	}
	return changes, nil
}

// ResetIDs genera un UUID nuevo para cada tarea (útil al partir de otra extensión).
func ResetIDs(root string) ([]Change, error) {
	dirs, err := TaskDirs(root)
	if err != nil {
		return nil, err
	}
	var changes []Change
	for _, d := range dirs {
		path := filepath.Join(root, "src", "tasks", d, "task.json")
		obj, err := ReadObject(path)
		if err != nil {
			return changes, err
		}
		var old string
		_, _ = obj.Get("id", &old)
		id := NewUUID()
		if err := obj.Set("id", id); err != nil {
			return changes, err
		}
		if err := obj.Write(path); err != nil {
			return changes, err
		}
		changes = append(changes, Change{File: filepath.Join("src", "tasks", d, "task.json"), From: old, To: id})
	}
	return changes, nil
}

// Issue es un problema detectado por Validate.
type Issue struct {
	Severity string `json:"severity"` // "error" o "warning"
	File     string `json:"file"`
	Message  string `json:"message"`
}

// Validate revisa la estructura de la extensión antes de empaquetarla.
func Validate(root string) []Issue {
	var issues []Issue
	add := func(sev, file, format string, a ...any) {
		issues = append(issues, Issue{Severity: sev, File: file, Message: fmt.Sprintf(format, a...)})
	}

	manifest, err := ReadObject(filepath.Join(root, "vss-extension.json"))
	if err != nil {
		add("error", "vss-extension.json", "%v", err)
		return issues
	}
	var contributions []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Properties struct {
			Name string `json:"name"`
			Icon string `json:"icon"`
		} `json:"properties"`
	}
	_, _ = manifest.Get("contributions", &contributions)
	registered := map[string]bool{}
	contribIDs := map[string]bool{}
	for _, c := range contributions {
		if contribIDs[c.ID] {
			add("error", "vss-extension.json", "id de contribución duplicado: %s", c.ID)
		}
		contribIDs[c.ID] = true
		switch c.Type {
		case "ms.vss-distributed-task.task":
			registered[strings.TrimPrefix(c.Properties.Name, "tasks/")] = true
		case "ms.vss-endpoint.service-endpoint-type":
			if c.Properties.Icon != "" {
				if _, err := os.Stat(filepath.Join(root, c.Properties.Icon)); err != nil {
					add("error", "vss-extension.json", "no existe el icono %s de la conexión '%s'", c.Properties.Icon, c.ID)
				}
			}
		}
	}
	var icons struct {
		Default string `json:"default"`
	}
	if ok, _ := manifest.Get("icons", &icons); ok && icons.Default != "" {
		if _, err := os.Stat(filepath.Join(root, icons.Default)); err != nil {
			add("error", "vss-extension.json", "no existe el icono de la extensión %s", icons.Default)
		}
	}

	for _, name := range []string{"dev", "release"} {
		rel := filepath.Join("config", name+".json")
		obj, err := ReadObject(filepath.Join(root, rel))
		if err != nil {
			add("warning", rel, "no se pudo leer: %v", err)
			continue
		}
		var v, pub string
		_, _ = obj.Get("version", &v)
		_, _ = obj.Get("publisher", &pub)
		if !IsSemver(v) {
			add("error", rel, "versión inválida '%s'", v)
		}
		if pub == "" {
			add("error", rel, "falta 'publisher'")
		}
	}

	dirs, err := TaskDirs(root)
	if err != nil {
		add("error", "src/tasks", "%v", err)
		return issues
	}
	ids := map[string]string{}
	names := map[string]string{}
	for _, d := range dirs {
		rel := filepath.Join("src", "tasks", d, "task.json")
		obj, err := ReadObject(filepath.Join(root, rel))
		if err != nil {
			add("error", rel, "%v", err)
			continue
		}
		var id, name string
		var execution map[string]struct {
			Target string `json:"target"`
		}
		_, _ = obj.Get("id", &id)
		_, _ = obj.Get("name", &name)
		_, _ = obj.Get("execution", &execution)
		if !IsUUID(id) {
			add("error", rel, "el id '%s' no es un UUID válido", id)
		} else if other, dup := ids[strings.ToLower(id)]; dup {
			add("error", rel, "el id %s está repetido en %s (usa 'extension reset-ids')", id, other)
		}
		ids[strings.ToLower(id)] = d
		if other, dup := names[name]; dup {
			add("error", rel, "el name '%s' está repetido en %s", name, other)
		}
		names[name] = d
		if !registered[d] {
			add("error", "vss-extension.json", "la tarea '%s' no está registrada en contributions", d)
		}
		delete(registered, d)
		if len(execution) == 0 {
			add("error", rel, "falta 'execution'")
		}
		if _, err := os.Stat(filepath.Join(root, "src", "tasks", d, "icon.png")); err != nil {
			add("warning", filepath.Join("src", "tasks", d), "falta icon.png")
		}
		var pkg struct {
			Scripts map[string]string `json:"scripts"`
		}
		if data, err := os.ReadFile(filepath.Join(root, "src", "tasks", d, "package.json")); err == nil {
			_ = json.Unmarshal(data, &pkg)
		}
		if pkg.Scripts["package"] == "" {
			add("warning", filepath.Join("src", "tasks", d, "package.json"), "no tiene script 'package'")
		}
	}
	for d := range registered {
		add("error", "vss-extension.json", "la contribución apunta a tasks/%s pero no existe src/tasks/%s", d, d)
	}
	return issues
}
