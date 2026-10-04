// Package ui agrupa los prompts interactivos y el formateo de salida de la CLI.
package ui

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/mattn/go-isatty"
)

// ErrAborted se devuelve cuando el usuario cancela un prompt (Ctrl+C / Esc).
var ErrAborted = errors.New("operación cancelada por el usuario")

// interactive indica si se permite preguntar al usuario. Se calcula en SetInteractive.
var interactive = false

// SetInteractive habilita los prompts solo si stdin y stdout son terminales.
func SetInteractive(allowed bool) {
	interactive = allowed && IsTerminal(os.Stdin) && IsTerminal(os.Stdout)
}

// Interactive indica si se pueden mostrar prompts.
func Interactive() bool { return interactive }

// IsTerminal indica si el archivo es una terminal.
func IsTerminal(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}

// MissingError construye el error que se devuelve cuando falta un dato y no se puede preguntar.
func MissingError(flag string) error {
	return fmt.Errorf("falta el flag --%s (o ejecútalo en una terminal para que se te pregunte)", flag)
}

func wrap(err error) error {
	if errors.Is(err, huh.ErrUserAborted) {
		return ErrAborted
	}
	return err
}

// Required valida que un texto no esté vacío.
func Required(s string) error {
	if strings.TrimSpace(s) == "" {
		return errors.New("este campo es obligatorio")
	}
	return nil
}

// Input pide un texto. value trae el valor por defecto. validate puede ser nil.
func Input(title, description, value string, validate func(string) error) (string, error) {
	in := huh.NewInput().Title(title).Description(description).Value(&value)
	if validate != nil {
		in = in.Validate(validate)
	}
	if err := in.Run(); err != nil {
		return "", wrap(err)
	}
	return strings.TrimSpace(value), nil
}

// Password pide un texto sin mostrarlo en pantalla.
func Password(title, description string) (string, error) {
	var value string
	err := huh.NewInput().Title(title).Description(description).
		EchoMode(huh.EchoModePassword).Validate(Required).Value(&value).Run()
	return strings.TrimSpace(value), wrap(err)
}

// Confirm pide una confirmación sí/no.
func Confirm(title string, def bool) (bool, error) {
	value := def
	err := huh.NewConfirm().Title(title).Affirmative("Sí").Negative("No").Value(&value).Run()
	return value, wrap(err)
}

// Option es una opción de Select/MultiSelect.
type Option[T comparable] struct {
	Label string
	Value T
}

func toHuh[T comparable](opts []Option[T]) []huh.Option[T] {
	out := make([]huh.Option[T], len(opts))
	for i, o := range opts {
		out[i] = huh.NewOption(o.Label, o.Value)
	}
	return out
}

// Select muestra una lista con filtro (tecla "/") y devuelve el valor elegido.
func Select[T comparable](title string, opts []Option[T]) (T, error) {
	var value T
	if len(opts) == 0 {
		return value, errors.New("no hay opciones para elegir")
	}
	err := huh.NewSelect[T]().Title(title).Options(toHuh(opts)...).
		Filtering(len(opts) > 8).Height(min(len(opts)+2, 15)).Value(&value).Run()
	return value, wrap(err)
}

// MultiSelect permite elegir varias opciones (espacio para marcar, enter para confirmar).
func MultiSelect[T comparable](title string, opts []Option[T]) ([]T, error) {
	if len(opts) == 0 {
		return nil, errors.New("no hay opciones para elegir")
	}
	var values []T
	err := huh.NewMultiSelect[T]().Title(title).
		Description("espacio: marcar · /: filtrar · enter: confirmar").
		Options(toHuh(opts)...).Filterable(true).Height(min(len(opts)+3, 15)).
		Validate(func(v []T) error {
			if len(v) == 0 {
				return errors.New("elige al menos una opción")
			}
			return nil
		}).Value(&values).Run()
	return values, wrap(err)
}
