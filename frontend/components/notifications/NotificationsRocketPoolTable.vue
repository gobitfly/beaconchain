<script setup lang="ts">
import type { NotificationRocketPoolTableRow } from '~/types/api/notifications'

defineEmits<{ (e: 'openDialog'): void }>()
const { t: $t } = useTranslation()

const {
  cursor,
  isLoading,
  onSort,
  pageSize,
  query,
  rocketpoolNotifications,
  setCursor,
  setPageSize,
  setSearch,
} = useNotificationsRocketpoolStore()

const { overview } = useNotificationsDashboardOverviewStore()

const { width } = useWindowSize()
const colsVisible = computed(() => {
  return {
    notifications: width.value > 768,
    timestamp: width.value > 640,
  }
})

const getEventTypeName = (eventType: NotificationRocketPoolTableRow['event_type']) => {
  if (eventType === 'collateral_max') return $t('notifications.rocketpool.event_types.collateral_max')
  if (eventType === 'collateral_min') return $t('notifications.rocketpool.event_types.collateral_min')
  if (eventType === 'reward_round') return $t('notifications.rocketpool.event_types.reward_round')
}
</script>

<template>
  <div>
    <BcTableControl
      :title="$t('notifications.tabs.rocketpool')"
      :search-placeholder="$t('notifications.rocketpool.search_placeholder')"
      @set-search="setSearch"
    >
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            :data="rocketpoolNotifications"
            :cursor
            data-key="timestamp"
            :page-size
            :loading="isLoading"
            :selected-sort="query?.sort"
            :expandable="!colsVisible.notifications"
            @set-cursor="setCursor"
            @sort="onSort"
            @set-page-size="setPageSize"
          >
            <Column
              field="timestamp"
              sortable
              header-class="col-age"
              body-class="col-age"
            >
              <template #header>
                <BcTableAgeHeader />
              </template>
              <template #body="slotProps">
                <BcFormatTimePassed
                  :value="slotProps.data.timestamp"
                />
              </template>
            </Column>
            <Column
              sortable
              field="event_type"
              :header="$t('notifications.rocketpool.col.notification')"
            >
              <template #body="slotProps">
                {{ getEventTypeName(slotProps.data.event_type) }}
                <span class="percentage">
                  ({{ formatToPercent(slotProps.data.alert_value ?? 0) }})
                </span>
              </template>
            </Column>
            <Column
              v-if="colsVisible.notifications"
              field="node_address"
              :header="$t('notifications.rocketpool.col.node_address')"
              sortable
            >
              <template #body="slotProps">
                <BcFormatHash
                  :ens="slotProps.data.node.ens"
                  :hash="slotProps.data.node.hash"
                  full
                  type="address"
                  no-copy
                />
              </template>
            </Column>
            <template #expansion="slotProps">
              <div class="expansion">
                <div class="group">
                  <div class="label">
                    {{ $t('notifications.rocketpool.col.node_address') }}
                  </div>
                  <BcFormatHash
                    :ens="slotProps.data.node.ens"
                    :hash="slotProps.data.node.hash"
                    type="address"
                    no-copy
                  />
                </div>
              </div>
            </template>
            <template #empty>
              <LazyNotificationsDashboardsTableEmpty
                v-if="!rocketpoolNotifications?.data.length"
                @open-dialog="$emit('openDialog')"
              />
            </template>
            <template
              v-if="overview?.rocket_pool_subscription_count ?? 0"
              #bc-table-footer-right
            >
              {{ $t('notifications.rocketpool.col.rocketpool_subscription', { count: overview?.rocket_pool_subscription_count }) }}
            </template>
          </BcTable>
        </ClientOnly>
      </template>
    </BcTableControl>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/main.scss";
@use "~/assets/css/utils.scss";

$breakpoint-sm: 630px;
$breakpoint-md: 780px;
$breakpoint-lg: 1024px;

:deep(.col-client-name) {

  @include utils.truncate-text;

  @media (max-width: $breakpoint-lg) {
    @include utils.set-all-width(240px);
  }

  @media (max-width: $breakpoint-md) {
    @include utils.set-all-width(200px);
  }

  @media (max-width: $breakpoint-sm) {
    @include utils.set-all-width(106px);
  }
}

:deep(.col-version) {

  @include utils.truncate-text;

  @media (max-width: $breakpoint-lg) {
    @include utils.set-all-width(140px);
  }

  @media (max-width: $breakpoint-sm) {
    @include utils.set-all-width(78px);
  }

}

:deep(.col-age) {
  @media (max-width: $breakpoint-lg) {
    @include utils.set-all-width(140px);
  }

  @media (max-width: $breakpoint-sm) {
    @include utils.set-all-width(78px);
  }

}

.management-table {
  @include main.container;
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  overflow-y: hidden;
  justify-content: space-between;

  :deep(.p-datatable-wrapper) {
    flex-grow: 1;
  }
}

.icon-wrapper {
  text-align: center;

}

td > .col-notification-name > a{
  text-overflow: ellipsis !important;
}

td > .col-notification-name {
  text-overflow: ellipsis !important;
}

// .percentage {
   // not color contrast ratio not AA compliant
  // color: var(--dark-disabled-grey, #A5A5A5);
// }

.expansion {
  background-color: var(--table-header-background);
  padding: 14px;

  .group {
    display: flex;
    gap: var(--padding);

    .label {
      font-weight: var(--standard_text_bold_font_weight);
    }
  }
}
:deep(.format-hash .prime) {
  color: var(--link-color);
}
</style>
