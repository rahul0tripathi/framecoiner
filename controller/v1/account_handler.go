package v1

import "C"
import (
	"errors"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/labstack/echo/v4"
	"github.com/rahul0tripathi/framecoiner/entity"
	"github.com/rahul0tripathi/framecoiner/pkg/server"
)

const (
	_paramOwner = "owner"
)

func (h *Handler) MakeGetAccountHandler(svc AccountService) echo.HandlerFunc {
	return func(c echo.Context) error {
		owner := c.Param(_paramOwner)
		if !common.IsHexAddress(owner) {
			return server.ResponseJSON(c, http.StatusBadRequest, map[string]interface{}{
				"error": "invalid owner address",
			})
		}

		owner, err := svc.GetTradingAccount(c.Request().Context(), common.HexToAddress(owner))
		switch {
		case err == nil:
		case errors.Is(err, entity.ErrNoAccountFound):
			return server.ResponseJSON(c, http.StatusNotFound, map[string]interface{}{
				"error": "account not found",
			})
		case err != nil:
			return server.ResponseJSON(c, http.StatusInternalServerError, map[string]interface{}{
				"error": err.Error(),
			})
		}

		return server.ResponseJSON(c, http.StatusOK, map[string]interface{}{
			"data": map[string]interface{}{
				"owner": owner,
			},
		})
	}
}
