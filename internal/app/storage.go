package app

import (
	"errors"
	data "github.com/pheelee/wishmas/internal/household"
)

// @impl CON-009-AC9 CON-009-AC10 CON-009-AC11 CON-009-AC13
func storageMessage(err error, fallback string) string {
	switch {
	case errors.Is(err, data.ErrReadOnly):
		return "HR113"
	case errors.Is(err, data.ErrUncertain):
		return "HR119"
	case errors.Is(err, data.ErrMultipleHouseholds):
		return "HR117"
	case errors.Is(err, data.ErrStorageConfig):
		return "HR116"
	case errors.Is(err, data.ErrStorageUnavailable):
		return "HR115"
	}
	return fallback
}
