package enums

// ------------------------------------------------------------
// Notifications Dashboard Table Columns

type NotificationDashboardsColumn int

var _ EnumConvertible[NotificationDashboardsColumn] = NotificationDashboardsColumn(0)

const (
	NotificationDashboardChainId NotificationDashboardsColumn = iota
	NotificationDashboardTimestamp
	NotificationDashboardDashboardName // sort by dashboard name
	NotificationDashboardGroupName     // sort by group name, internal use only
)

func (c NotificationDashboardsColumn) Int() int {
	return int(c)
}

func (NotificationDashboardsColumn) NewFromString(s string) NotificationDashboardsColumn {
	switch s {
	case "chain_id":
		return NotificationDashboardChainId
	case "timestamp":
		return NotificationDashboardTimestamp
	case "dashboard_name", "dashboard_id": // accepting id for frontend
		return NotificationDashboardDashboardName
	default:
		return NotificationDashboardsColumn(-1)
	}
}

// internal use, used to map to query column names
func (c NotificationDashboardsColumn) ToString() string {
	switch c {
	case NotificationDashboardChainId:
		return "chain_id"
	case NotificationDashboardTimestamp:
		return "epoch"
	case NotificationDashboardDashboardName:
		return "dashboard_name"
	case NotificationDashboardGroupName:
		return "group_name"
	default:
		return ""
	}
}

var NotificationsDashboardsColumns = struct {
	ChainId     NotificationDashboardsColumn
	Timestamp   NotificationDashboardsColumn
	DashboardId NotificationDashboardsColumn
	GroupId     NotificationDashboardsColumn
}{
	NotificationDashboardChainId,
	NotificationDashboardTimestamp,
	NotificationDashboardDashboardName,
	NotificationDashboardGroupName,
}

// ------------------------------------------------------------
// Notifications Machines Table Columns

type NotificationMachinesColumn int

var _ EnumFactory[NotificationMachinesColumn] = NotificationMachinesColumn(0)

const (
	NotificationMachineName NotificationMachinesColumn = iota
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

var NotificationsMachinesColumns = struct {
	MachineName NotificationMachinesColumn
	Threshold   NotificationMachinesColumn
	EventType   NotificationMachinesColumn
	Timestamp   NotificationMachinesColumn
}{
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

var NotificationNetworksColumns = struct {
	Timestamp NotificationNetworksColumn
	EventType NotificationNetworksColumn
}{
	NotificationNetworkTimestamp,
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
	DashboardId NotificationSettingsDashboardColumn
	GroupName   NotificationSettingsDashboardColumn
}{
	NotificationSettingsDashboardDashboardName,
	NotificationSettingsDashboardGroupName,
}
