package scaffold

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sampleProject() (Project, []Task, []Endpoint) {
	p := Project{
		ID: "mi-extension", Name: "Mi \"Extensión\"", Description: "Hace <cosas>", Publisher: "henda",
		DevPublisher: "dev-henda", Author: "Henry", Repository: "https://github.com/x/y",
		Category: "Azure Pipelines", Version: "0.1.0", LicenseID: "MIT",
	}
	ep := NewEndpoint(p.ID, "Mi Servicio", "https://api.example.com")
	tasks := []Task{
		NewTask(p.Publisher, "exec-thing", "Exec Thing", "", "Utility", ep),
		NewTask(p.Publisher, "otra", "", "", "", nil),
	}
	return p, tasks, []Endpoint{*ep}
}

func TestGenerateProducesValidProject(t *testing.T) {
	dir := t.TempDir()
	p, tasks, eps := sampleProject()
	created, err := Generate(dir, p, tasks, eps)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"package.json", "vss-extension.json", "config/dev.json", "config/release.json", "LICENSE", ".gitignore",
		"images/extension-icon.png", "images/service-endpoint.png", "scripts/tasks.js",
		"src/tasks/exec-thing/task.json", "src/tasks/exec-thing/src/exec_thing.ts", "src/tasks/exec-thing/icon.png",
		"src/tasks/exec-thing/.env-test", "src/tasks/otra/src/otra.ts",
	} {
		if _, err := os.Stat(filepath.Join(dir, want)); err != nil {
			t.Errorf("falta %s", want)
		}
	}
	if issues := Validate(dir); len(issues) > 0 {
		t.Errorf("el proyecto generado tiene problemas: %+v", issues)
	}

	var task struct {
		ID     string
		Name   string
		Inputs []struct{ Name, Type string }
	}
	data, _ := os.ReadFile(filepath.Join(dir, "src/tasks/exec-thing/task.json"))
	if err := json.Unmarshal(data, &task); err != nil {
		t.Fatal(err)
	}
	if !IsUUID(task.ID) || task.Name != "henda-exec-thing" {
		t.Errorf("task.json incorrecto: %+v", task)
	}
	if task.Inputs[1].Type != "connectedService:miExtensionService" {
		t.Errorf("input de conexión incorrecto: %+v", task.Inputs)
	}
	var other struct{ ID string }
	data, _ = os.ReadFile(filepath.Join(dir, "src/tasks/otra/task.json"))
	_ = json.Unmarshal(data, &other)
	if other.ID == task.ID {
		t.Error("cada tarea debe tener su propio UUID")
	}
	manifest, _ := os.ReadFile(filepath.Join(dir, "vss-extension.json"))
	if !strings.Contains(string(manifest), `"name": "Mi \"Extensión\""`) {
		t.Errorf("escape JSON incorrecto en el manifiesto:\n%s", manifest)
	}
	if len(created) < 20 {
		t.Errorf("se crearon pocos archivos: %d", len(created))
	}
	if _, err := Generate(dir, p, tasks, eps); err == nil {
		t.Error("no debe generar en un directorio con contenido")
	}
}

func TestAllTemplatesAreUsed(t *testing.T) {
	names, err := listTemplates()
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range names {
		_, inProject := projectFiles[n]
		_, inTask := taskFiles[n]
		if !inProject && !inTask && n != "project/LICENSE.tmpl" && n != "task/src/entry.ts.tmpl" {
			t.Errorf("plantilla sin usar: %s", n)
		}
	}
}

func TestAddTaskBumpAndResetIDs(t *testing.T) {
	dir := t.TempDir()
	p, tasks, eps := sampleProject()
	if _, err := Generate(dir, p, tasks[:1], eps); err != nil {
		t.Fatal(err)
	}
	info, endpoints, err := ProjectInfo(dir)
	if err != nil || info.Publisher != "henda" || len(endpoints) != 1 || endpoints[0].Name != "miExtensionService" {
		t.Fatalf("ProjectInfo = %+v %+v %v", info, endpoints, err)
	}
	if _, err := AddTask(dir, info, NewTask(info.Publisher, "nueva", "", "", "", &endpoints[0])); err != nil {
		t.Fatal(err)
	}
	if _, err := AddTask(dir, info, NewTask(info.Publisher, "nueva", "", "", "", nil)); err == nil {
		t.Error("no debe permitir una tarea repetida")
	}
	if issues := Validate(dir); len(issues) > 0 {
		t.Errorf("problemas tras add-task: %+v", issues)
	}
	// La contribución nueva va después de las tareas y antes de la conexión.
	var m struct {
		Contributions []struct{ ID string } `json:"contributions"`
	}
	data, _ := os.ReadFile(filepath.Join(dir, "vss-extension.json"))
	_ = json.Unmarshal(data, &m)
	if len(m.Contributions) != 3 || m.Contributions[1].ID != "nueva-task" || m.Contributions[2].ID != "service-endpoint" {
		t.Errorf("orden de contribuciones: %+v", m.Contributions)
	}
	if !strings.HasPrefix(string(data), "{\n  \"manifestVersion\": 1,") {
		t.Errorf("el manifiesto perdió el orden de claves:\n%.80s", data)
	}

	changes, err := BumpTasks(dir, LevelMinor, []string{"nueva"}, false)
	if err != nil || len(changes) != 1 || changes[0].From != "0.1.0" || changes[0].To != "0.2.0" {
		t.Errorf("BumpTasks = %+v %v", changes, err)
	}
	if _, err := BumpTasks(dir, LevelPatch, []string{"no-existe"}, false); err == nil {
		t.Error("se esperaba error para una tarea inexistente")
	}
	changes, err = BumpConfigs(dir, LevelPatch, []string{"dev"}, false)
	if err != nil || changes[0].To != "0.1.1" {
		t.Errorf("BumpConfigs = %+v %v", changes, err)
	}

	before, _ := os.ReadFile(filepath.Join(dir, "src/tasks/nueva/task.json"))
	changes, err = ResetIDs(dir)
	if err != nil || len(changes) != 2 || changes[0].From == changes[0].To {
		t.Errorf("ResetIDs = %+v %v", changes, err)
	}
	after, _ := os.ReadFile(filepath.Join(dir, "src/tasks/nueva/task.json"))
	if strings.Index(string(before), `"name"`) != strings.Index(string(after), `"name"`) {
		t.Error("ResetIDs no debe reordenar task.json")
	}
}

func TestValidateDetectsDuplicateIDs(t *testing.T) {
	dir := t.TempDir()
	p, tasks, _ := sampleProject()
	tasks[1].UUID = tasks[0].UUID
	tasks[0].Endpoint = nil
	if _, err := Generate(dir, p, tasks, nil); err != nil {
		t.Fatal(err)
	}
	issues := Validate(dir)
	found := false
	for _, i := range issues {
		if strings.Contains(i.Message, "repetido") {
			found = true
		}
	}
	if !found {
		t.Errorf("no detectó el UUID repetido: %+v", issues)
	}
}

func TestHelpers(t *testing.T) {
	if got := Slugify("  Ejecutar Pipeline Ñandú! "); got != "ejecutar-pipeline-nandu" {
		t.Errorf("Slugify = %q", got)
	}
	if Camel("exec-pipeline") != "execPipeline" || Snake("exec-pipeline") != "exec_pipeline" || Title("exec-pipeline") != "Exec Pipeline" {
		t.Error("conversiones de nombre incorrectas")
	}
	for i := 0; i < 50; i++ {
		if u := NewUUID(); !IsUUID(u) || u[14] != '4' {
			t.Fatalf("UUID inválido %s", u)
		}
	}
	if v, _ := BumpVersion("1.2.3", LevelMajor); v != "2.0.0" {
		t.Errorf("BumpVersion = %s", v)
	}
}

func TestObjectKeepsFormatting(t *testing.T) {
	in := "{\n  \"b\": 1,\n  \"a\": [\"x\", \"y\"],\n  \"c\": {\"d\": [1, 2]},\n  \"long\": [\"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\", \"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\"]\n}\n"
	o, err := ParseObject([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	out, _ := o.Bytes()
	if string(out) != in {
		t.Errorf("sin cambios debe quedar idéntico:\n%s", out)
	}
	_ = o.Set("c", map[string]any{"d": []int{1, 2}, "e": []string{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}})
	out, _ = o.Bytes()
	want := "{\n  \"b\": 1,\n  \"a\": [\"x\", \"y\"],\n  \"c\": {\n    \"d\": [1, 2],\n    \"e\": [\n      \"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\",\n      \"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\"\n    ]\n  },\n  \"long\": [\"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\", \"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\"]\n}\n"
	if string(out) != want {
		t.Errorf("got:\n%s\nwant:\n%s", out, want)
	}
}
