package cmd

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"azuredevops/internal/ui"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

// Output devuelve el formato de salida pedido con -o, o def si no se indicó.
// Con --query y sin -o explícito se usa json en lugar de table.
func Output(def string) (string, error) {
	out := strings.ToLower(globalFlags.output)
	if out == "" {
		out = def
		if out == "table" && globalFlags.query != "" {
			out = "json"
		}
	}
	if out != "table" && !slices.Contains(ui.DataFormats, out) {
		return "", fmt.Errorf("formato de salida '%s' no soportado (usa table, json, yaml o tsv)", out)
	}
	return out, nil
}

// Print imprime datos en el formato pedido (json por defecto), aplicando --query.
func Print(v any) error {
	out, err := Output("json")
	if err != nil {
		return err
	}
	if out == "table" {
		out = "json"
	}
	return ui.Render(ui.Out, v, out, globalFlags.query)
}

// WantsData indica si el usuario pidió explícitamente datos (-o json/yaml/tsv o --query),
// útil en comandos de acción que normalmente solo muestran mensajes.
func WantsData() bool {
	o := strings.ToLower(globalFlags.output)
	return globalFlags.query != "" || (o != "" && o != "table")
}

// ParseVariables interpreta entradas "clave=valor" o "secret:clave=valor".
func ParseVariables(entries []string) (map[string]models.VariableVal, error) {
	vars := make(map[string]models.VariableVal, len(entries))
	for _, entry := range entries {
		key, value, secret, err := parseEntry(entry)
		if err != nil {
			return nil, err
		}
		vars[key] = models.VariableVal{Value: value, IsSecret: secret}
	}
	return vars, nil
}

// ParseBuildVariables es como ParseVariables pero para variables de ejecución de pipelines.
func ParseBuildVariables(entries []string) (map[string]models.BuildVariable, error) {
	vars := make(map[string]models.BuildVariable, len(entries))
	for _, entry := range entries {
		key, value, secret, err := parseEntry(entry)
		if err != nil {
			return nil, err
		}
		vars[key] = models.BuildVariable{Value: value, IsSecret: secret}
	}
	return vars, nil
}

// ParseParams interpreta entradas "clave=valor".
func ParseParams(entries []string) (map[string]string, error) {
	params := make(map[string]string, len(entries))
	for _, entry := range entries {
		key, value, ok := strings.Cut(entry, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			return nil, fmt.Errorf("parámetro inválido '%s': usa clave=valor", entry)
		}
		params[key] = value
	}
	return params, nil
}

func parseEntry(entry string) (key, value string, secret bool, err error) {
	raw := entry
	if rest, ok := strings.CutPrefix(raw, "secret:"); ok {
		secret = true
		raw = rest
	}
	key, value, ok := strings.Cut(raw, "=")
	key = strings.TrimSpace(key)
	if !ok {
		return "", "", false, fmt.Errorf("formato de variable inválido '%s': usa clave=valor o secret:clave=valor", entry)
	}
	if key == "" {
		return "", "", false, fmt.Errorf("la clave no puede estar vacía en '%s'", entry)
	}
	return key, strings.TrimSpace(value), secret, nil
}

// Confirm pide confirmación salvo que yes sea true. Sin terminal exige --yes.
func Confirm(message string, yes bool) error {
	if yes {
		return nil
	}
	if !ui.Interactive() {
		return fmt.Errorf("%s: usa --yes para confirmar en modo no interactivo", message)
	}
	ok, err := ui.Confirm(message, false)
	if err != nil {
		return err
	}
	if !ok {
		return ui.ErrAborted
	}
	return nil
}

// PromptVariables pide variables una a una en modo interactivo.
func PromptVariables() (map[string]models.VariableVal, error) {
	vars := map[string]models.VariableVal{}
	for {
		key, err := ui.Input("Nombre de la variable", "Déjalo vacío para terminar", "", nil)
		if err != nil {
			return nil, err
		}
		if key == "" {
			return vars, nil
		}
		secret, err := ui.Confirm(fmt.Sprintf("¿'%s' es secreta?", key), false)
		if err != nil {
			return nil, err
		}
		var value string
		if secret {
			value, err = ui.Password("Valor de "+key, "No se mostrará en pantalla")
		} else {
			value, err = ui.Input("Valor de "+key, "", "", nil)
		}
		if err != nil {
			return nil, err
		}
		vars[key] = models.VariableVal{Value: value, IsSecret: secret}
	}
}

// PositiveInt valida que un texto sea un entero positivo (para ui.Input).
func PositiveInt(s string) error {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n <= 0 {
		return fmt.Errorf("debe ser un número entero positivo")
	}
	return nil
}

// NoFileComp desactiva el autocompletado de archivos para un flag.
func NoFileComp(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}
