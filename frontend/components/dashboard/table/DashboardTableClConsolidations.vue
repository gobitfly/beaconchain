<script setup lang="ts">
import type { DataTableSortEvent } from 'primevue/datatable'
import type {
  GetValidatorDashboardConsensusLayerConsolidationsResponse,
  VDBConsolidationsClTableRow,
} from '~/types/api/validator_dashboard'
import type {
  Cursor, TableQueryParams,
} from '~/types/datatable'
import BcTableControl from '~/components/bc/table/BcTableControl.vue'

const {
  clConsolidations,
} = defineProps<{
  clConsolidations?: GetValidatorDashboardConsensusLayerConsolidationsResponse,
  isLoading: boolean,
}>()

const { width } = useWindowSize()
const isMobile = computed(() => {
  return width.value < 768
})

const validatorDashboardOverviewStore = useValidatorDashboardOverviewStore()
const {
  hasValidators,
} = storeToRefs(validatorDashboardOverviewStore)

const {
  getTimestampFromSlot,
} = useNetwork()

const query = defineModel<TableQueryParams>('query')

const onSort = (sort: DataTableSortEvent) => {
  query.value = setQuerySort(sort, query.value)
}
const setCursor = (cursor: Cursor) => {
  query.value = setQueryCursor(cursor, query.value)
}
const setPageSize = (limit: number) => {
  query.value = setQueryPageSize(limit, query.value)
}
const setSearch = (value?: string) => {
  query.value = {
    ...query.value,
    search: value,
  }
}
</script>

<template>
  <BcTableControl
    :title="$t('dashboard.validator.cl_consolidations.title')"
    :search-placeholder="$t('dashboard.validator.cl_consolidations.search_placeholder')"
    @set-search="setSearch"
  >
    <template #table>
      <ClientOnly fallback-tag="span">
        <BcTable
          :data="clConsolidations"
          expandable
          :row-class="(row: VDBConsolidationsClTableRow) =>
            row.status === 'queued' ? 'dashboard-table-cl-consolidations__row--grayed-out' : ''"
          data-key="id"
          :selected-sort="query?.sort"
          :cursor="query?.cursor"
          :page-size="query?.limit"
          table-class="dashboard-table-cl-consolidations"
          @set-cursor="setCursor"
          @sort="onSort"
          @set-page-size="setPageSize"
        >
          <Column
            sortable
            field="timestamp"
            body-class="dashboard-table-cl-consolidations__age-cell"
          >
            <template #header>
              <BcTableAgeHeader />
            </template>
            <template #body="slotProps">
              <BcTableDateTime
                v-if="slotProps.data.slot !== undefined"
                :unix-timestamp="getTimestampFromSlot(slotProps.data.slot)"
              />
              <span v-else>-</span>
            </template>
          </Column>
          <Column
            v-if="!isMobile"
            field="source"
            :header="$t('table.source')"
          >
            <template #body="slotProps">
              <BcIcon
                name="desktop"
                size="sm"
                class="dashboard-table-cl-consolidations__desktop-icon"
              />
              <BcLink
                :to="`/validator/${slotProps.data.source}`"
                target="_blank"
                class="link"
              >
                {{ slotProps.data.source }}
              </BcLink>
            </template>
          </Column>
          <Column
            v-if="!isMobile"
            field="target"
            :header="$t('table.target')"
          >
            <template #body="slotProps">
              <BcIcon
                name="desktop"
                size="sm"
                class="dashboard-table-cl-consolidations__desktop-icon"
              />
              <BcLink
                :to="`/validator/${slotProps.data.target}`"
                target="_blank"
                class="link"
              >
                {{ slotProps.data.target }}
              </BcLink>
            </template>
          </Column>
          <Column
            v-if="!isMobile"
            field="status"
            :header="$t('table.status')"
          >
            <template #body="slotProps">
              <DashboardTableClConsolidationsStatus
                :status="slotProps.data.status"
                :reject-reason="slotProps.data.reject_reason"
              />
            </template>
          </Column>
          <Column
            field="fee"
            :header="$t('table.amount')"
          >
            <template #body="slotProps">
              <BcFormatAmount
                :value="slotProps.data.amount"
              />
            </template>
          </Column>
          <Column
            v-if="isMobile"
            field="status"
          >
            <template #body="slotProps">
              <DashboardTableClConsolidationsStatus
                :status="slotProps.data.status"
                is-compact
              />
            </template>
          </Column>
          <template #expansion="slotProps">
            <div class="dashboard-table-cl-consolidations__details">
              <div class="dashboard-table-cl-consolidations__details-grid">
                <div class="dashboard-table-cl-consolidations__details-row">
                  <div class="dashboard-table-cl-consolidations__details-label">
                    {{ $t("dashboard.validator.col.slot_queued") }}
                  </div>
                  <BcLink
                    v-if="slotProps.data.slot_queued !== undefined"
                    :to="`/slot/${slotProps.data.slot_queued}`"
                    target="_blank"
                    class="link"
                  >
                    <BcFormatNumber :value="slotProps.data.slot_queued" />
                  </BcLink>
                  <span v-else>-</span>
                </div>
                <div
                  class="dashboard-table-cl-consolidations__details-row"
                >
                  <div class="dashboard-table-cl-consolidations__details-label">
                    {{ $t("dashboard.validator.col.slot_processed") }}
                  </div>
                  <div class="dashboard-table-cl-consolidations__details-value">
                    <BcLink
                      v-if="slotProps.data.slot_processed !== undefined"
                      :to="`/slot/${slotProps.data.slot_processed}`"
                      target="_blank"
                      class="link"
                    >
                      <BcFormatNumber
                        :value="slotProps.data.slot_processed"
                      />
                    </BcLink>
                    <span v-else>-</span>

                    <BcTooltip
                      v-if="slotProps.data.status === 'queued'"
                      tooltip-width="200px"
                      tooltip-text-align="left"
                      class="dashboard-table-cl-consolidations__detail-value-tooltip"
                    >
                      <BcIcon name="circle-info" />
                      <template #tooltip>
                        <slot name="tooltip">
                          {{ $t('dashboard.validator.cl_consolidations.info_slot_processed_estimated') }}
                        </slot>
                      </template>
                    </BcTooltip>
                  </div>
                </div>
                <div
                  v-if="isMobile"
                  class="dashboard-table-cl-consolidations__details-row"
                >
                  <div class="dashboard-table-cl-consolidations__details-label">
                    {{ $t("table.source") }}
                  </div>
                  <div class="dashboard-table-cl-consolidations__details-value">
                    <BcIcon
                      name="desktop"
                      size="sm"
                      class="dashboard-table-cl-consolidations__desktop-icon"
                    />
                    <BcLink
                      :to="`/validator/${slotProps.data.source}`"
                      target="_blank"
                      class="link"
                    >
                      {{ slotProps.data.source }}
                    </BcLink>
                  </div>
                </div>
                <div
                  v-if="isMobile"
                  class="dashboard-table-cl-consolidations__details-row"
                >
                  <div class="dashboard-table-cl-consolidations__details-label">
                    {{ $t("table.target") }}
                  </div>
                  <div class="dashboard-table-cl-consolidations__details-value">
                    <BcIcon
                      name="desktop"
                      size="sm"
                      class="dashboard-table-cl-consolidations__desktop-icon"
                    />
                    <BcLink
                      :to="`/validator/${slotProps.data.target}`"
                      target="_blank"
                      class="link"
                    >
                      {{ slotProps.data.target }}
                    </BcLink>
                  </div>
                </div>
                <div
                  v-if="isMobile"
                  class="dashboard-table-cl-consolidations__details-row"
                >
                  <div class="dashboard-table-cl-consolidations__details-label">
                    {{ $t("table.status") }}
                  </div>
                  <div class="dashboard-table-cl-consolidations__details-label">
                    <DashboardTableClConsolidationsStatus
                      :status="slotProps.data.status"
                      :reject-reason="slotProps.data.reject_reason"
                    />
                  </div>
                </div>
              </div>
            </div>
          </template>
          <template #empty>
            <DashboardTableAddValidator v-if="!hasValidators" />
          </template>
        </BcTable>
      </ClientOnly>
    </template>
  </BcTableControl>
</template>

<style lang="scss" scoped>
@use '~/assets/css/breakpoints' as *;

.dashboard-table-cl-consolidations__desktop-icon {
  margin-right: var(--padding);
  color: var(--text-color-discreet)
}

.dashboard-table-cl-consolidations__details {
  padding: var(--padding-medium);
  color: var(--container-color);
  font-size: var(--small_text_font_size);
  background-color: var(--container-background);
}

.dashboard-table-cl-consolidations__details-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, max-content));
  gap: var(--padding-medium) calc(var(--padding-large) * 2);

  @media (min-width: $breakpoint-md) {
    grid-template-columns: repeat(4, minmax(0, max-content));
    grid-template-rows: repeat(2, auto);
    grid-auto-flow: column;
  }
}

.dashboard-table-cl-consolidations__details-row {
  display: grid;
  grid-template-columns: subgrid;
  grid-column: span 2;
  column-gap: var(--padding-large);
  align-items: center;
}

.dashboard-table-cl-consolidations__details-label {
  display: flex;
  font-weight: var(--standard_text_bold_font_weight);
}

.dashboard-table-cl-consolidations__details-value {
  display: flex;
  max-width: 140px;
}

.dashboard-table-cl-consolidations__detail-value-tooltip {
  margin-left: var(--padding);
}

:deep(.dashboard-table-cl-consolidations) {
  .dashboard-table-cl-consolidations__row--grayed-out {
    opacity: 0.5;
  }

  .dashboard-table-cl-consolidations__age-cell {
    padding-top: 0 !important;
    padding-bottom: 0 !important;
  }
}
</style>
