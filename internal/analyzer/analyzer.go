package analyzer

import (
	"context"
	"fmt"

	"docscan/internal/config"
	"docscan/internal/document"
)

func Analyze(
	ctx context.Context,
	doc document.Document,
	cfg *config.Config,
) ([]Result, error) {

	toolsDir := cfg.ToolsPath()

	category, exists := cfg.Category(string(doc.Type))
	if !exists {
		return nil, fmt.Errorf(
			"unknown document category %q",
			doc.Type,
		)
	}

	results := make(
		[]Result,
		0,
		len(category.Analyzers),
	)

	for _, step := range category.Analyzers {
		tool, exists := cfg.Tools[step.Tool]
		if !exists {
			return nil, fmt.Errorf(
				"analyzer %q references unknown tool %q",
				step.Name,
				step.Tool,
			)
		}

		result := Execute(
			ctx,
			step,
			tool,
			doc.SourcePath,
			toolsDir,
		)

		results = append(results, result)
	}

	return results, nil
}
