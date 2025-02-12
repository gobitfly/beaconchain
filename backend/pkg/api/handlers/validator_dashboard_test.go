package handlers

import (
	"context"
	"errors"
	"testing"

	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	"github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/stretchr/testify/assert"
)

// ------------------------------------------------------------

type dataAccessStub struct {
	dataaccess.DummyService
}

func (da *dataAccessStub) GetUserInfo(ctx context.Context, id uint64) (*types.UserInfo, error) {
	return &types.UserInfo{
		PremiumPerks: types.PremiumPerks{
			ValidatorGroupsPerDashboard: 1,
		},
	}, nil
}

func (da *dataAccessStub) GetValidatorDashboardGroupCount(ctx context.Context, dashboardId types.VDBIdPrimary) (uint64, error) {
	var count uint64
	if dashboardId == 1 {
		count = 1
	}
	return count, nil
}

// ------------------------------------------------------------

func TestInputPostValidatorDashboardGroupsValidate(t *testing.T) {
	var i inputPostValidatorDashboardGroups
	params := make(map[string]string)
	t.Run("success", func(t *testing.T) {
		params["dashboard_id"] = "1"
		body := stringAsBody(`{"name":"test"}`)
		i, err := i.Validate(params, body)
		assert.Nil(t, err)
		assert.Equal(t, types.VDBIdPrimary(1), i.dashboardId)
		assert.Equal(t, "test", i.name)
	})
	t.Run("empty name", func(t *testing.T) {
		params["dashboard_id"] = "1"
		body := stringAsBody(`{"name":""}`)
		_, err := i.Validate(params, body)
		assert.NotNil(t, err)
	})
}
func TestPostValidatorDashboardGroups(t *testing.T) {
	ctx, h := handlerTestSetup()

	t.Run("success", func(t *testing.T) {
		input := inputPostValidatorDashboardGroups{
			dashboardId: 0,
			name:        "test",
		}
		_, err := h.PostValidatorDashboardGroups(ctx, input)
		assert.Nil(t, err)
	})
	t.Run("group count reached", func(t *testing.T) {
		input := inputPostValidatorDashboardGroups{
			dashboardId: 1,
			name:        "test",
		}
		_, err := h.PostValidatorDashboardGroups(ctx, input)
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, errConflict))
	})
}

// ------------------------------------------------------------

func TestInputGetValidatorDashboardGroupSummaryValidate(t *testing.T) {
	var i inputGetValidatorDashboardGroupSummary
	t.Run("success", func(t *testing.T) {
		params := map[string]string{
			"dashboard_id": "1",
			"group_id":     "1",
			"period":       "all_time",
		}
		i, err := i.Validate(params, nil)
		assert.Nil(t, err)
		assert.Equal(t, types.VDBIdPrimary(1), i.dashboardIdParam)
		assert.Equal(t, int64(1), i.groupId)
	})
	t.Run("empty dashboard_id", func(t *testing.T) {
		params := map[string]string{
			"dashboard_id": "",
			"group_id":     "1",
			"period":       "all_time",
		}
		_, err := i.Validate(params, nil)
		assert.NotNil(t, err)
	})
	t.Run("empty group_id", func(t *testing.T) {
		params := map[string]string{
			"dashboard_id": "1",
			"group_id":     "",
			"period":       "all_time",
		}
		_, err := i.Validate(params, nil)
		assert.NotNil(t, err)
	})
	t.Run("empty period", func(t *testing.T) {
		params := map[string]string{
			"dashboard_id": "1",
			"group_id":     "1",
			"period":       "",
		}
		_, err := i.Validate(params, nil)
		assert.NotNil(t, err)
	})
}
