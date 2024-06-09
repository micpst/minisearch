package tokenizer

import "fmt"

type LanguageNotSupportedError struct {
	Language Language
}

func (e *LanguageNotSupportedError) Error() string {
	return fmt.Sprintf("Language '%s' is not supported", e.Language)
}
