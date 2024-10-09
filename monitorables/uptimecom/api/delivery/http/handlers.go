package http

import (
	"net/http"

	"github.com/monitoror/monitoror/internal/pkg/monitorable/delivery"
	"github.com/monitoror/monitoror/monitorables/uptimecom/api"
	"github.com/monitoror/monitoror/monitorables/uptimecom/api/models"

	"github.com/labstack/echo/v4"
)

type UptimecomDelivery struct {
	uptimecomUsecase api.Usecase
}

func NewUptimecomDelivery(p api.Usecase) *UptimecomDelivery {
	return &UptimecomDelivery{p}
}

func (h *UptimecomDelivery) GetCheck(c echo.Context) error {
	// Bind / Check Params
	params := &models.CheckParams{}
	if err := delivery.BindAndValidateParams(c, params); err != nil {
		return err
	}

	tile, err := h.uptimecomUsecase.Check(params)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, tile)
}
