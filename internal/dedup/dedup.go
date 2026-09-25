package dedup

import "docscan/internal/document"

func Mark(documents []document.Document) {
	seen := make(map[string]struct{})

	for i := range documents {
		hash := documents[i].SHA256

		if _, exists := seen[hash]; exists {
			documents[i].Duplicate = true
			documents[i].CollectionStatus = document.StatusSkipped
			continue
		}

		seen[hash] = struct{}{}
	}
}
