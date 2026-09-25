package main

import (
	"fmt"
	"log"
	"os"

	"docscan/internal/collector"
	"docscan/internal/config"
	"docscan/internal/dedup"
	"docscan/internal/scanner"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf(
			"usage: %s <source-directory> <output-directory>\n",
			os.Args[0],
		)
		os.Exit(1)
	}

	sourceDir := os.Args[1]
	outputDir := os.Args[2]

	cfg, err := config.LoadDefault()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("[*] Scanning...")

	docs, err := scanner.Scan(sourceDir, cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("[+] Found %d documents\n", len(docs))

	fmt.Println("[*] Deduplicating...")

	dedup.Mark(docs)

	fmt.Println("[*] Collecting...")

	if err := collector.Collect(docs, outputDir); err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	fmt.Println("Results:")

	for _, doc := range docs {
		if doc.Duplicate {
			fmt.Printf(
				"[DUP]  %-4s %s\n",
				doc.Type,
				doc.SourcePath,
			)
			continue
		}

		fmt.Printf(
			"[COPY] %-4s %s -> %s\n",
			doc.Type,
			doc.SourcePath,
			doc.OutputPath,
		)
	}
}
