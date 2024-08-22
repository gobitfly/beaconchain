package archiver

import (
	"context"
	"slices"
	"time"

	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	"github.com/gobitfly/beaconchain/pkg/api/handlers"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

type Archiver struct {
	das *dataaccess.DataAccessService
}

func NewArchiver(d *dataaccess.DataAccessService) (*Archiver, error) {
	archiver := &Archiver{
		das: d,
	}
	return archiver, nil
}

func (a *Archiver) Start() {
	for {
		err := a.updateArchivedStatus()
		if err != nil {
			log.Error(err, "failed updating dashboard archive status", 0)
		}
		time.Sleep(utils.Day)
	}
}

func (a *Archiver) updateArchivedStatus() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var dashboardsToBeArchived []t.ArchiverDashboardArchiveReason
	var dashboardsToBeDeleted []uint64

	// Get all dashboards for all users
	userDashboards, err := a.das.GetValidatorDashboardsCountInfo(ctx)
	if err != nil {
		return err
	}

	for userId, dashboards := range userDashboards {
		userInfo, err := a.das.GetUserInfo(ctx, userId)
		if err != nil {
			return err
		}

		if userInfo.UserGroup == t.UserGroupAdmin {
			// Don't archive or delete anything for admins
			continue
		}

		var archivedDashboards []uint64
		var activeDashboards []uint64

		// Check if the active user dashboard exceeds the maximum number of groups, or validators
		for _, dashboardInfo := range dashboards {
			if dashboardInfo.IsArchived {
				archivedDashboards = append(archivedDashboards, dashboardInfo.DashboardId)
			} else {
				if dashboardInfo.GroupCount > userInfo.PremiumPerks.ValidatorGroupsPerDashboard {
					dashboardsToBeArchived = append(dashboardsToBeArchived, t.ArchiverDashboardArchiveReason{DashboardId: dashboardInfo.DashboardId, ArchivedReason: enums.VDBArchivedReasons.Groups})
				} else if dashboardInfo.ValidatorCount > userInfo.PremiumPerks.ValidatorsPerDashboard {
					dashboardsToBeArchived = append(dashboardsToBeArchived, t.ArchiverDashboardArchiveReason{DashboardId: dashboardInfo.DashboardId, ArchivedReason: enums.VDBArchivedReasons.Validators})
				} else {
					activeDashboards = append(activeDashboards, dashboardInfo.DashboardId)
				}
			}
		}

		// Check if the user still exceeds the maximum number of active dashboards
		dashboardLimit := int(userInfo.PremiumPerks.ValidatorDashboards)
		if len(activeDashboards) > dashboardLimit {
			slices.Sort(activeDashboards)
			for id := 0; id < len(activeDashboards)-dashboardLimit; id++ {
				dashboardsToBeArchived = append(dashboardsToBeArchived, t.ArchiverDashboardArchiveReason{DashboardId: activeDashboards[id], ArchivedReason: enums.VDBArchivedReasons.Dashboards})
			}
		}

		// Check if the user exceeds the maximum number of archived dashboards
		archivedLimit := handlers.MaxArchivedDashboardsCount
		if len(archivedDashboards)+len(dashboardsToBeArchived) > archivedLimit {
			dashboardsToBeDeletedForUser := archivedDashboards
			for _, dashboard := range dashboardsToBeArchived {
				dashboardsToBeDeletedForUser = append(dashboardsToBeDeletedForUser, dashboard.DashboardId)
			}
			slices.Sort(dashboardsToBeDeletedForUser)
			dashboardsToBeDeletedForUser = dashboardsToBeDeletedForUser[:len(dashboardsToBeDeletedForUser)-archivedLimit]
			dashboardsToBeDeleted = append(dashboardsToBeDeleted, dashboardsToBeDeletedForUser...)
		}
	}

	// Remove dashboards that should be deleted from the to be archived list
	dashboardsToBeDeletedMap := utils.SliceToMap(dashboardsToBeDeleted)
	for i := 0; i < len(dashboardsToBeArchived); i++ {
		if _, ok := dashboardsToBeDeletedMap[dashboardsToBeArchived[i].DashboardId]; ok {
			dashboardsToBeArchived = append(dashboardsToBeArchived[:i], dashboardsToBeArchived[i+1:]...)
			i--
		}
	}

	// Archive dashboards
	if len(dashboardsToBeArchived) > 0 {
		err = a.das.UpdateValidatorDashboardsArchiving(ctx, dashboardsToBeArchived)
		if err != nil {
			return err
		}
	}

	// Delete dashboards
	if len(dashboardsToBeDeleted) > 0 {
		err = a.das.RemoveValidatorDashboards(ctx, dashboardsToBeDeleted)
		if err != nil {
			return err
		}
	}

	return nil
}
