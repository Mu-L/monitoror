//go:generate mockery --name Repository

package api

import (
	"github.com/monitoror/monitoror/monitorables/uptimecom/api/models"
)

type (
	Repository interface {
		GetCheck(checkID int) (*models.Check, error)
		GetChecks() ([]models.Check, error)
	}
)
