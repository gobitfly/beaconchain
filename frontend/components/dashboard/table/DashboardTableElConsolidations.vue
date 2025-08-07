<script setup lang="ts">
import type { DataTableSortEvent } from 'primevue/datatable'
import type {
  GetValidatorDashboardExecutionLayerConsolidationsResponse,
  VDBConsolidationsElTableRow,
} from '~/types/api/validator_dashboard'
import type {
  Cursor, TableQueryParams,
} from '~/types/datatable'

const {
  elConsolidations,
} = defineProps<{
  elConsolidations?: GetValidatorDashboardExecutionLayerConsolidationsResponse,
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
const v1Domain = useV1Domain()
const emit = defineEmits<{
  (e: 'add-validator'): void,
}>()
</script>

<template>
  <BcTableControl
    :title="$t('dashboard.validator.el_consolidations.title')"
    :search-placeholder="$t('dashboard.validator.el_consolidations.search_placeholder')
    "
    @set-search="setSearch"
  >
    <template #table>
      <ClientOnly fallback-tag="span">
        <BcTable
          :data="addIdentifier(elConsolidations, 'block_queued', 'tx_index_queued', 'itx_index_queued')"
          expandable
          :row-class="(row: VDBConsolidationsElTableRow) => row.status === 'queued' ? 'grayed-out-row' : ''"
          data-key="identifier"
          :selected-sort="query?.sort"
          :cursor="query?.cursor"
          :page-size="query?.limit"
          table-class="dashboard-table-el-consolidations"
          @set-cursor="setCursor"
          @sort="onSort"
          @set-page-size="setPageSize"
        >
          <Column
            sortable
            body-class="dashboard-table-el-consolidations__age-cell"
            field="timestamp"
          >
            <template #header>
              <BcTableAgeHeader />
            </template>
            <template #body="slotProps">
              <BcTableDateTime
                v-if="slotProps.data.timestamp_queued !== undefined"
                :unix-timestamp="slotProps.data.timestamp_queued"
              />
              <span v-else>-</span>
            </template>
          </Column>
          <Column
            v-if="!isMobile"
            field="source"
            :header="$t('table.source')"
            body-class="dashboard-table-el-consolidations__validator-cell"
          >
            <template #body="slotProps">
              <BcIcon
                name="desktop"
                size="sm"
                class="dashboard-table-el-consolidations__desktop-icon"
              />
              <BcLink
                :to="`${v1Domain}/validator/${slotProps.data.source}`"
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
            body-class="dashboard-table-el-consolidations__validator-cell"
          >
            <template #body="slotProps">
              <BcIcon
                name="desktop"
                size="sm"
                class="dashboard-table-el-consolidations__desktop-icon"
              />
              <BcLink
                :to="`${v1Domain}/validator/${slotProps.data.target}`"
                target="_blank"
                class="link"
              >
                {{ slotProps.data.target }}
              </BcLink>
            </template>
          </Column>
          <Column
            v-if="!isMobile"
            field="consolidator"
            :header="$t('table.consolidator')"
          >
            <template #body="slotProps">
              <BcFormatHash
                :hash="slotProps.data.consolidator.hash"
                type="address"
                :no-wrap="true"
                :ens="slotProps.data.consolidator.ens"
              />
            </template>
          </Column>
          <Column
            v-if="!isMobile"
            field="status"
            :header="$t('table.status')"
          >
            <template #body="slotProps">
              <DashboardTableElConsolidationsStatus :status="slotProps.data.status" />
            </template>
          </Column>
          <Column
            v-if="!isMobile"
            field="fee"
            :header="$t('table.consolidation_request_fee')"
          >
            <template #body="slotProps">
              <BcFormatAmount
                :value="slotProps.data.fee"
              />
            </template>
          </Column>
          <Column
            v-if="isMobile"
            field="block_processed"
            :header="$t('dashboard.validator.col.block_processed')"
          >
            <template #body="slotProps">
              <BcLink
                v-if="slotProps.data.block_processed !== undefined"
                :to="`${v1Domain}/block/${slotProps.data.block_processed}`"
                target="_blank"
                class="link"
              >
                <BcFormatNumber :value="slotProps.data.block_processed" />
              </BcLink>
              <span v-else>-</span>
            </template>
          </Column>
          <Column
            v-if="isMobile"
            field="status"
          >
            <template #body="slotProps">
              <DashboardTableElConsolidationsStatus
                :status="slotProps.data.status"
                is-compact
              />
            </template>
          </Column>
          <template #expansion="slotProps">
            <div class="dashboard-table-el-consolidations__details">
              <div class="dashboard-table-el-consolidations__details-grid">
                <div class="dashboard-table-el-consolidations__details-row">
                  <div class="dashboard-table-el-consolidations__details-label">
                    {{ $t("block.col.transaction_hash") }}
                  </div>
                  <BcFormatHash
                    :hash="slotProps.data.tx_hash"
                    :no-wrap="true"
                    type="tx"
                  />
                </div>
                <div class="dashboard-table-el-consolidations__details-row">
                  <div class="dashboard-table-el-consolidations__details-label">
                    {{ $t("dashboard.validator.col.block_queued") }}
                  </div>
                  <BcLink
                    v-if="slotProps.data.block_queued !== undefined"
                    :to="`${v1Domain}/block/${slotProps.data.block_queued}`"
                    target="_blank"
                    class="link"
                  >
                    <BcFormatNumber :value="slotProps.data.block_queued" />
                  </BcLink>
                  <span v-else>-</span>
                </div>
                <div
                  v-if="!isMobile"
                  class="dashboard-table-el-consolidations__details-row"
                >
                  <div class="dashboard-table-el-consolidations__details-label">
                    {{ $t("dashboard.validator.col.block_processed") }}
                  </div>
                  <div class="dashboard-table-el-consolidations__details-value">
                    <BcLink
                      v-if="slotProps.data.block_processed !== undefined"
                      :to="`${v1Domain}/block/${slotProps.data.block_processed}`"
                      target="_blank"
                      class="link"
                    >
                      <BcFormatNumber :value="slotProps.data.block_processed" />
                    </BcLink>
                    <span v-else>-</span>

                    <BcTooltip
                      v-if="slotProps.data.status === 'queued'"
                      tooltip-width="200px"
                      tooltip-text-align="left"
                      class="dashboard-table-el-consolidations__detail-value-tooltip"
                    >
                      <BcIcon name="circle-info" />
                      <template #tooltip>
                        <slot name="tooltip">
                          {{ $t('dashboard.validator.el_consolidations.info_block_processed_estimated') }}
                        </slot>
                      </template>
                    </BcTooltip>
                  </div>
                </div>
                <div
                  v-if="isMobile"
                  class="dashboard-table-el-consolidations__details-row"
                >
                  <div class="dashboard-table-el-consolidations__details-label">
                    {{ $t("table.source") }}
                  </div>
                  <div class="dashboard-table-el-consolidations__details-value">
                    <BcIcon
                      name="desktop"
                      size="sm"
                      class="dashboard-table-el-consolidations__desktop-icon"
                    />
                    <BcLink
                      :to="`${v1Domain}/validator/${slotProps.data.source}`"
                      target="_blank"
                      class="link"
                    >
                      {{ slotProps.data.source }}
                    </BcLink>
                  </div>
                </div>
                <div
                  v-if="isMobile"
                  class="dashboard-table-el-consolidations__details-row"
                >
                  <div class="dashboard-table-el-consolidations__details-label">
                    {{ $t("table.target") }}
                  </div>
                  <div class="dashboard-table-el-consolidations__details-value">
                    <BcIcon
                      name="desktop"
                      size="sm"
                      class="dashboard-table-el-consolidations__desktop-icon"
                    />
                    <BcLink
                      :to="`${v1Domain}/validator/${slotProps.data.target}`"
                      target="_blank"
                      class="link"
                    >
                      {{ slotProps.data.target }}
                    </BcLink>
                  </div>
                </div>
                <div
                  v-if="isMobile"
                  class="dashboard-table-el-consolidations__details-row"
                >
                  <div class="dashboard-table-el-consolidations__details-label">
                    {{ $t("table.consolidator") }}
                  </div>
                  <div class="dashboard-table-el-consolidations__details-label">
                    <BcFormatHash
                      :hash="slotProps.data.consolidator.hash"
                      type="address"
                      :no-wrap="true"
                      :ens="slotProps.data.consolidator.ens"
                    />
                  </div>
                </div>
                <div
                  v-if="isMobile"
                  class="dashboard-table-el-consolidations__details-row"
                >
                  <div class="dashboard-table-el-consolidations__details-label">
                    {{ $t("table.status") }}
                  </div>
                  <div class="dashboard-table-el-consolidations__details-label">
                    <DashboardTableElConsolidationsStatus
                      :status="slotProps.data.status"
                    />
                  </div>
                </div>
                <div
                  v-if="isMobile"
                  class="dashboard-table-el-consolidations__details-row"
                >
                  <div class="dashboard-table-el-consolidations__details-label">
                    {{ $t("table.consolidation_request_fee") }}
                  </div>
                  <div class="dashboard-table-el-consolidations__details-label">
                    <BcFormatAmount
                      :value="slotProps.data.fee"
                    />
                  </div>
                </div>
              </div>
            </div>
          </template>
          <template #empty>
            <DashboardTableAddValidator
              v-if="!hasValidators"
              @add-validator="emit('add-validator')"
            />
          </template>
        </BcTable>
      </ClientOnly>
    </template>
  </BcTableControl>
</template>

<style lang="scss" scoped>
@use '~/assets/css/breakpoints' as *;

.dashboard-table-el-consolidations {
  &__desktop-icon {
    margin-right: var(--padding);
    color: var(--text-color-discreet)
  }

  &__details {
    padding: var(--padding-medium);
    color: var(--container-color);
    font-size: var(--small_text_font_size);
    background-color: var(--container-background);
  }

  &__detail-value-tooltip {
    margin-left: var(--padding);
  }

  &__details-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, max-content));
    gap: var(--padding-medium) calc(var(--padding-large) * 2);

    @media (min-width: $breakpoint-md) {
      grid-template-columns: repeat(4, minmax(0, max-content));
      grid-template-rows: repeat(3, auto);
      grid-auto-flow: column;
    }
  }

  &__details-row {
    display: grid;
    grid-template-columns: subgrid;
    grid-column: span 2;
    column-gap: var(--padding-large);
    align-items: center;
  }

  &__details-label {
    font-weight: var(--standard_text_bold_font_weight);
  }

  &__details-value {
    display: flex;
    max-width: 140px;
  }
}

:deep(.dashboard-table-el-consolidations) {
  .grayed-out-row > td > * {
    opacity: 0.5;
  }

  .dashboard-table-el-consolidations__age-cell {
    padding-top: 0 !important;
    padding-bottom: 0 !important;
    white-space: nowrap;
  }

  .dashboard-table-el-consolidations__validator-cell {
    white-space: nowrap;
  }
}
</style>
