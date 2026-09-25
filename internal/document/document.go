package document

type Type string

type CollectionStatus string

const (
	StatusPending CollectionStatus = "PENDING"
	StatusCopied  CollectionStatus = "COPIED"
	StatusSkipped CollectionStatus = "SKIPPED"
)

type Document struct {
	SourcePath string `json:"sourcePath"`
	OutputPath string `json:"outputPath"`

	Name string `json:"name"`
	Size int64  `json:"size"`
	Type Type   `json:"type"`

	MD5    string `json:"md5"`
	SHA256 string `json:"sha256"`

	Duplicate bool `json:"duplicate"`

	CollectionStatus CollectionStatus `json:"collectionStatus"`
}
