package converter

import (
	"github.com/lazeratops/optimusdime/src/document"
)

type Converter interface {
	Convert(targetCurrency document.Currency, statement *document.Document) (*document.Document, *document.Document, error)
}
