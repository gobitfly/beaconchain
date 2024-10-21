package enums

import "github.com/doug-martin/goqu/v9"

// ------------------------------------------------------------
// Notifications Dashboard Table Columns

type NotificationDashboardsColumn int

var _ EnumFactory[NotificationDashboardsColumn] = NotificationDashboardsColumn(0)

const (
	NotificationDashboardChainId NotificationDashboardsColumn = iota
	NotificationDashboardEpoch
	NotificationDashboardDashboardName // sort by dashboard name
	NotificationDashboardDashboardId   // internal use
	NotificationDashboardGroupName     // internal use
	NotificationDashboardGroupId       // internal use
)

func (c NotificationDashboardsColumn) Int() int {
	return int(c)
}

func (NotificationDashboardsColumn) NewFromString(s string) NotificationDashboardsColumn {
	switch s {
	case "chain_id":
		return NotificationDashboardChainId
	case "epoch":
		return NotificationDashboardEpoch
	case "dashboard_name", "dashboard_id": // accepting id for frontend
		return NotificationDashboardDashboardName
	default:
		return NotificationDashboardsColumn(-1)
	}
}

// internal use, used to map to query column names
func (c NotificationDashboardsColumn) ToExpr() OrderableSortable {
	switch c {
	case NotificationDashboardChainId:
		return goqu.C("chain_id")
	case NotificationDashboardEpoch:
		return goqu.C("epoch")
	case NotificationDashboardDashboardName:
		return goqu.C("dashboard_name")
	case NotificationDashboardDashboardId:
		return goqu.C("dashboard_id")
	case NotificationDashboardGroupName:
		return goqu.C("group_name")
	case NotificationDashboardGroupId:
		return goqu.C("group_id")
	default:
		return nil
	}
}

var NotificationsDashboardsColumns = struct {
	ChainId       NotificationDashboardsColumn
	Timestamp     NotificationDashboardsColumn
	DashboardName NotificationDashboardsColumn
	DashboardId   NotificationDashboardsColumn
	GroupName     NotificationDashboardsColumn
	GroupId       NotificationDashboardsColumn
}{
	NotificationDashboardChainId,
	NotificationDashboardEpoch,
	NotificationDashboardDashboardName,
	NotificationDashboardDashboardId,
	NotificationDashboardGroupName,
	NotificationDashboardGroupId,
}

// ------------------------------------------------------------
// Notifications Machines Table Columns

type NotificationMachinesColumn int

var _ EnumFactory[NotificationMachinesColumn] = NotificationMachinesColumn(0)

const (
	NotificationMachineId NotificationMachinesColumn = iota // internal use
	NotificationMachineName
	NotificationMachineThreshold
	NotificationMachineEventType
	NotificationMachineTimestamp
)

func (c NotificationMachinesColumn) Int() int {
	return int(c)
}

func (NotificationMachinesColumn) NewFromString(s string) NotificationMachinesColumn {
	switch s {
	case "machine_name":
		return NotificationMachineName
	case "threshold":
		return NotificationMachineThreshold
	case "event_type":
		return NotificationMachineEventType
	case "timestamp":
		return NotificationMachineTimestamp
	default:
		return NotificationMachinesColumn(-1)
	}
}

// internal use, used to map to query column names
func (c NotificationMachinesColumn) ToExpr() OrderableSortable {
	switch c {
	case NotificationMachineId:
		return goqu.C("machine_id")
	case NotificationMachineName:
		return goqu.C("machine_name")
	case NotificationMachineThreshold:
		return goqu.C("threshold")
	case NotificationMachineEventType:
		return goqu.C("event_type")
	case NotificationMachineTimestamp:
		return goqu.C("epoch")
	default:
		return nil
	}
}

var NotificationsMachinesColumns = struct {
	MachineId   NotificationMachinesColumn
	MachineName NotificationMachinesColumn
	Threshold   NotificationMachinesColumn
	EventType   NotificationMachinesColumn
	Timestamp   NotificationMachinesColumn
}{
	NotificationMachineId,
	NotificationMachineName,
	NotificationMachineThreshold,
	NotificationMachineEventType,
	NotificationMachineTimestamp,
}

// ------------------------------------------------------------
// Notifications Clients Table Columns

type NotificationClientsColumn int

var _ EnumFactory[NotificationClientsColumn] = NotificationClientsColumn(0)

const (
	NotificationClientName NotificationClientsColumn = iota
	NotificationClientTimestamp
)

func (c NotificationClientsColumn) Int() int {
	return int(c)
}

func (NotificationClientsColumn) NewFromString(s string) NotificationClientsColumn {
	switch s {
	case "client_name":
		return NotificationClientName
	case "timestamp":
		return NotificationClientTimestamp
	default:
		return NotificationClientsColumn(-1)
	}
}

// internal use, used to map to query column names
func (c NotificationClientsColumn) ToExpr() OrderableSortable {
	switch c {
	case NotificationClientName:
		return goqu.C("client_name")
	case NotificationClientTimestamp:
		return goqu.C("epoch")
	default:
		return nil
	}
}

var NotificationsClientsColumns = struct {
	ClientName NotificationClientsColumn
	Timestamp  NotificationClientsColumn
}{
	NotificationClientName,
	NotificationClientTimestamp,
}

// ------------------------------------------------------------
// Notifications Rocket Pool Table Columns

type NotificationRocketPoolColumn int

var _ EnumFactory[NotificationRocketPoolColumn] = NotificationRocketPoolColumn(0)

const (
	NotificationRocketPoolTimestamp NotificationRocketPoolColumn = iota
	NotificationRocketPoolEventType
	NotificationRocketPoolNodeAddress
)

func (c NotificationRocketPoolColumn) Int() int {
	return int(c)
}

func (NotificationRocketPoolColumn) NewFromString(s string) NotificationRocketPoolColumn {
	switch s {
	case "timestamp":
		return NotificationRocketPoolTimestamp
	case "event_type":
		return NotificationRocketPoolEventType
	case "node_address":
		return NotificationRocketPoolNodeAddress
	default:
		return NotificationRocketPoolColumn(-1)
	}
}

var NotificationRocketPoolColumns = struct {
	Timestamp   NotificationRocketPoolColumn
	EventType   NotificationRocketPoolColumn
	NodeAddress NotificationRocketPoolColumn
}{
	NotificationRocketPoolTimestamp,
	NotificationRocketPoolEventType,
	NotificationRocketPoolNodeAddress,
}

// ------------------------------------------------------------
// Notifications Networks Table Columns

type NotificationNetworksColumn int

var _ EnumFactory[NotificationNetworksColumn] = NotificationNetworksColumn(0)

const (
	NotificationNetworkTimestamp NotificationNetworksColumn = iota
	NotificationNetworkNetwork                              // internal use
	NotificationNetworkEventType
)

func (c NotificationNetworksColumn) Int() int {
	return int(c)
}

func (NotificationNetworksColumn) NewFromString(s string) NotificationNetworksColumn {
	switch s {
	case "timestamp":
		return NotificationNetworkTimestamp
	case "event_type":
		return NotificationNetworkEventType
	default:
		return NotificationNetworksColumn(-1)
	}
}

// internal use, used to map to query column names
func (c NotificationNetworksColumn) ToExpr() OrderableSortable {
	switch c {
	case NotificationNetworkTimestamp:
		return goqu.C("epoch")
	case NotificationNetworkNetwork:
		return goqu.C("network")
	case NotificationNetworkEventType:
		return goqu.C("event_type")
	default:
		return nil
	}
}

var NotificationNetworksColumns = struct {
	Timestamp NotificationNetworksColumn
	Network   NotificationNetworksColumn
	EventType NotificationNetworksColumn
}{
	NotificationNetworkTimestamp,
	NotificationNetworkNetwork,
	NotificationNetworkEventType,
}

// ------------------------------------------------------------
// Notification Settings Dashboard Table Columns

type NotificationSettingsDashboardColumn int

var _ EnumFactory[NotificationSettingsDashboardColumn] = NotificationSettingsDashboardColumn(0)

const (
	NotificationSettingsDashboardDashboardName NotificationSettingsDashboardColumn = iota
	NotificationSettingsDashboardGroupName
)

func (c NotificationSettingsDashboardColumn) Int() int {
	return int(c)
}

func (NotificationSettingsDashboardColumn) NewFromString(s string) NotificationSettingsDashboardColumn {
	switch s {
	case "dashboard_name", "dashboard_id":
		return NotificationSettingsDashboardDashboardName
	case "group_name":
		return NotificationSettingsDashboardGroupName
	default:
		return NotificationSettingsDashboardColumn(-1)
	}
}

var NotificationSettingsDashboardColumns = struct {
	DashboardName NotificationSettingsDashboardColumn
	GroupName     NotificationSettingsDashboardColumn
}{
	NotificationSettingsDashboardDashboardName,
	NotificationSettingsDashboardGroupName,
}
