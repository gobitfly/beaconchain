// Code generated by tygo. DO NOT EDIT.
/* eslint-disable */
import type { ApiDataResponse } from './common'

//////////
// source: machine_metrics.go

export interface MachineMetricSystem {
  timestamp?: number /* uint64 */;
  exporter_version?: string;
  /**
   * system
   */
  cpu_cores?: number /* uint64 */;
  cpu_threads?: number /* uint64 */;
  cpu_node_system_seconds_total?: number /* uint64 */;
  cpu_node_user_seconds_total?: number /* uint64 */;
  cpu_node_iowait_seconds_total?: number /* uint64 */;
  cpu_node_idle_seconds_total?: number /* uint64 */;
  memory_node_bytes_total?: number /* uint64 */;
  memory_node_bytes_free?: number /* uint64 */;
  memory_node_bytes_cached?: number /* uint64 */;
  memory_node_bytes_buffers?: number /* uint64 */;
  disk_node_bytes_total?: number /* uint64 */;
  disk_node_bytes_free?: number /* uint64 */;
  disk_node_io_seconds?: number /* uint64 */;
  disk_node_reads_total?: number /* uint64 */;
  disk_node_writes_total?: number /* uint64 */;
  network_node_bytes_total_receive?: number /* uint64 */;
  network_node_bytes_total_transmit?: number /* uint64 */;
  misc_node_boot_ts_seconds?: number /* uint64 */;
  misc_os?: string;
  /**
   * do not store in bigtable but include them in generated model
   */
  machine?: string;
}
export interface MachineMetricValidator {
  timestamp?: number /* uint64 */;
  exporter_version?: string;
  /**
   * process
   */
  cpu_process_seconds_total?: number /* uint64 */;
  memory_process_bytes?: number /* uint64 */;
  client_name?: string;
  client_version?: string;
  client_build?: number /* uint64 */;
  sync_eth2_fallback_configured?: boolean;
  sync_eth2_fallback_connected?: boolean;
  /**
   * validator
   */
  validator_total?: number /* uint64 */;
  validator_active?: number /* uint64 */;
  /**
   * do not store in bigtable but include them in generated model
   */
  machine?: string;
}
export interface MachineMetricNode {
  timestamp?: number /* uint64 */;
  exporter_version?: string;
  /**
   * process
   */
  cpu_process_seconds_total?: number /* uint64 */;
  memory_process_bytes?: number /* uint64 */;
  client_name?: string;
  client_version?: string;
  client_build?: number /* uint64 */;
  sync_eth2_fallback_configured?: boolean;
  sync_eth2_fallback_connected?: boolean;
  /**
   * node
   */
  disk_beaconchain_bytes_total?: number /* uint64 */;
  network_libp2p_bytes_total_receive?: number /* uint64 */;
  network_libp2p_bytes_total_transmit?: number /* uint64 */;
  network_peers_connected?: number /* uint64 */;
  sync_eth1_connected?: boolean;
  sync_eth2_synced?: boolean;
  sync_beacon_head_slot?: number /* uint64 */;
  sync_eth1_fallback_configured?: boolean;
  sync_eth1_fallback_connected?: boolean;
  /**
   * do not store in bigtable but include them in generated model
   */
  machine?: string;
}
export interface MachineMetricsData {
  system_metrics: (MachineMetricSystem | undefined)[];
  validator_metrics: (MachineMetricValidator | undefined)[];
  node_metrics: (MachineMetricNode | undefined)[];
}
export type GetUserMachineMetricsRespone = ApiDataResponse<MachineMetricsData>;
