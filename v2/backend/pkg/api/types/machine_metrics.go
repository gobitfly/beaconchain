package types

import "github.com/gobitfly/beaconchain/pkg/commons/types"

type MachineMetricsData struct {
	SystemMetrics    []*types.MachineMetricSystem    `json:"system_metrics" faker:"slice_len=30"`
	ValidatorMetrics []*types.MachineMetricValidator `json:"validator_metrics" faker:"slice_len=30"`
	NodeMetrics      []*types.MachineMetricNode      `json:"node_metrics" faker:"slice_len=30"`
}

type GetUserMachineMetricsRespone ApiDataResponse[MachineMetricsData]
