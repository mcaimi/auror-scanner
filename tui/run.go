package tui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mcaimi/auror/internal/backend"
	"github.com/mcaimi/auror/internal/config"
	"github.com/mcaimi/auror/internal/flows"
	"github.com/mcaimi/auror/internal/utils"
	"github.com/sirupsen/logrus"
)

// Run starts the TUI, redirects stdout/stderr into the two center panes, runs
// the analysis flow in a background goroutine, and saves the report on success.
func Run(cfg *config.Config, pkgbuildSrc string, log *logrus.Logger) error {
	// Preserve the original file descriptors so bubbletea can render to the
	// real terminal while the analysis output is captured via pipes.
	origStdout := os.Stdout
	origStderr := os.Stderr

	rOut, wOut, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("creating stdout pipe: %w", err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("creating stderr pipe: %w", err)
	}

	os.Stdout = wOut
	os.Stderr = wErr
	log.SetOutput(wErr)

	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
		log.SetOutput(origStderr)
	}()

	pkgDisplay := utils.DerivePkgName(pkgbuildSrc)
	m := newModel(pkgDisplay, cfg.Backend.BaseURL, cfg.Backend.Model)

	p := tea.NewProgram(m,
		tea.WithOutput(origStdout),
		tea.WithAltScreen(),
	)

	// Pipe reader goroutines — forward lines to the TUI as messages.
	go func() {
		scanner := bufio.NewScanner(rOut)
		for scanner.Scan() {
			p.Send(StdoutMsg(scanner.Text()))
		}
	}()

	go func() {
		scanner := bufio.NewScanner(rErr)
		for scanner.Scan() {
			p.Send(StderrMsg(scanner.Text()))
		}
	}()

	// Analysis goroutine — runs the full pipeline and saves the report.
	go func() {
		defer wOut.Close()
		defer wErr.Close()

		ctx := context.Background()

		adapter := backend.OpenAIContext{}
		if initErr := adapter.GetOpenAIAdapter(ctx, cfg, log); initErr != nil {
			p.Send(DoneMsg{Err: initErr})
			return
		}
		adapter.SetCompletionParams()

		flow, flowErr := flows.AurorAnalyticsFlow(&adapter, cfg)
		if flowErr != nil {
			p.Send(DoneMsg{Err: flowErr})
			return
		}

		result, runErr := flow.Run(ctx, flows.AurorPackageInput{PackageName: pkgbuildSrc})
		if runErr != nil {
			p.Send(DoneMsg{Err: runErr})
			return
		}

		if saveErr := saveReport(cfg, pkgbuildSrc, result.Result); saveErr != nil {
			p.Send(DoneMsg{Err: saveErr})
			return
		}

		p.Send(DoneMsg{})
	}()

	_, runErr := p.Run()
	return runErr
}

func saveReport(cfg *config.Config, pkgbuildSrc, content string) error {
	pkgName := utils.DerivePkgName(pkgbuildSrc)
	date := time.Now().Format("2006-01-02")
	reportName := fmt.Sprintf("report-%s-%s.md", pkgName, date)

	if err := os.MkdirAll(cfg.Output.Reports, 0o755); err != nil {
		return fmt.Errorf("creating reports directory: %w", err)
	}

	reportPath := filepath.Join(cfg.Output.Reports, reportName)
	if err := os.WriteFile(reportPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing report: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Report saved to %s\n", reportPath)
	return nil
}
