<script setup lang="ts">
import type { DataTableSortEvent } from 'primevue/datatable'
import type { VDBBlocksTableRow } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { Cursor, TableQueryParams } from '~/types/datatable'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import { useValidatorDashboardBlocksStore } from '~/stores/dashboard/useValidatorDashboardBlocksStore'

// TODO: replace with dashboardKey provider once it's merged
interface Props {
  dashboardKey: DashboardKey
}
const props = defineProps<Props>()

const cursor = ref<Cursor>()
const pageSize = ref<number>(5)
const { t: $t } = useI18n()

const { blocks, query: lastQuery, getBlocks } = useValidatorDashboardBlocksStore()
const { value: query, bounce: setQuery } = useDebounceValue<TableQueryParams | undefined>(undefined, 500)

const { overview } = useValidatorDashboardOverviewStore()

const { width } = useWindowSize()
const colsVisible = computed(() => {
  return {
    duty: width.value > 1180,
    rewards: width.value >= 860,
    elRewards: width.value >= 740,
    age: width.value >= 620
  }
})

const loadData = (query?: TableQueryParams) => {
  if (!query) {
    query = { limit: pageSize.value }
  }
  setQuery(query, true, true)
}

watch(() => props.dashboardKey, () => {
  loadData()
}, { immediate: true })

watch(query, (q) => {
  if (q) {
    getBlocks(props.dashboardKey, q)
  }
}, { immediate: true })

const groupNameLabel = (groupId?: number) => {
  // Todo: use getGroupLabel funciton once Rewards Table PR is merged.
  if (groupId === undefined) {
    return ''
  }
  const group = overview.value?.groups?.find(g => g.id === groupId)
  if (!group) {
    return `${groupId}` // fallback if we could not match the group name
  }
  return `${group.name}`
}

const onSort = (sort: DataTableSortEvent) => {
  loadData(setQuerySort(sort, lastQuery.value))
}

const setCursor = (value: Cursor) => {
  cursor.value = value
  loadData(setQueryCursor(value, lastQuery.value))
}

const setPageSize = (value: number) => {
  pageSize.value = value
  loadData(setQueryPageSize(value, lastQuery.value))
}

const setSearch = (value?: string) => {
  loadData(setQuerySearch(value, lastQuery.value))
}

const getRowClass = (row: VDBBlocksTableRow) => {
  if (row.status === 'scheduled') {
    return 'future-row'
  }
}

const isRowExpandable = (row: VDBBlocksTableRow) => {
  return row.status !== 'scheduled'
}

</script>
<template>
  <div>
    <BcTableControl
      :title="$t('dashboard.validator.blocks.title')"
      :search-placeholder="$t('dashboard.validator.blocks.search_placeholder')"
      @set-search="setSearch"
    >
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            :data="blocks"
            data-key="epoch"
            :expandable="true"
            class="block-table"
            :cursor="cursor"
            :page-size="pageSize"
            :row-class="getRowClass"
            :add-spacer="true"
            :is-row-expandable="isRowExpandable"
            @set-cursor="setCursor"
            @sort="onSort"
            @set-page-size="setPageSize"
          >
            <Column field="proposer" :sortable="true" :header="$t('block.col.proposer')">
              <template #body="slotProps">
                <NuxtLink
                  :to="`/validator/${slotProps.data.proposer}`"
                  target="_blank"
                  class="link"
                  :no-prefetch="true"
                >
                  <BcFormatNumber :value="slotProps.data.proposer" />
                </NuxtLink>
              </template>
            </Column>
            <Column
              field="group_id"
              body-class="group-id"
              header-class="group-id"
              :header="$t('dashboard.validator.col.group')"
            >
              <template #body="slotProps">
                <span>
                  {{ groupNameLabel(slotProps.data.group_id) }}
                </span>
              </template>
            </Column>
            <Column field="slot" :sortable="true" :header="$t('common.slot')">
              <template #body="slotProps">
                <NuxtLink :to="`/slot/${slotProps.data.slot}`" target="_blank" class="link" :no-prefetch="true">
                  <BcFormatNumber :value="slotProps.data.slot" />
                </NuxtLink>
              </template>
            </Column>
            <Column field="block" :sortable="true" :header="$t('common.block')">
              <template #body="slotProps">
                <NuxtLink :to="`/block/${slotProps.data.block}`" target="_blank" class="link" :no-prefetch="true">
                  <BcFormatNumber :value="slotProps.data.block" />
                </NuxtLink>
              </template>
            </Column>
            <Column v-if="colsVisible.age" :sortable="true" field="age" body-class="age" header-class="age">
              <template #header>
                <BcTableAgeHeader />
              </template>
              <template #body="slotProps">
                <BcFormatTimePassed class="time-passed" :value="slotProps.data.epoch" />
              </template>
            </Column>
            <Column field="status" :header="$t('dashboard.validator.col.status')">
              <template #body="slotProps">
                <!--TODO: use status render once merged-->
                {{ slotProps.data.status }}
              </template>
            </Column>
            <Column
              v-if="colsVisible.rewards"
              field="reward_cl"
              body-class="reward"
              header-class="reward"
              :header="$t('dashboard.validator.col.rewards')"
            >
              <template #body="slotProps">
                <BcTooltip class="combine-rewards">
                  <BcFormatValue :value="slotProps.data.reward?.el" :use-colors="true" :options="{ addPlus: true }" />
                  <BcFormatValue :value="slotProps.data.reward?.cl" :use-colors="true" :options="{ addPlus: true }" />
                  <template #tooltip>
                    <div>
                      <div class="tt-row">
                        <span>{{ $t('dashboard.validator.blocks.el_rewards') }}: </span>
                        <BcFormatValue
                          :value="slotProps.data.reward?.el"
                          :use-colors="true"
                          :options="{ addPlus: true }"
                        />
                      </div>
                      <div class="tt-row">
                        <span>{{ $t('dashboard.validator.blocks.cl_rewards') }}: </span>
                        <BcFormatValue
                          :value="slotProps.data.reward?.cl"
                          :use-colors="true"
                          :options="{ addPlus: true }"
                        />
                      </div>
                    </div>
                  </template>
                </BcTooltip>
              </template>
            </Column>
            <template #expansion="slotProps">
              todo {{ slotProps }}
            </template>
          </BcTable>
        </ClientOnly>
      </template>
      <template #chart>
        <div class="chart-container">
          <!--TODO: chart-->
        </div>
      </template>
    </BcTableControl>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/utils.scss";

:deep(.block-table) {
  --col-width: 154px;

  .epoch {
    // TODO: use utils.set-all-width once merged
    // @include utils.set-all-width(80px);
  }

  .group_id,
  .reward {
    // @include utils.set-all-width(120px);
  }

  .time-passed {
    white-space: nowrap;
  }

  tr:not(.p-datatable-row-expansion) {
    @media (max-width: 1300px) {
      .duty {
        // @include utils.set-all-width(300px);
      }
    }
  }

  tr:has(+.total-row) {
    td {
      border-bottom-color: var(--primary-color);
    }
  }

  .future-row {
    td {

      >div,
      >span {
        opacity: 0.5;
      }
    }
  }
}

.tt-row {
  display: flex;
  flex-wrap: nowrap;
  white-space: nowrap;
  gap: 3px;
}

.combine-rewards {
  display: inline-flex;
  flex-direction: column;

  >div:last-child {
    font-size: var(--small_text_font_size);
  }
}

.chart-container {
  width: 100%;
  height: 625px;
}
</style>
