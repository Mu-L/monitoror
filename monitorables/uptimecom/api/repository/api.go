package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/monitoror/monitoror/monitorables/uptimecom/api"
	"github.com/monitoror/monitoror/monitorables/uptimecom/api/models"
	"github.com/monitoror/monitoror/monitorables/uptimecom/config"
	"github.com/monitoror/monitoror/pkg/gouptimecom"

	uptimecomAPI "github.com/jsdidierlaurent/uptime-client-go"
)

const PageSize = 50 // Uptime.com sell checks by batch of 50

type (
	uptimecomRepository struct {
		config *config.Uptimecom

		// Uptimecom check client
		uptimecomCheckService gouptimecom.UptimecomCheckService
	}
)

func NewUptimecomRepository(config *config.Uptimecom) api.Repository {
	// Add / if missing
	if !strings.HasSuffix(config.URL, "/") {
		config.URL += "/"
	}

	client, err := uptimecomAPI.NewClient(&uptimecomAPI.Config{
		BaseURL: config.URL,
		Token:   config.Token,
	})

	// Only if Uptimecom URL is not a valid URL
	if err != nil {
		panic(fmt.Sprintf("unable to initiate connection to Uptime.com\n. %v\n", err))
	}

	return &uptimecomRepository{
		config:                config,
		uptimecomCheckService: client.Checks,
	}
}

func (r *uptimecomRepository) GetCheck(id int) (result *models.Check, err error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.config.Timeout)*time.Millisecond)
	defer cancel()

	check, _, err := r.uptimecomCheckService.Get(ctx, id)
	if err != nil {
		return
	}

	result = &models.Check{
		ID:   check.PK,
		Name: check.Name,

		IsUP:               check.StateIsUp,
		IsPaused:           check.IsPaused,
		IsUnderMaintenance: check.IsUnderMaintenance,

		Tags: check.Tags,
	}

	return
}

func (r *uptimecomRepository) GetChecks() (results []models.Check, err error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.config.Timeout)*time.Millisecond)
	defer cancel()

	checks, err := r.uptimecomCheckService.ListAll(ctx, &uptimecomAPI.CheckListOptions{PageSize: PageSize})
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, check := range checks {
		results = append(results, models.Check{
			ID:   check.PK,
			Name: check.Name,

			IsUP:               check.StateIsUp,
			IsPaused:           check.IsPaused,
			IsUnderMaintenance: check.IsUnderMaintenance,

			Tags: check.Tags,
		})
	}

	return
}
