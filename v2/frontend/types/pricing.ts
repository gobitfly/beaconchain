// TODO: These structs are only temporary, they will be defined in the backend and automatically be added to a file in types/api/
// Once this is done, delete this file (pricing.ts)
// Also see mockPremiumPlanResponse

export interface PremiumPlan {
  Name: string; // TODO: Was not part of the struct draft coming from backend team
  AdFree: boolean;
  ValidatorDashboards: number;
  ValidatorsPerDashboard: number;
  ValidatorGroupsPerDashboard: number;
  ShareCustomDashboards: boolean;
  ManageDashboardViaApi: boolean;
  HeatmapHistorySeconds: number;
  SummaryChartHistorySeconds: number;
  EmailNotificationsPerDay: number;
  ConfigureNotificationsViaApi: boolean;
  ValidatorGroupNotifications: number;
  WebhookEndpoints: number;
  MobileAppCustomThemes: boolean;
  MobileAppWidget: boolean;
  MonitorMachines: number;
  MachineMonitoringHistorySeconds: number;
  CustomMachineAlerts: boolean;
}

export interface PremiumPlanAPIresponse {
  data?: PremiumPlan[]
}
