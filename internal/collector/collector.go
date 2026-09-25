package collector

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"docscan/internal/document"
	"docscan/internal/hashing"
)

func Collect(documents []document.Document, outputDir string) error {
	for i := range documents {
		doc := &documents[i]

		if doc.Duplicate {
			continue
		}

		categoryDir := filepath.Join(
			outputDir,
			string(doc.Type),
		)

		if err := os.MkdirAll(categoryDir, 0755); err != nil {
			return fmt.Errorf(
				"create directory %s: %w",
				categoryDir,
				err,
			)
		}

		outputPath := filepath.Join(
			categoryDir,
			doc.Name,
		)

		resolved, skip, err := resolveDestination(
			outputPath,
			doc.MD5,
			doc.SHA256,
		)
		if err != nil {
			return fmt.Errorf(
				"resolve destination for %s: %w",
				doc.SourcePath,
				err,
			)
		}

		if skip {
			doc.OutputPath = resolved
			doc.CollectionStatus = document.StatusSkipped
			continue
		}

		if err := copyFile(doc.SourcePath, resolved); err != nil {
			return fmt.Errorf(
				"copy %s -> %s: %w",
				doc.SourcePath,
				resolved,
				err,
			)
		}

		doc.OutputPath = resolved
		doc.CollectionStatus = document.StatusCopied
	}

	return nil
}

func resolveDestination(
	path string,
	md5sum string,
	sha256sum string,
) (string, bool, error) {

	// Najpierw sprawdzamy oryginalną nazwę.
	same, exists, err := compareExisting(path, sha256sum)
	if err != nil {
		return "", false, err
	}

	if !exists {
		return path, false, nil
	}

	if same {
		return path, true, nil
	}

	ext := filepath.Ext(path)
	base := path[:len(path)-len(ext)]

	shortHash := md5sum
	if len(shortHash) > 8 {
		shortHash = shortHash[:8]
	}

	// Pierwsza kolizja:
	//
	// raport.pdf
	// raport_aabbccdd.pdf

	candidate := fmt.Sprintf(
		"%s_%s%s",
		base,
		shortHash,
		ext,
	)

	same, exists, err = compareExisting(candidate, sha256sum)
	if err != nil {
		return "", false, err
	}

	if !exists {
		return candidate, false, nil
	}

	if same {
		return candidate, true, nil
	}

	// Jeżeli nawet nazwa z MD5 jest zajęta:
	//
	// raport_aabbccdd_2.pdf
	// raport_aabbccdd_3.pdf
	// ...

	for i := 2; ; i++ {
		candidate = fmt.Sprintf(
			"%s_%s_%d%s",
			base,
			shortHash,
			i,
			ext,
		)

		same, exists, err = compareExisting(
			candidate,
			sha256sum,
		)
		if err != nil {
			return "", false, err
		}

		if !exists {
			return candidate, false, nil
		}

		if same {
			return candidate, true, nil
		}
	}
}

func compareExisting(
	path string,
	expectedSHA256 string,
) (same bool, exists bool, err error) {

	_, err = os.Stat(path)

	if os.IsNotExist(err) {
		return false, false, nil
	}

	if err != nil {
		return false, false, err
	}

	hashes, err := hashing.File(path)
	if err != nil {
		return false, true, err
	}

	return hashes.SHA256 == expectedSHA256, true, nil
}

func copyFile(source, destination string) error {
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(
		destination,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0644,
	)
	if err != nil {
		return err
	}

	success := false

	defer func() {
		_ = dst.Close()

		if !success {
			_ = os.Remove(destination)
		}
	}()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	if err := dst.Sync(); err != nil {
		return err
	}

	success = true

	return nil
}
