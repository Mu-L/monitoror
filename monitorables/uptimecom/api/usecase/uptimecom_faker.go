//go:build faker

package usecase

import (
	"fmt"
	"time"

	uiConfigModels "github.com/monitoror/monitoror/api/config/models"
	"github.com/monitoror/monitoror/internal/pkg/monitorable/faker"
	"github.com/monitoror/monitoror/models"
	"github.com/monitoror/monitoror/monitorables/uptimecom/api"
	uptimecomModels "github.com/monitoror/monitoror/monitorables/uptimecom/api/models"
	"github.com/monitoror/monitoror/pkg/nonempty"
)

type (
	uptimecomUsecase struct {
		timeRefByCheck map[string]time.Time
	}
)

var availableStatuses = faker.Statuses{
	{models.SuccessStatus, time.Second * 30},
	{models.FailedStatus, time.Second * 30},
	{models.DisabledStatus, time.Second * 10},
	{models.WarningStatus, time.Second * 10},
}

func NewUptimecomUsecase() api.Usecase {
	return &uptimecomUsecase{make(map[string]time.Time)}
}

func (pu *uptimecomUsecase) Check(params *uptimecomModels.CheckParams) (tile *models.Tile, error error) {
	tile = models.NewTile(api.UptimecomCheckTileType)
	tile.Label = fmt.Sprintf(fmt.Sprintf("Check %d", *params.ID))

	// Code
	tile.Status = nonempty.Struct(params.Status, pu.computeStatus(fmt.Sprintf("%d", *params.ID))).(models.TileStatus)

	return
}

func (pu *uptimecomUsecase) CheckGenerator(params interface{}) ([]uiConfigModels.GeneratedTile, error) {
	panic("unimplemented")
}

func (pu *uptimecomUsecase) computeStatus(id string) models.TileStatus {
	value, ok := pu.timeRefByCheck[id]
	if !ok {
		pu.timeRefByCheck[id] = faker.GetRefTime()
	}

	return faker.ComputeStatus(value, availableStatuses)
}
