package azdevops

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultBaseURL      = "https://dev.azure.com"
	DefaultVSSPSBaseURL = "https://vssps.dev.azure.com"
)

// Version se inyecta en tiempo de compilación (ver Makefile) y se usa en el User-Agent.
var Version = "dev"

type Client struct {
	Org     string
	Project string
	PAT     string
	HTTP    *http.Client

	// BaseURL y VSSPSBaseURL permiten apuntar a otro host (por ejemplo en pruebas).
	BaseURL      string
	VSSPSBaseURL string

	// MaxRetries es el número de reintentos ante 429/5xx.
	MaxRetries int
	// Debug imprime cada petición en stderr.
	Debug bool
}

func NewClient(org, project, pat string) *Client {
	return &Client{
		Org:          org,
		Project:      project,
		PAT:          pat,
		HTTP:         &http.Client{Timeout: 30 * time.Second},
		BaseURL:      DefaultBaseURL,
		VSSPSBaseURL: DefaultVSSPSBaseURL,
		MaxRetries:   3,
	}
}

// GetClientFromEnv construye un cliente a partir de AZURE_ORG, AZURE_PROJECT y AZURE_PAT.
func GetClientFromEnv() (*Client, error) {
	org := os.Getenv("AZURE_ORG")
	project := os.Getenv("AZURE_PROJECT")
	pat := os.Getenv("AZURE_PAT")

	if org == "" || project == "" || pat == "" {
		return nil, errors.New("AZURE_ORG, AZURE_PROJECT y AZURE_PAT deben estar definidos en las variables de entorno")
	}

	return NewClient(org, project, pat), nil
}

func (c *Client) AuthHeader() string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(":"+c.PAT))
}

// ProjectURL arma https://dev.azure.com/{org}/{project}/_apis/{path}?{query}.
func (c *Client) ProjectURL(path string, query url.Values) string {
	return buildURL(fmt.Sprintf("%s/%s/%s/_apis/%s", c.BaseURL, url.PathEscape(c.Org), url.PathEscape(c.Project), path), query)
}

// OrgURL arma https://dev.azure.com/{org}/_apis/{path}?{query}.
func (c *Client) OrgURL(path string, query url.Values) string {
	return buildURL(fmt.Sprintf("%s/%s/_apis/%s", c.BaseURL, url.PathEscape(c.Org), path), query)
}

// VSSPSURL arma https://vssps.dev.azure.com/{org}/_apis/{path}?{query}.
func (c *Client) VSSPSURL(path string, query url.Values) string {
	return buildURL(fmt.Sprintf("%s/%s/_apis/%s", c.VSSPSBaseURL, url.PathEscape(c.Org), path), query)
}

// WebURL arma una URL del portal web del proyecto, p. ej. "_build?definitionId=1".
func (c *Client) WebURL(path string) string {
	return fmt.Sprintf("%s/%s/%s/%s", c.BaseURL, url.PathEscape(c.Org), url.PathEscape(c.Project), path)
}

// VariableGroupWebURL devuelve el enlace del portal a un Variable Group.
func (c *Client) VariableGroupWebURL(id int) string {
	return c.WebURL(fmt.Sprintf("_library?itemType=variableGroups&view=VariableGroupView&variableGroupId=%d", id))
}

func buildURL(base string, query url.Values) string {
	if len(query) == 0 {
		return base
	}
	return base + "?" + query.Encode()
}

// APIError representa una respuesta de error de la API de Azure DevOps.
type APIError struct {
	StatusCode int
	Message    string
	TypeKey    string
	Method     string
	URL        string
}

func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = http.StatusText(e.StatusCode)
	}
	return fmt.Sprintf("Azure DevOps respondió %d: %s", e.StatusCode, msg)
}

// IsNotFound indica si el error es un 404 de la API.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// Do ejecuta una petición contra la API. body (si no es nil) se serializa a JSON;
// out (si no es nil) recibe la respuesta decodificada. Un *[]byte o *string en out
// recibe el cuerpo sin decodificar.
func (c *Client) Do(method, rawURL string, body, out any) error {
	return c.DoContentType(method, rawURL, "application/json", body, out)
}

// DoContentType es como Do pero permite otro Content-Type (p. ej. application/json-patch+json).
// Si body es []byte se envía tal cual (p. ej. un archivo con application/octet-stream).
func (c *Client) DoContentType(method, rawURL, contentType string, body, out any) error {
	var payload []byte
	if raw, ok := body.([]byte); ok {
		payload = raw
	} else if body != nil {
		var err error
		if payload, err = json.Marshal(body); err != nil {
			return fmt.Errorf("error al serializar el payload: %w", err)
		}
	}

	var resp *http.Response
	var respBody []byte
	for attempt := 0; ; attempt++ {
		req, err := http.NewRequest(method, rawURL, bytes.NewReader(payload))
		if err != nil {
			return fmt.Errorf("petición inválida: %w", err)
		}
		req.Header.Set("Authorization", c.AuthHeader())
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "azdevops-cli/"+Version)
		if body != nil {
			req.Header.Set("Content-Type", contentType)
		}

		start := time.Now()
		resp, err = c.HTTP.Do(req)
		if err != nil {
			if attempt < c.MaxRetries && isIdempotent(method) {
				c.debugf("%s %s -> error de red (%v), reintentando", method, rawURL, err)
				time.Sleep(backoff(attempt, nil))
				continue
			}
			return fmt.Errorf("la solicitud %s %s falló: %w", method, rawURL, err)
		}
		respBody, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("error al leer la respuesta: %w", err)
		}
		c.debugf("%s %s -> %d (%s)", method, rawURL, resp.StatusCode, time.Since(start).Round(time.Millisecond))

		if shouldRetry(method, resp.StatusCode) && attempt < c.MaxRetries {
			time.Sleep(backoff(attempt, resp))
			continue
		}
		break
	}

	if err := checkResponse(method, rawURL, resp, respBody); err != nil {
		return err
	}

	if out == nil || len(respBody) == 0 {
		return nil
	}
	switch o := out.(type) {
	case *[]byte:
		*o = respBody
		return nil
	case *string:
		*o = string(respBody)
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("error al decodificar la respuesta de %s: %w", rawURL, err)
	}
	return nil
}

// GetWithHeaders hace un GET como Do pero devuelve también las cabeceras (paginación).
func (c *Client) GetWithHeaders(rawURL string, out any) (http.Header, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.AuthHeader())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "azdevops-cli/"+Version)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("la solicitud GET %s falló: %w", rawURL, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	c.debugf("GET %s -> %d", rawURL, resp.StatusCode)
	if err := checkResponse(http.MethodGet, rawURL, resp, body); err != nil {
		return nil, err
	}
	if out != nil && len(body) > 0 {
		if err := json.Unmarshal(body, out); err != nil {
			return nil, fmt.Errorf("error al decodificar la respuesta: %w", err)
		}
	}
	return resp.Header, nil
}

func checkResponse(method, rawURL string, resp *http.Response, body []byte) error {
	// Con un PAT inválido Azure DevOps responde 203 con la página HTML de login.
	if resp.StatusCode == http.StatusNonAuthoritativeInfo ||
		(resp.StatusCode < 300 && strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html")) {
		return &APIError{StatusCode: http.StatusUnauthorized, Method: method, URL: rawURL,
			Message: "autenticación fallida: el PAT es inválido, expiró o no tiene acceso a la organización"}
	}
	if resp.StatusCode < 300 {
		return nil
	}

	apiErr := &APIError{StatusCode: resp.StatusCode, Method: method, URL: rawURL}
	var parsed struct {
		Message string `json:"message"`
		TypeKey string `json:"typeKey"`
	}
	if json.Unmarshal(body, &parsed) == nil && parsed.Message != "" {
		apiErr.Message = parsed.Message
		apiErr.TypeKey = parsed.TypeKey
	} else {
		apiErr.Message = strings.TrimSpace(truncate(string(body), 300))
	}
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		apiErr.Message = "no autorizado: revisa tu PAT (" + apiErr.Message + ")"
	case http.StatusForbidden:
		apiErr.Message = "acceso denegado: el PAT no tiene los permisos/scopes necesarios (" + apiErr.Message + ")"
	}
	return apiErr
}

func isIdempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions:
		return true
	}
	return false
}

func shouldRetry(method string, status int) bool {
	if status == http.StatusTooManyRequests {
		return true
	}
	return status >= 500 && isIdempotent(method)
}

func backoff(attempt int, resp *http.Response) time.Duration {
	if resp != nil {
		if s := resp.Header.Get("Retry-After"); s != "" {
			if secs, err := strconv.Atoi(s); err == nil && secs >= 0 {
				return time.Duration(secs) * time.Second
			}
		}
	}
	return time.Duration(1<<attempt) * 500 * time.Millisecond
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func (c *Client) debugf(format string, args ...any) {
	if c.Debug {
		fmt.Fprintf(os.Stderr, "[debug] "+format+"\n", args...)
	}
}

// WithProject devuelve una copia del cliente apuntando a otro proyecto de la misma organización.
func (c *Client) WithProject(project string) *Client {
	cc := *c
	cc.Project = project
	return &cc
}
