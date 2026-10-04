package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Out recibe los datos (JSON, tablas). Err recibe mensajes de estado.
	Out io.Writer = os.Stdout
	Err io.Writer = os.Stderr

	styleSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	styleError   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	styleWarn    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	styleInfo    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	styleMuted   = lipgloss.NewStyle().Faint(true)
	styleBold    = lipgloss.NewStyle().Bold(true)
)

func Success(format string, a ...any) {
	fmt.Fprintln(Err, styleSuccess.Render("✔ ")+fmt.Sprintf(format, a...))
}

func Errorf(format string, a ...any) {
	fmt.Fprintln(Err, styleError.Render("✖ ")+fmt.Sprintf(format, a...))
}

func Warn(format string, a ...any) {
	fmt.Fprintln(Err, styleWarn.Render("! ")+fmt.Sprintf(format, a...))
}

func Info(format string, a ...any) {
	fmt.Fprintln(Err, styleInfo.Render("› ")+fmt.Sprintf(format, a...))
}

// Link imprime una URL atenuada.
func Link(url string) {
	if url != "" {
		fmt.Fprintln(Err, styleMuted.Render("  🔗 "+url))
	}
}

func Bold(s string) string { return styleBold.Render(s) }

// Status colorea un estado/resultado de ejecución.
func Status(s string) string {
	switch strings.ToLower(s) {
	case "succeeded", "completed", "online", "approved":
		return styleSuccess.Render(s)
	case "failed", "canceled", "cancelled", "offline", "rejected":
		return styleError.Render(s)
	case "partiallysucceeded", "inprogress", "cancelling", "notstarted", "postponed", "pending":
		return styleWarn.Render(s)
	}
	return s
}

// PrintJSON imprime v como JSON indentado en Out.
func PrintJSON(v any) error {
	enc := json.NewEncoder(Out)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// Table imprime una tabla alineada en Out.
func Table(headers []string, rows [][]string) {
	tw := tabwriter.NewWriter(Out, 0, 0, 2, ' ', 0)
	upper := make([]string, len(headers))
	for i, h := range headers {
		upper[i] = strings.ToUpper(h)
	}
	fmt.Fprintln(tw, strings.Join(upper, "\t"))
	for _, r := range rows {
		fmt.Fprintln(tw, strings.Join(r, "\t"))
	}
	tw.Flush()
}

// Progress es un indicador de progreso en stderr cuyo título puede cambiar.
type Progress struct {
	mu    sync.Mutex
	title string
	done  chan struct{}
	wg    sync.WaitGroup
}

// StartProgress inicia un indicador. Si stderr no es una terminal no dibuja nada.
func StartProgress(title string) *Progress {
	p := &Progress{title: title, done: make(chan struct{})}
	if !IsTerminal(os.Stderr) {
		return p
	}
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		t := time.NewTicker(90 * time.Millisecond)
		defer t.Stop()
		for i := 0; ; i++ {
			select {
			case <-p.done:
				fmt.Fprint(Err, "\r\033[K")
				return
			case <-t.C:
				p.mu.Lock()
				title := p.title
				p.mu.Unlock()
				fmt.Fprintf(Err, "\r\033[K%s %s", styleInfo.Render(frames[i%len(frames)]), title)
			}
		}
	}()
	return p
}

// SetTitle cambia el texto mostrado junto al indicador.
func (p *Progress) SetTitle(title string) {
	p.mu.Lock()
	p.title = title
	p.mu.Unlock()
}

// Stop detiene el indicador y limpia la línea.
func (p *Progress) Stop() {
	select {
	case <-p.done:
		return
	default:
		close(p.done)
	}
	p.wg.Wait()
}

// Spinner muestra un indicador de progreso mientras corre fn.
func Spinner(title string, fn func() error) error {
	p := StartProgress(title)
	defer p.Stop()
	return fn()
}

// FormatTime formatea una fecha opcional de forma compacta.
func FormatTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return "-"
	}
	return t.Local().Format("2006-01-02 15:04")
}

// FormatDuration devuelve la duración entre dos fechas opcionales.
func FormatDuration(start, end *time.Time) string {
	if start == nil || start.IsZero() {
		return "-"
	}
	stop := time.Now()
	if end != nil && !end.IsZero() {
		stop = *end
	}
	return stop.Sub(*start).Round(time.Second).String()
}
