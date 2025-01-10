package acctest

import "github.com/google/uuid"

func IsValidUUID(value string) error {
	_, err := uuid.Parse(value)
	return err
}
