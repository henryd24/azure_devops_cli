package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/jmespath/go-jmespath"
	"gopkg.in/yaml.v3"
)

// DataFormats son los formatos de salida para datos (además de "table").
var DataFormats = []string{"json", "yaml", "tsv"}

// Render escribe v en w con el formato indicado (json, yaml o tsv), aplicando
// antes la consulta JMESPath query si no está vacía.
func Render(w io.Writer, v any, format, query string) error {
	if format == "json" && query == "" {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		return enc.Encode(v)
	}

	data, err := toGeneric(v)
	if err != nil {
		return err
	}
	if query != "" {
		// JMESPath solo compara números float64.
		if data, err = jmespath.Search(query, data); err != nil {
			return fmt.Errorf("consulta --query inválida: %w", err)
		}
	}
	data = normalizeNumbers(data)

	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		return enc.Encode(data)
	case "yaml":
		enc := yaml.NewEncoder(w)
		enc.SetIndent(2)
		if err := enc.Encode(data); err != nil {
			return err
		}
		return enc.Close()
	case "tsv":
		return writeTSV(w, data)
	}
	return fmt.Errorf("formato de salida '%s' no soportado", format)
}

// toGeneric convierte v a mapas/slices genéricos respetando las etiquetas json.
func toGeneric(v any) (any, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// normalizeNumbers convierte los float64 enteros en int64 para que YAML/TSV no
// los muestren con decimales o notación científica.
func normalizeNumbers(v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			t[k] = normalizeNumbers(val)
		}
	case []any:
		for i, val := range t {
			t[i] = normalizeNumbers(val)
		}
	case float64:
		if t == float64(int64(t)) {
			return int64(t)
		}
	}
	return v
}

func writeTSV(w io.Writer, data any) error {
	rows, ok := data.([]any)
	if !ok {
		rows = []any{data}
	}
	for _, row := range rows {
		if _, err := fmt.Fprintln(w, tsvLine(row)); err != nil {
			return err
		}
	}
	return nil
}

func tsvLine(v any) string {
	switch t := v.(type) {
	case []any:
		parts := make([]string, len(t))
		for i, item := range t {
			parts[i] = tsvValue(item)
		}
		return strings.Join(parts, "\t")
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, len(keys))
		for i, k := range keys {
			parts[i] = tsvValue(t[k])
		}
		return strings.Join(parts, "\t")
	}
	return tsvValue(v)
}

func tsvValue(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return strings.NewReplacer("\t", " ", "\n", " ").Replace(t)
	case map[string]any, []any:
		b, _ := json.Marshal(t)
		return string(b)
	}
	return fmt.Sprint(v)
}
