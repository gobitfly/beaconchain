package dataaccess

import (
	"context"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
)

type AdminRepository interface {
	CreateAdConfiguration(ctx context.Context, key, jquerySelector string, insertMode enums.AdInsertMode, refreshInterval uint64, forAllUsers bool, bannerId uint64, htmlContent string, enabled bool) error
	GetAdConfigurations(ctx context.Context, keys []string) ([]t.AdConfigurationData, error)
	UpdateAdConfiguration(ctx context.Context, key, jquerySelector string, insertMode enums.AdInsertMode, refreshInterval uint64, forAllUsers bool, bannerId uint64, htmlContent string, enabled bool) error
	RemoveAdConfiguration(ctx context.Context, key string) error
}

func (d *DataAccessService) CreateAdConfiguration(ctx context.Context, key, jquerySelector string, insertMode enums.AdInsertMode, refreshInterval uint64, forAllUsers bool, bannerId uint64, htmlContent string, enabled bool) error {
	// TODO @DATA-ACCESS
	return d.dummy.CreateAdConfiguration(ctx, key, jquerySelector, insertMode, refreshInterval, forAllUsers, bannerId, htmlContent, enabled)
}

func (d *DataAccessService) GetAdConfigurations(ctx context.Context, keys []string) ([]t.AdConfigurationData, error) {
	// TODO @DATA-ACCESS
	return d.dummy.GetAdConfigurations(ctx, keys)
}

func (d *DataAccessService) UpdateAdConfiguration(ctx context.Context, key, jquerySelector string, insertMode enums.AdInsertMode, refreshInterval uint64, forAllUsers bool, bannerId uint64, htmlContent string, enabled bool) error {
	// TODO @DATA-ACCESS
	return d.dummy.UpdateAdConfiguration(ctx, key, jquerySelector, insertMode, refreshInterval, forAllUsers, bannerId, htmlContent, enabled)
}

func (d *DataAccessService) RemoveAdConfiguration(ctx context.Context, key string) error {
	// TODO @DATA-ACCESS
	return d.dummy.RemoveAdConfiguration(ctx, key)
}
