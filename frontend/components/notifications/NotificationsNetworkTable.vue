<script setup lang="ts">
import type { Cursor } from '~/types/datatable'

defineEmits<{ (e: 'openDialog'): void }>()

const { width } = useWindowSize()
const cursor = ref<Cursor>()
const pageSize = ref<number>(10)
const { t: $t } = useTranslation()

const {
  onSort,
  setCursor,
  setPageSize,
  setSearch,
  networkNotifications,
  query,
  isLoading,
} = useNotificationsNetworkStore()
</script>

<template>
  <div>
    <BcTableControl
      :title="$t('notifications.dashboards.title')"
      :search-placeholder="$t('notifications.dashboards.search_placeholder')"
      @set-search="setSearch"
    >
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            :data="networkNotifications"
            data-key="notification_id"
            :cursor="cursor"
            :page-size="pageSize"
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
                <I18nT
                  :keypath="`notifications.network.event_type.${slotProps.data.event_type}`"
                  scope="global"
                  tag="span"
                >
                  <template #_link>
                    <BcFormatValue
                      v-if="slotProps.data.event_type.includes('gas')"
                      :value="slotProps.data.alert_value"
                    />
                    <BcFormatPercent
                      v-else
                      :percent="Number(slotProps.data.alert_value) * 100"
                    />
                  </template>
                </I18nT>
              </template>
            </Column>
            <template #empty>
              <NotificationsDashboardsTableEmpty
                v-if="!networkNotifications?.data.length"
                @open-dialog="$emit('openDialog')"
              />
            </template>
            <!-- TODO: implement number of subscriptions -->
            <template #bc-table-footer-right>
              <template v-if="width > 1024">
                {{ $t('notifications.network.footer.subscriptions', { count: 1 }) }}
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
