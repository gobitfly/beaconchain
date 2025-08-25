<script setup lang="ts">
import type { NotificationMachinesTableRow } from '~/types/api/notifications'

const { machinesSubscriptionCount } = defineProps<{
  machinesSubscriptionCount?: number,
}>()

defineEmits<{ (e: 'openDialog'): void }>()

const { width } = useWindowSize()
const { t: $t } = useTranslation()

const query = useDefaultQuery({
  limit: 10, sort: 'timestamp:desc',
})
const {
  data: machineNotifications,
  status,
} = useApi('/api/bff/users/me/notifications/machines', {
  query,
})

const colsVisible = computed(() => {
  return {
    footer: 1024,
    threshold: width.value > 830,
  }
})
const machineEvent = (eventType: NotificationMachinesTableRow['event_type']) => {
  if (eventType === 'cpu') return $t('notifications.machine.event_type.cpu_overheated')
  if (eventType === 'memory') return $t('notifications.machine.event_type.high_memory_usage')
  if (eventType === 'offline') return $t('notifications.machine.event_type.machine_offline')
  if (eventType === 'storage') return $t('notifications.machine.event_type.no_storage')
}
</script>

<template>
  <div>
    <BcTableControl
      v-model:search="query.search"
      :title="$t('notifications.machine.title')"
      :search-placeholder="$t('notifications.machine.search_placeholder')"
    >
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            :data="machineNotifications"
            :query
            data-key="machine_name"
            :is-loading="status === 'pending'"
            :add-spacer="true"
            :expandable="!colsVisible.threshold"
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
                <BaseFormatPercent
                  v-if="slotProps.data.threshold"
                  :value="slotProps.data.threshold"
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
                  {{ machineEvent(slotProps.data.event_type) }}
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
                <BcTableDateTime
                  :unix-timestamp="slotProps.data.timestamp"
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
                  <BaseFormatPercent
                    v-if="slotProps.data.threshold"
                    :value="slotProps.data.threshold"
                  />
                  <span v-else>-</span>
                </div>
              </div>
            </template>
            <template #empty>
              <NotificationsTableEmpty
                v-if="!machineNotifications?.data.length"
                @open-dialog="$emit('openDialog')"
              />
            </template>
            <template #bc-table-footer-right>
              <template v-if="colsVisible">
                {{ $t('notifications.machine.footer.subscriptions', { count: machinesSubscriptionCount }) }}
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
  @include utils.set-all-width(140px);
  *:not([data-pc-section="sort"]) {
    @include utils.truncate-text;
  }
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
