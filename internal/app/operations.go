package app

import (
	"errors"
	"fmt"
	data "github.com/pheelee/wishmas/internal/household"
	"net/http"
	"os"
)

// @impl OPS-001-AC3: only fixed operation names and safe error categories are logged.
type operationResponse struct {
	http.ResponseWriter
	status int
}

func (w *operationResponse) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
		w.ResponseWriter.WriteHeader(status)
	}
}
func (w *operationResponse) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}
func (w *operationResponse) Unwrap() http.ResponseWriter { return w.ResponseWriter }
func operationRoute(path string) string {
	switch path {
	case "/manage/list/create", "/manage/list/rename", "/manage/list/delete", "/manage/user/create", "/manage/user/rename", "/manage/user/delete", "/manage/recipient/create", "/manage/recipient/rename", "/manage/recipient/delete", "/manage/selection/add", "/manage/selection/remove", "/manage/link/regenerate", "/setup/finish", "/login", "/logout", "/claims/toggle", "/wishes/save", "/wishes/delete", "/wishes/assign", "/wishes/unassign":
		return path
	}
	// Never include unknown user-controlled paths, names or query parameters.
	return "household"
}
func diagnosticKind(err error) string {
	for _, known := range []error{data.ErrStale, data.ErrReadOnly, data.ErrStorageUnavailable, data.ErrStorageConfig, data.ErrUncertain, errPictureURL, errPicturePrivate, errPictureRedirect, errPictureTimeout, errPictureDownload, errPictureBytes, errPictureFormat, errPicturePixels} {
		if errors.Is(err, known) {
			return known.Error()
		}
	}
	if errors.Is(err, os.ErrPermission) {
		return "permission denied"
	}
	if errors.Is(err, os.ErrNotExist) {
		return "file not found"
	}
	return fmt.Sprintf("%T", err)
}
