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
    slot: width.value > 960,
    age: width.value > 880,
    recipient: width.value > 800,
    status: width.value > 700,
    mobileStatus: width.value < 1000,
    rewards: width.value > 600,
    groupSort: width.value > 400
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
              :sortable="colsVisible.groupSort"
              :header="$t('dashboard.validator.col.group')"
            >
              <template #body="slotProps">
                <span>
                  {{ groupNameLabel(slotProps.data.group_id) }}
                </span>
              </template>
            </Column>
            <Column v-if="colsVisible.slot" field="slot" :sortable="true" :header="$t('common.slot')">
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
            <Column
              v-if="colsVisible.status"
              field="status"
              :sortable="!colsVisible.mobileStatus"
              :header="$t('dashboard.validator.col.status')"
              :body-class="{ 'status-mobile': colsVisible.mobileStatus }"
            >
              <template #body="slotProps">
                <!--TODO: use status render once merged-->
                <span :class="{ 'status-mobile': colsVisible.mobileStatus }" />
                {{ slotProps.data.status }}
              </template>
            </Column>
            <Column
              v-if="colsVisible.rewards"
              field="reward_cl"
              :sortable="true"
              body-class="reward"
              header-class="reward"
              :header="$t('dashboard.validator.col.rewards')"
            >
              <template #body="slotProps">
                <BlockTableRewardItem :reward="slotProps.data.reward" />
              </template>
            </Column>
            <template #expansion="slotProps">
              <div class="expansion">
                <div v-if="!colsVisible.slot" class="row">
                  <div class="label">
                    {{ $t('common.slot') }}:
                  </div>
                  <NuxtLink :to="`/slot/${slotProps.data.slot}`" target="_blank" class="link" :no-prefetch="true">
                    <BcFormatNumber :value="slotProps.data.slot" />
                  </NuxtLink>
                </div>
                <div class="row">
                  <div class="label">
                    {{ $t('common.epoch') }}:
                  </div>
                  <NuxtLink :to="`/epoch/${slotProps.data.block}`" target="_blank" class="link" :no-prefetch="true">
                    <BcFormatNumber :value="slotProps.data.epoch" />
                  </NuxtLink>
                </div>
                <div v-if="!colsVisible.age" class="row">
                  <div class="label">
                    <BcTableAgeHeader />
                  </div>
                  <BcFormatTimePassed class="time-passed" :value="slotProps.data.epoch" />
                </div>
                <div v-if="!colsVisible.status" class="row">
                  <div class="label">
                    {{ $t('dashboard.validator.col.status') }}:
                  </div>
                  <div class="value">
                    <!--TODO: use status render once merged-->
                    {{ slotProps.data.status }}
                  </div>
                </div>
                <div v-if="!colsVisible.rewards" class="row">
                  <div class="label">
                    {{ $t('dashboard.validator.col.rewards') }}:
                  </div>
                  <BlockTableRewardItem :reward="slotProps.data.reward" />
                </div>
                <div class="row">
                  <div class="label">
                    {{ $t('block.col.graffiti') }}:
                  </div>
                  <div class="value">
                    {{ slotProps.data.graffiti }}
                  </div>
                </div>
              </div>
            </template>
          </BcTable>
        </ClientOnly>
      </template>
    </BcTableControl>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/utils.scss";

:deep(.block-table) {

  .group-id {
    //TODO: @include utils.set-all-width(120px);
    max-width: 120px;
    width: 120px;
    min-width: 120px;
    @include utils.truncate-text;
  }

  @media (max-width: 399px) {
    .group-id {
      //TODO: @include utils.set-all-width(80px);
      max-width: 80px;
      width: 80px;
      min-width: 80px;
      @include utils.truncate-text;
    }
  }

  .status-mobile {
    //TODO: @include utils.set-all-width(40px);
    max-width: 40px;
    width: 40px;
    min-width: 40px;
    @include utils.truncate-text;
  }

  .time-passed {
    white-space: nowrap;
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

.expansion {
  display: flex;
  align-items: center;
  flex-direction: column;
  gap: var(--padding-small);
  padding: var(--padding) 41px;
  background: var(--table-header-background);
  font-size: var(--small_text_font_size);

  .row {

    display: flex;
    align-items: center;

    .label {
      width: 90px;
      font-weight: var(--small_text_bold_font_weight);
      flex-shrink: 0;
    }

    .value {
      text-wrap: wrap;
      word-break: break-all;
    }
  }
}
</style>
