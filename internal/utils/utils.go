package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/mcaimi/auror/internal/config"
)

// derivePkgName extracts a human-readable package name from a PKGBUILD source,
// which may be an HTTP/HTTPS URL, a filesystem path, or a bare package name.
func DerivePkgName(source string) string {
	u, err := url.Parse(source)
	if err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		if h := u.Query().Get("h"); h != "" {
			return SanitizeName(h)
		}
		// Fall back to the last non-empty path segment
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] != "" && !strings.EqualFold(parts[i], "PKGBUILD") {
				return SanitizeName(parts[i])
			}
		}
	}

	base := filepath.Base(source)
	if strings.EqualFold(base, "PKGBUILD") {
		parent := filepath.Base(filepath.Dir(source))
		if parent != "." && parent != "" {
			return SanitizeName(parent)
		}
	}
	return SanitizeName(base)
}

func SanitizeName(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' {
			b.WriteRune('_')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func FetchPKGBUILD(source string) (string, error) {
	u, err := url.Parse(source)
	if err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		resp, err := http.Get(source) //nolint:gosec // URL is user-supplied and intentional
		if err != nil {
			return "", fmt.Errorf("fetching PKGBUILD from %q: %w", source, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("HTTP %d from %q", resp.StatusCode, source)
		}
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("reading response body: %w", err)
		}
		return string(data), nil
	}

	data, err := os.ReadFile(source)
	if err != nil {
		return "", fmt.Errorf("reading PKGBUILD from %q: %w", source, err)
	}
	return string(data), nil
}

// directly load a skill-file in markdown format from the filesystem
func LoadSkillFile(cfg *config.SkillsConfig) (string, error) {
	path := filepath.Join(cfg.SkillsPath, cfg.SkillFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading skill file %q: %w", path, err)
	}
	return string(data), nil
}
