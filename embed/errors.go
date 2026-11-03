package embed

import "errors"

// Error categories let callers distinguish invalid input from unsupported output.
// Underlying I/O errors remain available through errors.Is and errors.As.
var (
	ErrInvalidInput        = errors.New("invalid input")
	ErrUnsupportedLanguage = errors.New("unsupported language")
)
