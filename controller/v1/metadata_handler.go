package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rahul0tripathi/framecoiner/pkg/server"
)

const (
	FrameCoinerVersion = "v0.0.1"
)

func (h *Handler) MakeGetFrameCoinerMetadataHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		return server.ResponseJSON(c, http.StatusOK, map[string]interface{}{
			"version": FrameCoinerVersion,
		})
	}
}
