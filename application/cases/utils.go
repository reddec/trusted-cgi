package cases

import "github.com/google/uuid"

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
