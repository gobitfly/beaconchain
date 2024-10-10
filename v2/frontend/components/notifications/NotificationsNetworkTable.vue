<script setup lang="ts">
import type { NotificationNetworksTableRow } from '~/types/api/notifications'
import type { Cursor } from '~/types/datatable'

defineEmits<{ (e: 'openDialog'): void }>()

const { width } = useWindowSize()
const cursor = ref<Cursor>()
const pageSize = ref<number>(10)
const { t: $t } = useTranslation()

const {
  isLoading,
  networkNotifications,
  onSort,
  query,
  setCursor,
  setPageSize,
  setSearch,
} = useNotificationsNetworkStore()
const { overview } = useNotificationsDashboardOverviewStore()

const textNotifications = (eventType: NotificationNetworksTableRow['event_type']) => {
  if (eventType === 'gas_above') return $t('notifications.network.event_type.gas_above')
  if (eventType === 'gas_below') return $t('notifications.network.event_type.gas_below')
  if (eventType === 'new_reward_round') return $t('notifications.network.event_type.new_reward_round')
  if (eventType === 'participation_rate') return $t('notifications.network.event_type.participation_rate')
  logError(`Unknown network notification event_type: ${eventType}`)
  return eventType
}
const textThreshold = (row: NotificationNetworksTableRow) => {
  const {
    event_type,
    threshold,
  } = row
  if (
    event_type === 'gas_above' || event_type === 'gas_below'
  ) {
    return `${formatWeiTo(threshold ?? '0', { unit: 'gwei' })} ${$t('common.units.GWEI')}`
  }
  if (event_type === 'participation_rate') {
    return `${formatToFraction(threshold ?? 0)} %`
  }
  logError(`Unknown network notification event-type: ${row.event_type}`)
  return threshold
}
</script>

<template>
  <div>
    <BcTableControl
      :title="$t('notifications.network.title')"
      @set-search="setSearch"
    >
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            :data="networkNotifications"
            data-key="notification_id"
            :cursor
            :page-size
            :selected-sort="query?.sort"
            :loading="isLoading"
            :add-spacer="true"
            @set-cursor="setCursor"
            @sort="onSort"
            @set-page-size="setPageSize"
          >
            <Column
              field="chain_id"
              header-class="col-header-network"
              body-class="col-network"
            >
              <template #body="slotProps">
                <div class="icon-wrapper">
                  <IconNetwork
                    colored
                    :chain-id="slotProps.data.chain_id"
                    class="icon-network"
                  />
                </div>
              </template>
            </Column>
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
                  type="go-timestamp"
                />
              </template>
            </Column>
            <Column
              field="event_type"
              :sortable="true"
              header-class="col-event_type"
              body-class="col-event_type"
              :header="$t('notifications.col.notification')"
            >
              <template #body="slotProps">
                {{ textNotifications(slotProps.data.event_type) }}
              </template>
            </Column>
            <Column
              field="threshold"
              sortable
              :header="$t('notifications.col.threshold')"
            >
              <template #body="slotProps">
                {{ textThreshold(slotProps.data) }}
              </template>
            </Column>
            <template #empty>
              <NotificationsDashboardsTableEmpty
                v-if="!networkNotifications?.data.length"
                @open-dialog="$emit('openDialog')"
              />
            </template>
            <template #bc-table-footer-right>
              <template v-if="width > 1024">
                {{ $t('notifications.network.footer.subscriptions', { count: overview?.networks_subscription_count }) }}
              </template>
            </template>
          </BcTable>
        </ClientOnly>
      </template>
    </BcTableControl>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/utils.scss";

$breakpoint-sm: 640px;
$breakpoint-lg: 1024px;

:deep(.col-header-network .p-column-header-content){
  justify-content: center;
}
:deep(.col-header-network) {
  @include utils.set-all-width(35px);
  padding-left: 0px !important;
}
:deep(.col-network) {
  @include utils.set-all-width(35px);
  padding-inline: 0px !important;
}
:deep(.col-age) {
  @media (max-width: $breakpoint-lg) {
    @include utils.set-all-width(140px);
  }
  @media (max-width: $breakpoint-sm) {
    @include utils.set-all-width(78px);
  }
  *:not([data-pc-section="sort"]) {
    @include utils.truncate-text;
  }
}
:deep(.col-event_type) {
  @include utils.truncate-text;
  @media (max-width: $breakpoint-lg) {
    @include utils.set-all-width(400px);
  }
  @media (max-width: $breakpoint-sm) {
    @include utils.set-all-width(200px);
  }
}

.icon-wrapper {
  text-align: center;

  .icon-network {
    height: 14px;
    width: 14px;
  }
}
svg {
  flex-shrink: 0;
}
</style>
