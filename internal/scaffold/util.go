// Package scaffold genera proyectos de extensiones de Azure DevOps con la estructura
// de tareas en src/tasks, empaquetado con ncc y configuraciones dev/release.
package scaffold

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// NewUUID devuelve un UUID v4 aleatorio.
func NewUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

var (
	uuidRe    = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	nonSlugRe = regexp.MustCompile(`[^a-z0-9]+`)
	slugRe    = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	semverRe  = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
)

// IsUUID indica si s tiene formato de UUID.
func IsUUID(s string) bool { return uuidRe.MatchString(s) }

// Slugify convierte un texto en kebab-case ASCII ("Ejecutar Pipeline" -> "ejecutar-pipeline").
func Slugify(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	ascii, _, _ := transform.String(t, s)
	return strings.Trim(nonSlugRe.ReplaceAllString(strings.ToLower(ascii), "-"), "-")
}

// IsSlug valida un identificador en kebab-case.
func IsSlug(s string) bool { return slugRe.MatchString(s) }

// IsSemver valida una versión X.Y.Z.
func IsSemver(s string) bool { return semverRe.MatchString(s) }

// Snake convierte kebab-case en snake_case ("exec-pipeline" -> "exec_pipeline").
func Snake(slug string) string { return strings.ReplaceAll(slug, "-", "_") }

// Camel convierte kebab-case en camelCase ("exec-pipeline" -> "execPipeline").
func Camel(slug string) string {
	parts := strings.Split(slug, "-")
	for i := 1; i < len(parts); i++ {
		if parts[i] != "" {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// Title convierte kebab-case en un título ("exec-pipeline" -> "Exec Pipeline").
func Title(slug string) string {
	parts := strings.Split(slug, "-")
	for i, p := range parts {
		if p != "" {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

// EnvName convierte un nombre de input al formato de variable de entorno del agente.
func EnvName(s string) string {
	return strings.ToUpper(strings.NewReplacer(".", "_", " ", "_", "-", "_").Replace(s))
}
