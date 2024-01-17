package giant

// Redact is a type for private strings.
type Redact string

// MarshalJSON implements the Marshaler interface.
func (redact Redact) MarshalJSON() ([]byte, error) {

	if redact == "" {
		return []byte(`"--unset--"`), nil
	}

	return []byte(`"--redacted--"`), nil
}
