package ui

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// OpenBrowser abre url en el navegador. Respeta la variable BROWSER.
func OpenBrowser(url string) error {
	if b := os.Getenv("BROWSER"); b != "" {
		parts := strings.Fields(b)
		return exec.Command(parts[0], append(parts[1:], url)...).Start()
	}
	var c *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		c = exec.Command("open", url)
	case "windows":
		c = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		c = exec.Command("xdg-open", url)
	}
	return c.Start()
}
