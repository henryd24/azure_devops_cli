package azdevops

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func newTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c := NewClient("mi org", "mi proyecto", "token")
	c.BaseURL = srv.URL
	c.VSSPSBaseURL = srv.URL
	return c
}

func TestProjectURLEscapes(t *testing.T) {
	c := NewClient("mi org", "mi proyecto", "x")
	got := c.ProjectURL("distributedtask/variablegroups", map[string][]string{"groupName": {"a b&c"}})
	want := "https://dev.azure.com/mi%20org/mi%20proyecto/_apis/distributedtask/variablegroups?groupName=a+b%26c"
	if got != want {
		t.Errorf("ProjectURL = %s, want %s", got, want)
	}
}

func TestDoDecodesAndSendsAuth(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != c0Auth() {
			t.Errorf("Authorization incorrecto: %s", r.Header.Get("Authorization"))
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["k"] != "v" {
			t.Errorf("body = %v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id": 7}`))
	})
	var out struct{ ID int }
	if err := c.Do("POST", c.ProjectURL("x", nil), map[string]string{"k": "v"}, &out); err != nil {
		t.Fatal(err)
	}
	if out.ID != 7 {
		t.Errorf("ID = %d", out.ID)
	}
}

func c0Auth() string { return NewClient("", "", "token").AuthHeader() }

func TestDoParsesAPIError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Variable group not found","typeKey":"NotFound"}`))
	})
	err := c.Do("GET", c.ProjectURL("x", nil), nil, nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("se esperaba APIError, got %v", err)
	}
	if apiErr.StatusCode != 404 || !strings.Contains(err.Error(), "Variable group not found") {
		t.Errorf("error inesperado: %v", err)
	}
	if !IsNotFound(err) {
		t.Error("IsNotFound debería ser true")
	}
}

func TestDoDetectsInvalidPAT(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		w.Write([]byte("<html>login</html>"))
	})
	err := c.Do("GET", c.ProjectURL("x", nil), nil, &struct{}{})
	if err == nil || !strings.Contains(err.Error(), "PAT") {
		t.Fatalf("se esperaba error de autenticación, got %v", err)
	}
}

func TestDoRetriesOn429(t *testing.T) {
	var calls atomic.Int32
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte(`{}`))
	})
	if err := c.Do("POST", c.ProjectURL("x", nil), map[string]int{}, nil); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Errorf("llamadas = %d, se esperaban 2", calls.Load())
	}
}

func TestDoDoesNotRetryPOSTOn500(t *testing.T) {
	var calls atomic.Int32
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	})
	if err := c.Do("POST", c.ProjectURL("x", nil), nil, nil); err == nil {
		t.Fatal("se esperaba error")
	}
	if calls.Load() != 1 {
		t.Errorf("un POST no debe reintentarse ante 500 (llamadas = %d)", calls.Load())
	}
}
