import type { ApiDataResponse } from '~/types/api/common';

export interface NotificationsOverview {
  id: number;
  EmailNotifications: boolean;
  EmailLimitCount: number;
  pushNotifications: boolean;
  mostNotifications30d: {
    providers: string[];
    abo: string[];
  };
  mostNotifications24h: {
    Email: number;
    Webhook: number;
    Push: number;
  };
}

export type GetNotificationsOverviewResponse = ApiDataResponse<NotificationsOverview>;

export type NotificationsOverviewResponse = ApiDataResponse<NotificationsOverview>;
