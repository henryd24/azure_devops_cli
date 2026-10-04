package securefile

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"azuredevops/azdevops"
	"azuredevops/models"
)

// fakeServer simula los endpoints de secure files necesarios para Replace.
type fakeServer struct {
	mu       sync.Mutex
	files    map[string]models.SecureFile
	content  map[string]string
	nextID   int
	perms    map[string][]int
	roles    map[string][]models.SecurityRoleAssignment
	inherit  map[string]bool
	checks   map[string]int
	failOn   string // fragmento de ruta+método que debe fallar
	requests []string
}

func newFake() *fakeServer {
	return &fakeServer{files: map[string]models.SecureFile{}, content: map[string]string{}, perms: map[string][]int{},
		roles: map[string][]models.SecurityRoleAssignment{}, inherit: map[string]bool{}, checks: map[string]int{}}
}

func (f *fakeServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := r.Method + " " + r.URL.Path
	f.requests = append(f.requests, key)
	if f.failOn != "" && strings.Contains(key, f.failOn) {
		w.WriteHeader(500)
		w.Write([]byte(`{"message":"falla simulada"}`))
		return
	}
	body, _ := io.ReadAll(r.Body)
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	last := parts[len(parts)-1]
	switch {
	case strings.HasSuffix(r.URL.Path, "/distributedtask/securefiles") && r.Method == "POST":
		name := r.URL.Query().Get("name")
		for _, existing := range f.files {
			if existing.Name == name {
				w.WriteHeader(409)
				return
			}
		}
		f.nextID++
		id := "new-" + string(rune('0'+f.nextID))
		f.files[id] = models.SecureFile{ID: id, Name: name}
		f.content[id] = string(body)
		json.NewEncoder(w).Encode(f.files[id])
	case strings.Contains(r.URL.Path, "/distributedtask/securefiles/") && r.Method == "PATCH":
		var sf models.SecureFile
		json.Unmarshal(body, &sf)
		f.files[last] = sf
		json.NewEncoder(w).Encode(sf)
	case strings.Contains(r.URL.Path, "/distributedtask/securefiles/") && r.Method == "DELETE":
		delete(f.files, last)
	case strings.Contains(r.URL.Path, "/pipelinepermissions/securefile/") && r.Method == "PATCH":
		var p struct{ Pipelines []struct{ ID int } }
		json.Unmarshal(body, &p)
		for _, pl := range p.Pipelines {
			f.perms[last] = append(f.perms[last], pl.ID)
		}
		w.Write([]byte(`{}`))
	case strings.Contains(r.URL.Path, "/roleassignments/") && r.Method == "PUT":
		var roles []models.SecurityRoleAssignment
		json.Unmarshal(body, &roles)
		id := last[strings.Index(last, "$")+1:]
		f.roles[id] = append(f.roles[id], roles...)
	case strings.Contains(r.URL.Path, "/roleassignments/") && r.Method == "PATCH":
		id := last[strings.Index(last, "$")+1:]
		f.inherit[id] = r.URL.Query().Get("inheritPermissions") == "true"
	case strings.HasSuffix(r.URL.Path, "/checks/configurations") && r.Method == "POST":
		var c models.CheckConfiguration
		json.Unmarshal(body, &c)
		f.checks[c.Resource.ID]++
		w.Write([]byte(`{}`))
	default:
		w.WriteHeader(404)
	}
}

func setup(t *testing.T) (*fakeServer, *azdevops.Client, *Snapshot) {
	fake := newFake()
	fake.files["old"] = models.SecureFile{ID: "old", Name: "cert.pfx", Properties: map[string]string{"env": "prod"}}
	srv := httptest.NewServer(fake)
	t.Cleanup(srv.Close)
	client := azdevops.NewClient("org", "proj", "pat")
	client.BaseURL = srv.URL
	client.MaxRetries = 0

	snap := &Snapshot{File: fake.files["old"]}
	snap.Permissions.Pipelines = []models.PipelineAuthorization{{ID: 61, Authorized: true}, {ID: 62, Authorized: false}}
	user := models.ResourceRoleAssignment{Identity: models.IdentityRef{ID: "u1"}, Access: "assigned"}
	user.Role.Name = "User"
	inherited := models.ResourceRoleAssignment{Identity: models.IdentityRef{ID: "g1"}, Access: "inherited"}
	inherited.Role.Name = "Reader"
	snap.Roles = []models.ResourceRoleAssignment{user, inherited}
	var check models.CheckConfiguration
	check.Type.Name = "Approval"
	check.Settings = json.RawMessage(`{"approvers":[{"id":"u1"}]}`)
	snap.Checks = []models.CheckConfiguration{check}
	return fake, client, snap
}

func TestReplacePreservesAccess(t *testing.T) {
	fake, client, snap := setup(t)
	res, err := Replace(client, "pid", snap, []byte("nuevo"), ReplaceOptions{})
	if err != nil {
		t.Fatal(err)
	}
	newID := res.New.ID
	if _, ok := fake.files["old"]; ok {
		t.Error("el archivo anterior debería haberse eliminado")
	}
	got := fake.files[newID]
	if got.Name != "cert.pfx" || got.Properties["env"] != "prod" || fake.content[newID] != "nuevo" {
		t.Errorf("archivo nuevo incorrecto: %+v contenido=%q", got, fake.content[newID])
	}
	if len(fake.perms[newID]) != 1 || fake.perms[newID][0] != 61 {
		t.Errorf("pipelines autorizados = %v, se esperaba [61]", fake.perms[newID])
	}
	if len(fake.roles[newID]) != 1 || fake.roles[newID][0].UserID != "u1" || fake.roles[newID][0].RoleName != "User" {
		t.Errorf("roles = %+v", fake.roles[newID])
	}
	if _, touched := fake.inherit[newID]; touched {
		t.Error("con roles heredados no se debe tocar la herencia")
	}
	if fake.checks[newID] != 1 || res.Checks != 1 {
		t.Errorf("checks copiados = %d", fake.checks[newID])
	}
	if !res.OldDeleted {
		t.Error("OldDeleted debería ser true")
	}
}

func TestReplaceSkipsSelfRole(t *testing.T) {
	fake, client, snap := setup(t)
	self := models.ResourceRoleAssignment{Identity: models.IdentityRef{ID: "me"}, Access: "assigned"}
	self.Role.Name = "Reader"
	snap.Roles = append(snap.Roles, self)
	res, err := Replace(client, "pid", snap, []byte("x"), ReplaceOptions{SelfID: "ME"})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range fake.roles[res.New.ID] {
		if r.UserID == "me" {
			t.Error("no se debe enviar el rol de la propia identidad")
		}
	}
	if res.Roles != 2 || len(res.Warnings) != 1 {
		t.Errorf("roles=%d warnings=%v", res.Roles, res.Warnings)
	}
}

func TestReplaceDisablesInheritance(t *testing.T) {
	fake, client, snap := setup(t)
	snap.Roles = snap.Roles[:1] // solo asignados => herencia desactivada
	res, err := Replace(client, "pid", snap, []byte("x"), ReplaceOptions{KeepOld: true})
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := fake.inherit[res.New.ID]; !ok || v {
		t.Error("debería desactivar la herencia en el archivo nuevo")
	}
	if old, ok := fake.files["old"]; !ok || !strings.HasPrefix(old.Name, "cert.pfx.replaced-") {
		t.Errorf("con KeepOld el anterior debe quedar renombrado: %+v", old)
	}
}

func TestReplaceRollsBackOnFailure(t *testing.T) {
	fake, client, snap := setup(t)
	fake.failOn = "POST /org/proj/_apis/pipelines/checks"
	_, err := Replace(client, "pid", snap, []byte("x"), ReplaceOptions{})
	if err == nil || !strings.Contains(err.Error(), "revertido") {
		t.Fatalf("se esperaba error de reemplazo revertido, got %v", err)
	}
	if len(fake.files) != 1 || fake.files["old"].Name != "cert.pfx" {
		t.Errorf("tras revertir solo debe quedar el original con su nombre: %+v", fake.files)
	}
}

func TestReplaceUploadFailureRestoresName(t *testing.T) {
	fake, client, snap := setup(t)
	fake.failOn = "POST /org/proj/_apis/distributedtask/securefiles"
	if _, err := Replace(client, "pid", snap, []byte("x"), ReplaceOptions{}); err == nil {
		t.Fatal("se esperaba error")
	}
	if fake.files["old"].Name != "cert.pfx" {
		t.Errorf("el nombre original no se restauró: %+v", fake.files["old"])
	}
}
