package importer

import "github.com/lazeratops/optimusdime/src/document"

type Importer interface {
	Import(filePath string) *document.Document
}
