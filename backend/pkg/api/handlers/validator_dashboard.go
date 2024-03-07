package handlers

import (
	"errors"
	"net/http"

	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	types "github.com/gobitfly/beaconchain/pkg/api/types"
)

type vdbFetcher interface {
	getValidatorDashboardOverview() (types.VDBOverviewData, error)
	removeValidatorDashboard(r *http.Request) error
}

func (h *HandlerService) getDashboardFetcher(r *http.Request, dashboardId interface{}) (vdbFetcher, error) {
	switch id := dashboardId.(type) {
	case types.VDBIdPrimary:
		fetcher := &dashboardFetcher{h: h, dashboardId: id, authenticated: false}
		// check if user has access to this dashboard
		access, err := fetcher.hasAccess(r)
		if err != nil {
			return nil, err
		}
		if !access {
			return nil, ErrForbidden
		}
		fetcher.authenticated = true
		return fetcher, nil
	case types.VDBIdPublic:
		dashboardInfo, err := h.da.GetValidatorDashboardInfoByPublicId(id)
		if err != nil {
			return nil, err
		}
		return &dashboardFetcher{h: h, dashboardId: dashboardInfo.Id, authenticated: false}, nil
	case []string:
		validators, err := h.da.GetValidatorsFromStrings(id)
		if err != nil {
			return nil, err
		}
		return &validatorsFetcher{da: h.da, validators: validators}, nil
	default:
		return nil, ErrParseDashboardId
	}
}

// ------------------
// Primary Id Fetcher

// TODO move this to a more appropriate place
var ErrForbidden = errors.New("user does not have access to this dashboard")

type dashboardFetcher struct {
	h             *HandlerService
	dashboardId   types.VDBIdPrimary
	authenticated bool
}

var _ vdbFetcher = (*validatorsFetcher)(nil)

func (f *dashboardFetcher) hasAccess(r *http.Request) (bool, error) {
	if f.authenticated {
		return false, nil
	}
	user, err := getUser(r)
	if err != nil {
		return false, err
	}
	dashboardInfo, err := f.h.da.GetValidatorDashboardInfo(f.dashboardId)
	if err != nil {
		return false, err
	}
	if user.Id != dashboardInfo.UserId {
		return false, nil
	}
	return true, nil
}

func (f *dashboardFetcher) removeValidatorDashboard(r *http.Request) error {
	// check access
	access, err := f.hasAccess(r)
	if err != nil {
		return err
	}
	if !access {
		return ErrForbidden
	}
	return f.h.da.RemoveValidatorDashboard(f.dashboardId)
}

func (f *dashboardFetcher) getValidatorDashboardOverview() (types.VDBOverviewData, error) {
	return f.h.da.GetValidatorDashboardOverview(f.dashboardId)
}

// ------------------
// Validator Set Fetcher

type validatorsFetcher struct {
	da         dataaccess.DataAccessor
	validators types.VDBIdValidatorSet
}

var _ vdbFetcher = (*validatorsFetcher)(nil)

func (f *validatorsFetcher) getValidatorDashboardOverview() (types.VDBOverviewData, error) {
	return f.da.GetValidatorDashboardOverviewByValidators(f.validators)
}

func (f *validatorsFetcher) removeValidatorDashboard(r *http.Request) error {
	return nil
}
