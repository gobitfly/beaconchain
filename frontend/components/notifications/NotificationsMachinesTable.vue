<script setup lang="ts">
import type { Cursor } from '~/types/datatable'

defineEmits<{ (e: 'openDialog'): void }>()

const { width } = useWindowSize()
const cursor = ref<Cursor>()
const pageSize = ref<number>(10)
const { t: $t } = useTranslation()

const {
  isLoading,
  machineNotifications,
  onSort,
  query,
  setCursor,
  setPageSize,
  setSearch,
} = useNotificationsMachineStore()

const colsVisible = computed(() => {
  return {
    footer: 1024,
    threshold: width.value > 830,
  }
})
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
            :data="machineNotifications"
            data-key="notification_id"
            :cursor
            :page-size
            :selected-sort="query?.sort"
            :loading="isLoading"
            :add-spacer="true"
            :expandable="!colsVisible.threshold"
            @set-cursor="setCursor"
            @sort="onSort"
            @set-page-size="setPageSize"
          >
            <Column
              field="machine_name"
              sortable
              header-class="col-machine-name"
              body-class="col-machine-name"
              :header="$t('notifications.machine.col.machine_name')"
            >
              <template #body="slotProps">
                <div>
                  {{ slotProps.data.machine_name }}
                </div>
              </template>
            </Column>
            <Column
              v-if="colsVisible.threshold"
              field="threshold"
              sortable
              header-class="col-threshold"
              body-class="col-threshold"
              :header="$t('notifications.machine.col.threshold')"
            >
              <template #body="slotProps">
                <BcFormatPercent
                  v-if="slotProps.data.threshold"
                  :percent="slotProps.data.threshold * 100"
                />
                <span v-else>-</span>
              </template>
            </Column>
            <Column
              field="event_type"
              sortable
              header-class="col-event-type"
              body-class="col-event-type"
              :header="$t('notifications.machine.col.event_type')"
            >
              <template #body="slotProps">
                <div>
                  {{ $t(`notifications.machine.event_type.${slotProps.data.event_type}`, slotProps.data.event_type) }}
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
            <template #expansion="slotProps">
              <div class="expansion">
                <div class="group">
                  <div class="label">
                    {{
                      $t('notifications.machine.col.threshold')
                    }}
                  </div>
                  <BcFormatPercent
                    v-if="slotProps.data.threshold"
                    :percent="slotProps.data.threshold * 100"
                  />
                  <span v-else>-</span>
                </div>
              </div>
            </template>
            <template #empty>
              <NotificationsDashboardsTableEmpty
                v-if="!machineNotifications?.data.length"
                @open-dialog="$emit('openDialog')"
              />
            </template>
            <!-- TODO: implement number of subscriptions -->
            <template #bc-table-footer-right>
              <template v-if="colsVisible">
                {{ $t('notifications.machine.footer.subscriptions', { count: 1 }) }}
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

$breakpoint-sm: 630px;
$breakpoint-md: 780px;
$breakpoint-lg: 1024px;

:deep(.col-event-type),
:deep(.col-machine-name) {
  *:not([data-pc-section="sort"]) {
    @include utils.truncate-text;
  }

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

:deep(.col-threshold) {
  *:not([data-pc-section="sort"]) {
    @include utils.truncate-text;
  }

  @include utils.set-all-width(140px);
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

.expansion {
  background-color: var(--table-header-background);
  padding: 14px 7px;

  .group {
    display: flex;
    gap: var(--padding);

    .label {
      width: 78px;
      font-weight: var(--standard_text_bold_font_weight);
    }
  }
}
</style>
