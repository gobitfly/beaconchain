import type { ApiDataResponse } from '~/types/api/common'

export interface NotificationsOverview {
    // TODO: Update types with given structure.
    id: number;
    EmailNotifications: boolean;
    pushNotifications: boolean;
    mostNotifications30d: number;
    mostNotifications24h: number;
  }

export type GetNotificationsOverviewResponse = ApiDataResponse<NotificationsOverview>;

export type NotificationsOverviewResponse = ApiDataResponse<NotificationsOverview>;
