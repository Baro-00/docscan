package document

type Type string
type CollectionStatus string

const (
	StatusPending CollectionStatus = "PENDING"
	StatusCopied  CollectionStatus = "COPIED"
	StatusSkipped CollectionStatus = "SKIPPED"
)

type Document struct {
	SourcePath string
	OutputPath string

	Name string
	Size int64
	Type string

	MD5    string
	SHA256 string

	Duplicate bool

	CollectionStatus CollectionStatus
}
