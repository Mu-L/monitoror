package usecase

import (
	"errors"
	"testing"
	"time"

	coreModels "github.com/monitoror/monitoror/models"
	"github.com/monitoror/monitoror/monitorables/uptimecom/api"
	"github.com/monitoror/monitoror/monitorables/uptimecom/api/mocks"
	"github.com/monitoror/monitoror/monitorables/uptimecom/api/models"

	"github.com/AlekSi/pointer"
	"github.com/jsdidierlaurent/echo-middleware/cache"
	"github.com/stretchr/testify/assert"
)

func initUsecase(mockRepository api.Repository) api.Usecase {
	store := cache.NewGoCacheStore(time.Minute*5, time.Second)
	pu := NewUptimecomUsecase(mockRepository, store, 1000)
	return pu
}

func TestUptimecomUsecase_Check_LoadChecks_Error(t *testing.T) {
	mockRepository := new(mocks.Repository)
	mockRepository.On("GetChecks").
		Return(nil, errors.New("boom"))

	pu := initUsecase(mockRepository)

	tile, err := pu.Check(&models.CheckParams{ID: pointer.ToInt(1000)})
	if assert.Error(t, err) {
		assert.Nil(t, tile)
		assert.IsType(t, &coreModels.MonitororError{}, err)
		assert.Equal(t, "unable to load checks", err.Error())
		mockRepository.AssertNumberOfCalls(t, "GetChecks", 1)
		mockRepository.AssertExpectations(t)
	}
}

func TestUptimecomUsecase_Check_NotFound_Error(t *testing.T) {
	mockRepository := new(mocks.Repository)
	mockRepository.On("GetChecks").
		Return([]models.Check{
			{ID: 1000, Name: "Check 1", IsUP: true},
			{ID: 1100, Name: "Check 2", IsUP: true},
		}, nil)

	pu := initUsecase(mockRepository)

	tile, err := pu.Check(&models.CheckParams{ID: pointer.ToInt(2000)})
	if assert.Error(t, err) {
		assert.Nil(t, tile)
		assert.IsType(t, &coreModels.MonitororError{}, err)
		assert.Equal(t, "unable to find check", err.Error())
		mockRepository.AssertNumberOfCalls(t, "GetChecks", 1)
		mockRepository.AssertExpectations(t)
	}
}

func TestUptimecomUsecase_Check_WithoutCache(t *testing.T) {
	mockRepository := new(mocks.Repository)
	mockRepository.On("GetChecks").
		Return([]models.Check{
			{ID: 1000, Name: "Check 1", IsUP: true},
			{ID: 1100, Name: "Check 2", IsUP: true},
			{ID: 1200, Name: "Check 3", IsUP: true},
		}, nil)

	pu := initUsecase(mockRepository)

	tile, err := pu.Check(&models.CheckParams{ID: pointer.ToInt(1000)})
	if assert.NoError(t, err) {
		assert.NotNil(t, tile)
		assert.Equal(t, coreModels.SuccessStatus, tile.Status)
		assert.Equal(t, "Check 1", tile.Label)
		mockRepository.AssertNumberOfCalls(t, "GetChecks", 1)
		mockRepository.AssertExpectations(t)
	}
}

func TestUptimecomUsecase_Check_WithCache(t *testing.T) {
	mockRepository := new(mocks.Repository)
	mockRepository.On("GetChecks").
		Return([]models.Check{
			{ID: 1000, Name: "Check 1", IsUP: true},
			{ID: 1100, Name: "Check 2", IsUP: true},
			{ID: 1200, Name: "Check 3", IsUP: true},
		}, nil)

	pu := initUsecase(mockRepository).(*uptimecomUsecase)

	tile, err := pu.Check(&models.CheckParams{ID: pointer.ToInt(1000)})
	if assert.NoError(t, err) {
		assert.NotNil(t, tile)
		assert.Equal(t, coreModels.SuccessStatus, tile.Status)
		assert.Equal(t, "Check 1", tile.Label)

		// call second time to check cache
		_, err := pu.loadChecks()
		assert.NoError(t, err)

		mockRepository.AssertNumberOfCalls(t, "GetChecks", 1)
		mockRepository.AssertExpectations(t)
	}
}

func TestUptimecomUsecase_CheckGenerator_Error(t *testing.T) {
	mockRepository := new(mocks.Repository)
	mockRepository.On("GetChecks").Return(nil, errors.New("boom"))

	pu := initUsecase(mockRepository)

	results, err := pu.CheckGenerator(&models.CheckGeneratorParams{SortBy: "name"})
	if assert.Error(t, err) {
		assert.Nil(t, results)
		mockRepository.AssertNumberOfCalls(t, "GetChecks", 1)
		mockRepository.AssertExpectations(t)
	}
}

//nolint:dupl
func TestUptimecomUsecase_CheckGenerator(t *testing.T) {
	mockRepository := new(mocks.Repository)
	mockRepository.On("GetChecks").
		Return([]models.Check{
			{ID: 1000, Name: "Check 2", Tags: []string{"tag1"}},
			{ID: 1100, Name: "Check 1", Tags: []string{"tag1"}},
			{ID: 1200, Name: "Check 3", Tags: []string{"tag2"}},
			{ID: 1200, Name: "Check 4", Tags: []string{"tag1"}, IsPaused: true},
		}, nil)

	pu := initUsecase(mockRepository)

	results, err := pu.CheckGenerator(&models.CheckGeneratorParams{SortBy: "name", Tags: "tag1"})
	if assert.NoError(t, err) {
		assert.NotNil(t, results)
		assert.Len(t, results, 2)
		assert.Equal(t, "Check 1", results[0].Label)
		assert.Equal(t, 1100, *results[0].Params.(models.CheckParams).ID)
		assert.Equal(t, "Check 2", results[1].Label)
		assert.Equal(t, 1000, *results[1].Params.(models.CheckParams).ID)
		mockRepository.AssertNumberOfCalls(t, "GetChecks", 1)
		mockRepository.AssertExpectations(t)
	}
}

func TestUptimecomUsecase_ParseStatus(t *testing.T) {
	assert.Equal(t, coreModels.SuccessStatus, parseCheckStatus(&models.Check{IsUP: true}))
	assert.Equal(t, coreModels.FailedStatus, parseCheckStatus(&models.Check{IsUP: false}))
	assert.Equal(t, coreModels.WarningStatus, parseCheckStatus(&models.Check{IsUnderMaintenance: true}))
	assert.Equal(t, coreModels.DisabledStatus, parseCheckStatus(&models.Check{IsPaused: true}))
}
