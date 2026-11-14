package embed

import "errors"

// Error categories let callers distinguish invalid input from unsupported
// output. Underlying I/O errors remain available through errors.Is and
// errors.As when they are wrapped by an operation.
var (
	// ErrInvalidInput reports an empty, malformed, duplicate, or otherwise
	// unsupported input value.
	ErrInvalidInput = errors.New("invalid input")
	// ErrUnsupportedLanguage reports a target language that mmmbed cannot emit.
	ErrUnsupportedLanguage = errors.New("unsupported language")
)
