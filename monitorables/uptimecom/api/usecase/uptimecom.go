//go:build !faker

package usecase

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	uiConfigModels "github.com/monitoror/monitoror/api/config/models"
	coreModels "github.com/monitoror/monitoror/models"
	"github.com/monitoror/monitoror/monitorables/uptimecom/api"
	"github.com/monitoror/monitoror/monitorables/uptimecom/api/models"

	"github.com/AlekSi/pointer"
	"github.com/jsdidierlaurent/echo-middleware/cache"
	uuid "github.com/satori/go.uuid"
)

type (
	uptimecomUsecase struct {
		repository api.Repository
		// Used to generate store key by repository
		repositoryUID string

		// Mutex for lock multi access on Uptimecom
		sync.Mutex
		// Used for caching result of uptimecom (to avoid bursting query limit)
		store           cache.Store
		cacheExpiration int
	}
)

const (
	UptimecomCheckStoreKeyPrefix  = "monitoror.uptimecom.check.store"
	UptimecomChecksStoreKeyPrefix = "monitoror.uptimecom.checks.store"

	UptimecomTagsSeparator = ","
)

func NewUptimecomUsecase(repository api.Repository, store cache.Store, cacheExpiration int) api.Usecase {
	return &uptimecomUsecase{
		repository:      repository,
		repositoryUID:   uuid.NewV4().String(),
		store:           store,
		cacheExpiration: cacheExpiration,
	}
}

func (pu *uptimecomUsecase) Check(params *models.CheckParams) (*coreModels.Tile, error) {
	return pu.check(*params.ID)
}

func (pu *uptimecomUsecase) check(checkID int) (*coreModels.Tile, error) {
	tile := coreModels.NewTile(api.UptimecomCheckTileType)

	var result models.Check

	if err := pu.store.Get(pu.getCheckStoreKey(checkID), &result); err != nil {
		if _, err := pu.loadChecks(); err != nil {
			return nil, &coreModels.MonitororError{Err: err, Tile: tile, Message: "unable to load checks"}
		}

		if err := pu.store.Get(pu.getCheckStoreKey(checkID), &result); err != nil {
			if err != nil {
				return nil, &coreModels.MonitororError{Err: err, Tile: tile, Message: "unable to find check"}
			}
		}
	}

	// Parse result to tile
	tile.Label = result.Name
	tile.Status = parseCheckStatus(&result)

	return tile, nil
}

func (pu *uptimecomUsecase) CheckGenerator(params interface{}) ([]uiConfigModels.GeneratedTile, error) {
	cParams := params.(*models.CheckGeneratorParams)
	return pu.checkGenerator(cParams.Tags, cParams.SortBy)
}

func (pu *uptimecomUsecase) checkGenerator(tags, sortBy string) ([]uiConfigModels.GeneratedTile, error) {
	checks, err := pu.loadChecks()
	if err != nil {
		return nil, &coreModels.MonitororError{Err: err, Message: "unable to list checks"}
	}

	if sortBy == "name" {
		sort.SliceStable(checks, func(i, j int) bool { return checks[i].Name < checks[j].Name })
	}

	var results []uiConfigModels.GeneratedTile

	for _, check := range checks {
		// Filter paused Checks
		if check.IsPaused {
			continue
		}

		// Filter by tags
		if tags != "" {
			if !check.MatchOneTag(strings.Split(tags, UptimecomTagsSeparator)) {
				continue
			}
		}

		// Build results
		p := models.CheckParams{
			ID: pointer.ToInt(check.ID),
		}

		results = append(results, uiConfigModels.GeneratedTile{
			Label:  check.Name,
			Params: p,
		})
	}

	return results, err
}

func (pu *uptimecomUsecase) loadChecks() (results []models.Check, err error) {
	// Synchronize to avoid multi call on uptimecom api
	// We only use loadChecks due to restricted rate limit on uptime.com API (60 calls per minute, 500 calls per hours)
	pu.Lock()
	defer pu.Unlock()

	// Lookup in cache
	key := pu.getChecksStoreKey()
	if err = pu.store.Get(key, &results); err == nil {
		// Cache found, return
		return
	}

	results, err = pu.repository.GetChecks()

	if err != nil {
		return
	}

	// Adding result in store
	_ = pu.store.Set(key, results, time.Millisecond*time.Duration(pu.cacheExpiration))
	for _, check := range results {
		_ = pu.store.Set(pu.getCheckStoreKey(check.ID), check, time.Millisecond*time.Duration(pu.cacheExpiration))
	}

	return
}

func (pu *uptimecomUsecase) getCheckStoreKey(id int) string {
	return fmt.Sprintf("%s:%s-%d", UptimecomCheckStoreKeyPrefix, pu.repositoryUID, id)
}

func (pu *uptimecomUsecase) getChecksStoreKey() string {
	return fmt.Sprintf("%s:%s", UptimecomChecksStoreKeyPrefix, pu.repositoryUID)
}

func parseCheckStatus(check *models.Check) coreModels.TileStatus {
	if check.IsUnderMaintenance {
		return coreModels.WarningStatus
	} else if check.IsPaused {
		return coreModels.DisabledStatus
	} else if check.IsUP {
		return coreModels.SuccessStatus
	} else {
		return coreModels.FailedStatus
	}
}
