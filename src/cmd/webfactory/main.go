package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"webfactory/src/internal/builder"

	"github.com/LixenWraith/logger/quick"
)

type buildConfig struct {
	sourcePath string
	targetPath string
	logPath    string
}

func main() {
	cfg := processCLI()

	fmt.Println("Source directory: ", cfg.sourcePath)
	fmt.Println("Target directory: ", cfg.targetPath)

	quick.Info("Starting site build", "source path", "target path", cfg.targetPath)

	builder := builder.New(cfg.sourcePath, cfg.targetPath)

	if err := builder.Build(); err != nil {
		quick.Error("Error building site", "error", err.Error())
		fmt.Fprintf(os.Stderr, "Error building site: %v\n", err)
		os.Exit(1)
	}

	quick.Info("Site build completed successfully")
	quick.Shutdown()
	time.Sleep(300 * time.Millisecond)
}

func processCLI() *buildConfig {
	cfg := &buildConfig{}

	flag.StringVar(&cfg.targetPath, "t", ".", "Output directory path")
	flag.StringVar(&cfg.sourcePath, "s", ".", "Source blueprints and components path")
	flag.StringVar(&cfg.logPath, "l", "logs", "Log directory path")
	flag.Parse()

	// Clean and make absolute paths
	var err error
	cfg.sourcePath, err = filepath.Abs(filepath.Clean(cfg.sourcePath))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing source path: %v\n", err)
		os.Exit(1)
	}

	cfg.targetPath, err = filepath.Abs(filepath.Clean(cfg.targetPath))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing output path: %v\n", err)
		os.Exit(1)
	}

	cfg.logPath, err = filepath.Abs(filepath.Clean(cfg.logPath))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing log path: %v\n", err)
		os.Exit(1)
	}

	// Verify source directory exists
	if _, err := os.Stat(cfg.sourcePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Source directory does not exist: %s\n", cfg.sourcePath)
		os.Exit(1)
	}

	return cfg
}