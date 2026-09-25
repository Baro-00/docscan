package analyzer

import "docscan/internal/document"

type Analysis struct {
	Document document.Document `json:"document"`
	Results  []Result          `json:"results"`
}
