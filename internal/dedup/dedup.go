package dedup

import "docscan/internal/document"

func Mark(documents []document.Document) {
	seen := make(map[string]struct{})

	for i := range documents {
		doc := &documents[i]

		doc.Duplicate = false
		doc.CollectionStatus = document.StatusPending

		if _, exists := seen[doc.SHA256]; exists {
			doc.Duplicate = true
			doc.CollectionStatus = document.StatusSkipped

			continue
		}

		seen[doc.SHA256] = struct{}{}
	}
}
