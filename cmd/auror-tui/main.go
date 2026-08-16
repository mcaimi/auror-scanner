package main

import (
	"os"

	"github.com/mcaimi/auror/internal/config"
	"github.com/mcaimi/auror/tui"
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
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stderr)
	log.SetLevel(logrus.InfoLevel)
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "auror-tui",
		Short: "AUROR TUI - AI-powered AUR Package Build Analyzer",
		Long:  "An AI-powered security analyzer for PKGBUILDs with a terminal UI.",
		Run:   run,
	}

	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to configuration file")
	rootCmd.Flags().StringVarP(&pkgbuildSrc, "pkgfile", "p", "", "PKGBUILD URL or local path to analyze")
	_ = rootCmd.MarkFlagRequired("pkgfile")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func run(_ *cobra.Command, _ []string) {
	cfg, err := config.Load(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := tui.Run(cfg, pkgbuildSrc, log); err != nil {
		log.Fatalf("TUI error: %v", err)
	}
}
