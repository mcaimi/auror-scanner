package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mcaimi/auror/internal/backend"
	"github.com/mcaimi/auror/internal/config"
	"github.com/mcaimi/auror/internal/flows"
	"github.com/mcaimi/auror/internal/utils"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	configFile  string
	pkgbuildSrc string
	log         *logrus.Logger
)

func init() {
	log = logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(logrus.InfoLevel)
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "auror",
		Short: "AUROR - AI-powered AUR Package Build Analyzer",
		Long:  "An AI-powered security analyzer for PKGBUILDs hosted @aur.archlinux.org.",
		Run:   run,
	}

	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to configuration file")
	rootCmd.Flags().StringVarP(&pkgbuildSrc, "pkgbuild", "p", "", "PKGBUILD URL or local path to analyze")
	_ = rootCmd.MarkFlagRequired("pkgbuild")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func run(c *cobra.Command, args []string) {
	cfg, err := config.Load(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Info("Starting AUROR...")
	ctx := context.Background()

	// create openai adapter
	backendAdapter := backend.OpenAIContext{}
	if err := backendAdapter.GetOpenAIAdapter(ctx, cfg, log); err != nil {
		log.Fatalf("Failed to initialize backend: %v", err)
	}

	backendAdapter.SetCompletionParams()

	flow, err := flows.AurorAnalyticsFlow(&backendAdapter, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize analysis flow: %v", err)
	}

	result, err := flow.Run(ctx, flows.AurorPackageInput{PackageName: pkgbuildSrc})
	if err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}

	log.Info("Analysis complete")

	pkgName := utils.DerivePkgName(pkgbuildSrc)
	date := time.Now().Format("2006-01-02")
	reportName := fmt.Sprintf("report-%s-%s.md", pkgName, date)

	reportsDir := cfg.Output.Reports
	if err := os.MkdirAll(reportsDir, 0o755); err != nil {
		log.Fatalf("Failed to create reports directory %q: %v", reportsDir, err)
	}

	reportPath := filepath.Join(reportsDir, reportName)
	if err := os.WriteFile(reportPath, []byte(result.Result), 0o644); err != nil {
		log.Fatalf("Failed to write report to %q: %v", reportPath, err)
	}

	log.Infof("Report saved to %s", reportPath)
}
