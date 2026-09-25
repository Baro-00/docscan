package main

import (
	"context"
	"fmt"

	"docscan/internal/analyzer"
	"docscan/internal/collector"
	"docscan/internal/config"
	"docscan/internal/dedup"
	"docscan/internal/document"
	"docscan/internal/scanner"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
	cfg *config.Config
}

func NewApp() (*App, error) {
	cfg, err := config.LoadDefault()
	if err != nil {
		return nil, fmt.Errorf("load default config: %w", err)
	}

	return &App{
		cfg: cfg,
	}, nil
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetToolStatus() []analyzer.ToolStatus {
	return analyzer.CheckTools(a.cfg)
}

// SelectSourceDirectory otwiera natywny Windows directory picker.
func (a *App) SelectSourceDirectory() (string, error) {
	return runtime.OpenDirectoryDialog(
		a.ctx,
		runtime.OpenDialogOptions{
			Title: "Select source directory",
		},
	)
}

// SelectOutputDirectory otwiera natywny Windows directory picker.
func (a *App) SelectOutputDirectory() (string, error) {
	return runtime.OpenDirectoryDialog(
		a.ctx,
		runtime.OpenDialogOptions{
			Title: "Select output directory",
		},
	)
}

// Scan uruchamia:
//
// scanner -> hashing -> dedup
//
// ale jeszcze niczego nie kopiuje.
func (a *App) Scan(path string) ([]document.Document, error) {
	if path == "" {
		return nil, fmt.Errorf("source directory is empty")
	}

	docs, err := scanner.Scan(path, a.cfg)
	if err != nil {
		return nil, fmt.Errorf("scan directory: %w", err)
	}

	dedup.Mark(docs)

	return docs, nil
}

// Collect kopiuje wcześniej przeskanowane dokumenty
// do katalogu wynikowego.
func (a *App) Collect(
	docs []document.Document,
	outputDir string,
) ([]document.Document, error) {

	if outputDir == "" {
		return nil, fmt.Errorf("output directory is empty")
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("no documents to collect")
	}

	if err := collector.Collect(docs, outputDir); err != nil {
		return nil, fmt.Errorf("collect documents: %w", err)
	}

	return docs, nil
}

func (a *App) AnalyzeDocument(
	doc document.Document,
) (analyzer.Analysis, error) {

	results, err := analyzer.Analyze(
		a.ctx,
		doc,
		a.cfg,
	)
	if err != nil {
		return analyzer.Analysis{}, fmt.Errorf(
			"analyze document: %w",
			err,
		)
	}

	return analyzer.Analysis{
		Document: doc,
		Results:  results,
	}, nil
}
