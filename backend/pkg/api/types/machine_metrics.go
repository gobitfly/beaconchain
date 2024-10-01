package types

type MachineMetricSystem struct {
	Timestamp       uint64 `json:"timestamp,omitempty" faker:"boundary_start=1725166800, boundary_end=1725177600"`
	ExporterVersion string `json:"exporter_version,omitempty"`
	// system
	CpuCores                      uint64 `json:"cpu_cores,omitempty"`
	CpuThreads                    uint64 `json:"cpu_threads,omitempty"`
	CpuNodeSystemSecondsTotal     uint64 `json:"cpu_node_system_seconds_total,omitempty"`
	CpuNodeUserSecondsTotal       uint64 `json:"cpu_node_user_seconds_total,omitempty"`
	CpuNodeIowaitSecondsTotal     uint64 `json:"cpu_node_iowait_seconds_total,omitempty"`
	CpuNodeIdleSecondsTotal       uint64 `json:"cpu_node_idle_seconds_total,omitempty"`
	MemoryNodeBytesTotal          uint64 `json:"memory_node_bytes_total,omitempty"`
	MemoryNodeBytesFree           uint64 `json:"memory_node_bytes_free,omitempty"`
	MemoryNodeBytesCached         uint64 `json:"memory_node_bytes_cached,omitempty"`
	MemoryNodeBytesBuffers        uint64 `json:"memory_node_bytes_buffers,omitempty"`
	DiskNodeBytesTotal            uint64 `json:"disk_node_bytes_total,omitempty"`
	DiskNodeBytesFree             uint64 `json:"disk_node_bytes_free,omitempty"`
	DiskNodeIoSeconds             uint64 `json:"disk_node_io_seconds,omitempty"`
	DiskNodeReadsTotal            uint64 `json:"disk_node_reads_total,omitempty"`
	DiskNodeWritesTotal           uint64 `json:"disk_node_writes_total,omitempty"`
	NetworkNodeBytesTotalReceive  uint64 `json:"network_node_bytes_total_receive,omitempty"`
	NetworkNodeBytesTotalTransmit uint64 `json:"network_node_bytes_total_transmit,omitempty"`
	MiscNodeBootTsSeconds         uint64 `json:"misc_node_boot_ts_seconds,omitempty"`
	MiscOs                        string `json:"misc_os,omitempty"`
	// do not store in bigtable but include them in generated model
	Machine *string `json:"machine,omitempty"`
}

type MachineMetricValidator struct {
	Timestamp       uint64 `json:"timestamp,omitempty" faker:"boundary_start=1725166800, boundary_end=1725177600"`
	ExporterVersion string `json:"exporter_version,omitempty"`
	// process
	CpuProcessSecondsTotal     uint64 `json:"cpu_process_seconds_total,omitempty"`
	MemoryProcessBytes         uint64 `json:"memory_process_bytes,omitempty"`
	ClientName                 string `json:"client_name,omitempty"`
	ClientVersion              string `json:"client_version,omitempty"`
	ClientBuild                uint64 `json:"client_build,omitempty"`
	SyncEth2FallbackConfigured bool   `json:"sync_eth2_fallback_configured,omitempty"`
	SyncEth2FallbackConnected  bool   `json:"sync_eth2_fallback_connected,omitempty"`
	// validator
	ValidatorTotal  uint64 `json:"validator_total,omitempty"`
	ValidatorActive uint64 `json:"validator_active,omitempty"`
	// do not store in bigtable but include them in generated model
	Machine *string `json:"machine,omitempty"`
}

type MachineMetricNode struct {
	Timestamp       uint64 `json:"timestamp,omitempty" faker:"boundary_start=1725166800, boundary_end=1725177600"`
	ExporterVersion string `json:"exporter_version,omitempty"`
	// process
	CpuProcessSecondsTotal     uint64 `json:"cpu_process_seconds_total,omitempty"`
	MemoryProcessBytes         uint64 `json:"memory_process_bytes,omitempty"`
	ClientName                 string `json:"client_name,omitempty"`
	ClientVersion              string `json:"client_version,omitempty"`
	ClientBuild                uint64 `json:"client_build,omitempty"`
	SyncEth2FallbackConfigured bool   `json:"sync_eth2_fallback_configured,omitempty"`
	SyncEth2FallbackConnected  bool   `json:"sync_eth2_fallback_connected,omitempty"`
	// node
	DiskBeaconchainBytesTotal       uint64 `json:"disk_beaconchain_bytes_total,omitempty"`
	NetworkLibp2PBytesTotalReceive  uint64 `json:"network_libp2p_bytes_total_receive,omitempty"`
	NetworkLibp2PBytesTotalTransmit uint64 `json:"network_libp2p_bytes_total_transmit,omitempty"`
	NetworkPeersConnected           uint64 `json:"network_peers_connected,omitempty"`
	SyncEth1Connected               bool   `json:"sync_eth1_connected,omitempty"`
	SyncEth2Synced                  bool   `json:"sync_eth2_synced,omitempty"`
	SyncBeaconHeadSlot              uint64 `json:"sync_beacon_head_slot,omitempty"`
	SyncEth1FallbackConfigured      bool   `json:"sync_eth1_fallback_configured,omitempty"`
	SyncEth1FallbackConnected       bool   `json:"sync_eth1_fallback_connected,omitempty"`
	// do not store in bigtable but include them in generated model
	Machine *string `json:"machine,omitempty"`
}

type MachineMetricsData struct {
	SystemMetrics    []*MachineMetricSystem    `json:"system_metrics" faker:"slice_len=30"`
	ValidatorMetrics []*MachineMetricValidator `json:"validator_metrics" faker:"slice_len=30"`
	NodeMetrics      []*MachineMetricNode      `json:"node_metrics" faker:"slice_len=30"`
}

type GetUserMachineMetricsRespone ApiDataResponse[MachineMetricsData]
