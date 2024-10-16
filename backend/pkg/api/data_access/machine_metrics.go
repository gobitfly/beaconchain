package dataaccess

import (
	"context"
	"fmt"

	"github.com/gobitfly/beaconchain/pkg/api/types"
	ctypes "github.com/gobitfly/beaconchain/pkg/commons/types"
	"golang.org/x/sync/errgroup"
)

type MachineRepository interface {
	GetUserMachineMetrics(context context.Context, userID uint64, limit uint64, offset uint64) (*types.MachineMetricsData, error)
}

func (d *DataAccessService) GetUserMachineMetrics(ctx context.Context, userID uint64, limit uint64, offset uint64) (*types.MachineMetricsData, error) {
	resp := &types.MachineMetricsData{
		SystemMetrics:    []*types.MachineMetricSystem{},
		ValidatorMetrics: []*types.MachineMetricValidator{},
		NodeMetrics:      []*types.MachineMetricNode{},
	}

	g := errgroup.Group{}
	g.Go(func() error {
		nodeData, err := d.bigtable.GetMachineMetricsNode(ctypes.UserId(userID), int(limit), int(offset))
		if err != nil {
			return fmt.Errorf("failed to get machine metrics for user %d with limit %d and offset %d: %w", userID, limit, offset, err)
		}
		for _, entry := range nodeData {
			resp.NodeMetrics = append(resp.NodeMetrics, &types.MachineMetricNode{
				Timestamp:                       entry.Timestamp,
				ExporterVersion:                 entry.ExporterVersion,
				CpuProcessSecondsTotal:          entry.CpuProcessSecondsTotal,
				MemoryProcessBytes:              entry.MemoryProcessBytes,
				ClientName:                      entry.ClientName,
				ClientVersion:                   entry.ClientVersion,
				ClientBuild:                     entry.ClientBuild,
				SyncEth2FallbackConfigured:      entry.SyncEth2FallbackConfigured,
				SyncEth2FallbackConnected:       entry.SyncEth2FallbackConnected,
				DiskBeaconchainBytesTotal:       entry.DiskBeaconchainBytesTotal,
				NetworkLibp2PBytesTotalReceive:  entry.NetworkLibp2PBytesTotalReceive,
				NetworkLibp2PBytesTotalTransmit: entry.NetworkLibp2PBytesTotalTransmit,
				NetworkPeersConnected:           entry.NetworkPeersConnected,
				SyncEth1Connected:               entry.SyncEth1Connected,
				SyncEth2Synced:                  entry.SyncEth2Synced,
				SyncBeaconHeadSlot:              entry.SyncBeaconHeadSlot,
				SyncEth1FallbackConfigured:      entry.SyncEth1FallbackConfigured,
				SyncEth1FallbackConnected:       entry.SyncEth1FallbackConnected,
				Machine:                         entry.Machine,
			})
		}
		return nil
	})

	g.Go(func() error {
		validatorData, err := d.bigtable.GetMachineMetricsValidator(ctypes.UserId(userID), int(limit), int(offset))
		if err != nil {
			return fmt.Errorf("failed to get validator metrics for user %d with limit %d and offset %d: %w", userID, limit, offset, err)
		}
		for _, entry := range validatorData {
			resp.ValidatorMetrics = append(resp.ValidatorMetrics, &types.MachineMetricValidator{
				Timestamp:                  entry.Timestamp,
				ExporterVersion:            entry.ExporterVersion,
				CpuProcessSecondsTotal:     entry.CpuProcessSecondsTotal,
				MemoryProcessBytes:         entry.MemoryProcessBytes,
				ClientName:                 entry.ClientName,
				ClientVersion:              entry.ClientVersion,
				ClientBuild:                entry.ClientBuild,
				SyncEth2FallbackConfigured: entry.SyncEth2FallbackConfigured,
				SyncEth2FallbackConnected:  entry.SyncEth2FallbackConnected,
				ValidatorTotal:             entry.ValidatorTotal,
				ValidatorActive:            entry.ValidatorActive,
				Machine:                    entry.Machine,
			})
		}
		return nil
	})

	g.Go(func() error {
		systemData, err := d.bigtable.GetMachineMetricsSystem(ctypes.UserId(userID), int(limit), int(offset))
		if err != nil {
			return fmt.Errorf("failed to get system metrics for user %d with limit %d and offset %d: %w", userID, limit, offset, err)
		}
		for _, entry := range systemData {
			resp.SystemMetrics = append(resp.SystemMetrics, &types.MachineMetricSystem{
				Timestamp:                     entry.Timestamp,
				ExporterVersion:               entry.ExporterVersion,
				DiskNodeWritesTotal:           entry.DiskNodeWritesTotal,
				NetworkNodeBytesTotalReceive:  entry.NetworkNodeBytesTotalReceive,
				NetworkNodeBytesTotalTransmit: entry.NetworkNodeBytesTotalTransmit,
				MiscNodeBootTsSeconds:         entry.MiscNodeBootTsSeconds,
				MiscOs:                        entry.MiscOs,
				CpuCores:                      entry.CpuCores,
				CpuThreads:                    entry.CpuThreads,
				CpuNodeSystemSecondsTotal:     entry.CpuNodeSystemSecondsTotal,
				CpuNodeUserSecondsTotal:       entry.CpuNodeUserSecondsTotal,
				MemoryNodeBytesTotal:          entry.MemoryNodeBytesTotal,
				MemoryNodeBytesFree:           entry.MemoryNodeBytesFree,
				MemoryNodeBytesCached:         entry.MemoryNodeBytesCached,
				MemoryNodeBytesBuffers:        entry.MemoryNodeBytesBuffers,
				DiskNodeBytesTotal:            entry.DiskNodeBytesTotal,
				DiskNodeBytesFree:             entry.DiskNodeBytesFree,
				DiskNodeIoSeconds:             entry.DiskNodeIoSeconds,
				DiskNodeReadsTotal:            entry.DiskNodeReadsTotal,
				CpuNodeIowaitSecondsTotal:     entry.CpuNodeIowaitSecondsTotal,
				CpuNodeIdleSecondsTotal:       entry.CpuNodeIdleSecondsTotal,
				Machine:                       entry.Machine,
			})
		}
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return resp, nil
}
