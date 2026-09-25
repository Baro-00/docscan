package scanner

import (
	"io/fs"
	"path/filepath"
	"strings"

	"docscan/internal/config"
	"docscan/internal/document"
	"docscan/internal/hashing"
)

func Scan(
	root string,
	cfg *config.Config,
) ([]document.Document, error) {
	var documents []document.Document

	extensions := cfg.ExtensionMap()

	err := filepath.WalkDir(root, func(
		path string,
		entry fs.DirEntry,
		err error,
	) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))

		category, ok := extensions[ext]
		if !ok {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		hashes, err := hashing.File(path)
		if err != nil {
			return err
		}

		documents = append(documents, document.Document{
			SourcePath: path,
			Name:       entry.Name(),
			Size:       info.Size(),
			Type:       document.Type(category),

			MD5:    hashes.MD5,
			SHA256: hashes.SHA256,
		})

		return nil
	})

	return documents, err
}
