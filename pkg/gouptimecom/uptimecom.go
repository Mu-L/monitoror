//go:generate mockery --name UptimecomCheckService

package gouptimecom

import (
	"context"
	"net/http"

	uptimecom "github.com/jsdidierlaurent/uptime-client-go"
)

type UptimecomCheckService interface {
	List(ctx context.Context, opt *uptimecom.CheckListOptions) ([]*uptimecom.Check, *http.Response, error)
	ListAll(ctx context.Context, opt *uptimecom.CheckListOptions) ([]*uptimecom.Check, error)
	Get(ctx context.Context, pk int) (*uptimecom.Check, *http.Response, error)
}
