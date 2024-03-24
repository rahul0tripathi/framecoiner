package v1

import (
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/labstack/echo/v4"
	"github.com/rahul0tripathi/framecoiner/pkg/server"
)

const (
	FrameCoinerVersion = "v0.0.1"
	_paramTokenAddress = "tokenAddress"
)

func (h *Handler) MakeGetFrameCoinerMetadataHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		return server.ResponseJSON(c, http.StatusOK, map[string]interface{}{
			"version": FrameCoinerVersion,
		})
	}
}

func (h *Handler) MakeGetTokenMetadataHandler(svc TokenMetadataService) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Param(_paramTokenAddress)
		if !common.IsHexAddress(token) {
			return server.ResponseJSON(c, http.StatusBadRequest, map[string]interface{}{
				"error": "invalid token address",
			})
		}

		metadata, err := svc.GetTokenMetadata(c.Request().Context(), common.HexToAddress(token))
		if err != nil {
			return server.ResponseJSON(c, http.StatusInternalServerError, map[string]interface{}{
				"error": err.Error(),
			})
		}

		return server.ResponseJSON(c, http.StatusOK, metadata)
	}
}
