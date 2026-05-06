package parser

type FileRecord struct {
	Path          string
	SizeBytes     int64
	MTimeUnixMS   int64
	ContentHash   string
	Language      string
	LastIndexedAt int64
}

type SymbolRecord struct {
	UID         string
	FilePath    string
	Name        string
	Kind        string
	Package     string
	Receiver    string
	Signature   string
	Doc         string
	Body        string
	LineStart   int
	LineEnd     int
	ColStart    int
	ColEnd      int
	ContentHash string
	UpdatedAt   int64
}

type RefRecord struct {
	ToUID    string
	FilePath string
	Line     int
	Col      int
}

type ParseResult struct {
	Files   []FileRecord
	Symbols []SymbolRecord
	Refs    []RefRecord
}
