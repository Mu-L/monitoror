package repository

import (
	"errors"
	"testing"

	"github.com/monitoror/monitoror/monitorables/uptimecom/config"
	pkgUptimecom "github.com/monitoror/monitoror/pkg/gouptimecom"
	"github.com/monitoror/monitoror/pkg/gouptimecom/mocks"

	uptimecom "github.com/jsdidierlaurent/uptime-client-go"
	"github.com/stretchr/testify/assert"
	. "github.com/stretchr/testify/mock"
)

func initRepository(t *testing.T, checkAPI pkgUptimecom.UptimecomCheckService) *uptimecomRepository {
	conf := &config.Uptimecom{
		URL:             "https://uptimecom.example.com",
		Token:           "TEST",
		Timeout:         config.Default.Timeout,
		CacheExpiration: config.Default.CacheExpiration,
	}
	repository := NewUptimecomRepository(conf)

	assert.Equal(t, "https://uptimecom.example.com/", conf.URL)

	apiUptimecomRepository, ok := repository.(*uptimecomRepository)
	if assert.True(t, ok) {
		apiUptimecomRepository.uptimecomCheckService = checkAPI
		return apiUptimecomRepository
	}
	return nil
}

func TestUptimecomRepository_NewUptimecomRepository_Error(t *testing.T) {
	conf := &config.Uptimecom{
		URL:             "wrong%url",
		Token:           config.Default.Token,
		Timeout:         config.Default.Timeout,
		CacheExpiration: config.Default.CacheExpiration,
	}

	assert.Panics(t, func() { _ = NewUptimecomRepository(conf) })
}

func TestUptimecomRepository_GetUptimecomCheck_Success(t *testing.T) {
	mock := new(mocks.UptimecomCheckService)
	mock.On("Get", Anything, Anything).Return(&uptimecom.Check{PK: 1000, Name: "Check 1", StateIsUp: true, IsUnderMaintenance: false, IsPaused: false, Tags: []string{"tag1"}}, nil, nil)

	repository := initRepository(t, mock)
	check, err := repository.GetCheck(1000)
	if assert.NoError(t, err) {
		assert.Equal(t, "Check 1", check.Name)
		assert.Equal(t, true, check.IsUP)
		assert.Equal(t, false, check.IsPaused)
		assert.Equal(t, false, check.IsUnderMaintenance)
		assert.Len(t, check.Tags, 1)
	}

	mock.AssertNumberOfCalls(t, "Get", 1)
	mock.AssertExpectations(t)
}

func TestUptimecomRepository_GetUptimecomCheck_Error(t *testing.T) {
	mock := new(mocks.UptimecomCheckService)
	mock.On("Get", Anything, Anything).Return(nil, nil, errors.New("boom"))

	repository := initRepository(t, mock)
	_, err := repository.GetCheck(1000)
	assert.Error(t, err)
	mock.AssertNumberOfCalls(t, "Get", 1)
	mock.AssertExpectations(t)
}

func TestUptimecomRepository_GetUptimecomChecks_Success(t *testing.T) {
	mock := new(mocks.UptimecomCheckService)
	mock.On("ListAll", Anything, Anything).Return([]*uptimecom.Check{
		{PK: 1000, Name: "Check 1", StateIsUp: true, IsUnderMaintenance: false, IsPaused: false},
		{PK: 2000, Name: "Check 2", StateIsUp: true, IsUnderMaintenance: false, IsPaused: false},
		{PK: 3000, Name: "Check 3", StateIsUp: false, IsUnderMaintenance: false, IsPaused: false},
	}, nil)

	repository := initRepository(t, mock)
	checks, err := repository.GetChecks()
	if assert.NoError(t, err) {
		assert.Len(t, checks, 3)
	}

	mock.AssertNumberOfCalls(t, "ListAll", 1)
	mock.AssertExpectations(t)
}

func TestUptimecomRepository_GetUptimecomChecks_Error(t *testing.T) {
	mock := new(mocks.UptimecomCheckService)
	mock.On("ListAll", Anything, Anything).Return(nil, errors.New("boom"))

	repository := initRepository(t, mock)
	_, err := repository.GetChecks()
	assert.Error(t, err)
	mock.AssertNumberOfCalls(t, "ListAll", 1)
	mock.AssertExpectations(t)
}
