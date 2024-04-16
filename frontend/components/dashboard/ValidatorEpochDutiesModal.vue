<script lang="ts" setup>
import type { DataTableSortEvent } from 'primevue/datatable'
import type { DashboardKey } from '~/types/dashboard'
import type { Cursor } from '~/types/datatable'
import type { InternalGetValidatorDashboardDutiesResponse } from '~/types/api/validator_dashboard'
import type { ValidatorHistoryDuties } from '~/types/api/common'

const { t: $t } = useI18n()
const { fetch } = useCustomFetch()

const { width } = useWindowSize()
const size = computed(() => {
  return {
    expandable: width.value <= 1000
  }
})

interface Props {
  dashboardKey: DashboardKey // TODO: once the dashboardKey PR is finished replace this props!
  groupId: number
  groupName?: string
  epoch: number
}

const { props, setHeader } = useBcDialog<Props>({ showHeader: size.value.expandable, contentClass: 'epoch-duties-modal' })

const isLoading = ref(false)
const cursor = ref<Cursor>()
const pageSize = ref<number>(5)

const { value: query, bounce: setQuery } = useDebounceValue<PathValues | undefined>({ limit: pageSize.value }, 500)

const data = ref<InternalGetValidatorDashboardDutiesResponse | undefined>()

const onSort = (sort: DataTableSortEvent) => {
  setQuery(setQuerySort(sort, query?.value))
}

const setCursor = (value: Cursor) => {
  cursor.value = value
  setQuery(setQueryCursor(value, query?.value))
}

const setPageSize = (value: number) => {
  pageSize.value = value
  setQuery(setQueryPageSize(value, query?.value))
}

const setSearch = (value?: string) => {
  setQuery(setQuerySearch(value, query?.value))
}

const loadData = async () => {
  if (props.value?.dashboardKey) {
    isLoading.value = !data.value
    const testQ = JSON.stringify(query.value)
    const result = await fetch<InternalGetValidatorDashboardDutiesResponse>(API_PATH.DASHBOARD_VALIDATOR_EPOCH_DUTY, { query: { ...query.value, group_id: props.value.groupId } }, { dashboardKey: props.value.dashboardKey, epoch: props.value.epoch }, query.value)

    // Make sure that during loading the query did not change
    if (testQ === JSON.stringify(query.value)) {
      data.value = result
    }
    isLoading.value = false
  }
}

watch(() => [props.value, query.value], () => {
  loadData()
}, { immediate: true })

const mapDuties = (duties: ValidatorHistoryDuties) => {
  const list = []
  if (duties.attestation_head || duties.attestation_source || duties.attestation_target) {
    list.push($t('dashboard.validator.rewards.attestation'))
  }
  if (duties.proposal) {
    list.push($t('dashboard.validator.rewards.proposal'))
  }
  if (duties.sync) {
    list.push($t('dashboard.validator.rewards.sync_committee'))
  }
  if (duties.slashing) {
    list.push($t('dashboard.validator.rewards.slashing'))
  }
  return list.join(', ')
}

const title = computed(() => {
  let t = $t('dashboard.validator.duties.title')
  if (props.value?.groupName && !size.value.expandable) {
    t += ` (${props.value.groupName})`
  }
  return t
})

watch([title, size], () => {
  setHeader(title.value, size.value.expandable)
}, { immediate: true })

</script>

<template>
  <BcTableControl :search-placeholder="$t('dashboard.validator.duties.search_placeholder')" @set-search="setSearch">
    <template v-if="size.expandable" #header-left>
      <div class="small-title">
        {{ props?.groupName }}
      </div>
    </template>
    <template v-else #header-center>
      <div>
        <span class="h1">{{ title }}</span>
        <BcFormatTimePassed :value="props?.epoch" class="time-passed" />
      </div>
    </template>
    <template #table>
      <ClientOnly fallback-tag="span">
        <BcTable
          :data="data"
          data-key="validator"
          :expandable="size.expandable"
          class="duties-table"
          :cursor="cursor"
          :loading="isLoading"
          :page-size="pageSize"
          @set-cursor="setCursor"
          @sort="onSort"
          @set-page-size="setPageSize"
        >
          <Column field="validator" :sortable="true" :header="$t('dashboard.validator.duties.col.validator')">
            <template #body="slotProps">
              <NuxtLink
                :to="`/validator/${slotProps.data.validator}`"
                target="_blank"
                class="link validator_link"
                :no-prefetch="true"
              >
                <BcFormatNumber :value="slotProps.data.validator" />
              </NuxtLink>
            </template>
          </Column>
          <Column field="duties" :header="$t('dashboard.validator.duties.col.duties')">
            <template #body="slotProps">
              <div class="col-duties">
                {{ mapDuties(slotProps.data.duties) }}
              </div>
            </template>
          </Column>
          <Column v-if="!size.expandable" field="result" :header="$t('dashboard.validator.duties.col.result')">
            <template #body="slotProps">
              <ValidatorTableDutyStatus :data="slotProps.data.duties" />
            </template>
          </Column>
          <Column field="rewards" :sortable="!size.expandable" :header="$t('dashboard.validator.duties.col.rewards')">
            <template #body="slotProps">
              <ValidatorTableDutyRewards :data="slotProps.data.duties" />
            </template>
          </Column>
          <template #expansion="slotProps">
            <div class="expansion">
              <div class="info">
                <div class="label">
                  {{ $t('dashboard.validator.duties.col.result') }}
                </div>
                <div>
                  <ValidatorTableDutyStatus :data="slotProps.data.duties" />
                </div>
              </div>
            </div>
          </template>
        </BcTable>
      </ClientOnly>
    </template>
  </BcTableControl>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';
@use '~/assets/css/utils.scss';
@use '~/assets/css/fonts.scss';

:global(.epoch-duties-modal) {
  width: 960px;
  max-width: 100%;
  height: 643px;
  max-height: 100%;
  display: flex;
  flex-direction: column;

  @media (max-width: 1000px) {
    width: 100%;
    min-width: 100%;
  }

}

:global(.epoch-duties-modal .bc-table-header) {
  padding: 0;
}

.small-title {
  @include fonts.subtitle_text;
}

.col-duties {
  white-space: wrap;
  text-wrap: wrap;
}

.time-passed {
  color: var(--text-color-disabled);
  margin-left: var(--padding-small);
}

.duties-table {
  @include main.container;
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  overflow-y: hidden;

  :deep(.p-datatable-wrapper) {
    flex-grow: 1;
  }
}

.expansion {
  @include main.container;
  padding: var(--padding-large) var(--padding);
  display: flex;
  flex-direction: column;
  gap: var(--padding);

  .info {
    display: flex;
    align-items: center;
    gap: var(--padding);

    .label {
      font-weight: var(--standard_text_bold_font_weight);
      margin: 0 30px;
    }
  }
}
</style>
