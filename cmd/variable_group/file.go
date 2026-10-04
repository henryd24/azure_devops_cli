package variable_group

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"azuredevops/models"
)

// parseEnvFile interpreta un archivo .env: KEY=VALUE, comentarios con '#',
// prefijo opcional "export " y valores entre comillas simples o dobles.
func parseEnvFile(data []byte) (map[string]models.VariableVal, error) {
	vars := map[string]models.VariableVal{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		text = strings.TrimPrefix(text, "export ")
		key, value, ok := strings.Cut(text, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			return nil, fmt.Errorf("línea %d inválida: se esperaba CLAVE=VALOR", line)
		}
		value = strings.TrimSpace(value)
		if len(value) >= 2 {
			switch {
			case value[0] == '"' && value[len(value)-1] == '"':
				if unq, err := strconv.Unquote(value); err == nil {
					value = unq
				} else {
					value = value[1 : len(value)-1]
				}
			case value[0] == '\'' && value[len(value)-1] == '\'':
				value = value[1 : len(value)-1]
			}
		}
		vars[key] = models.VariableVal{Value: value}
	}
	return vars, scanner.Err()
}

// parseJSONFile acepta {"k": "v"} o {"k": {"value": "v", "isSecret": true}}.
func parseJSONFile(data []byte) (map[string]models.VariableVal, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("JSON inválido: %w", err)
	}
	vars := make(map[string]models.VariableVal, len(raw))
	for k, v := range raw {
		var s string
		if err := json.Unmarshal(v, &s); err == nil {
			vars[k] = models.VariableVal{Value: s}
			continue
		}
		var vv models.VariableVal
		if err := json.Unmarshal(v, &vv); err != nil {
			return nil, fmt.Errorf("valor inválido para '%s': se esperaba texto u objeto {value, isSecret}", k)
		}
		vars[k] = vv
	}
	return vars, nil
}

// formatEnv genera un archivo .env. Las variables secretas no tienen valor
// legible desde la API y se emiten comentadas.
func formatEnv(vars map[string]models.VariableVal) []byte {
	var b strings.Builder
	for _, k := range sortedKeys(vars) {
		v := vars[k]
		if v.IsSecret {
			fmt.Fprintf(&b, "# %s es secreta: la API no devuelve su valor\n# %s=\n", k, k)
			continue
		}
		value := v.Value
		if strings.ContainsAny(value, " \t#\"'\n") {
			value = strconv.Quote(value)
		}
		fmt.Fprintf(&b, "%s=%s\n", k, value)
	}
	return []byte(b.String())
}

func sortedKeys(vars map[string]models.VariableVal) []string {
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
