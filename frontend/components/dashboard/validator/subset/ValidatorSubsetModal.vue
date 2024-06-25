<script lang="ts" setup>
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faCaretRight
} from '@fortawesome/pro-solid-svg-icons'
import { uniqBy } from 'lodash-es'
import type { DashboardValidatorContext, SummaryTimeFrame } from '~/types/dashboard/summary'
import type { DashboardKey } from '~/types/dashboard'
import type { ValidatorSubset } from '~/types/validator'
import { sortSummaryValidators } from '~/utils/dashboard/validator'
import { API_PATH } from '~/types/customFetch'
import { type InternalGetValidatorDashboardSummaryValidatorsResponse, type VDBSummaryValidator, type VDBSummaryValidatorsData } from '~/types/api/validator_dashboard'

const { t: $t } = useI18n()
const { fetch } = useCustomFetch()

interface Props {
  context: DashboardValidatorContext;
  timeFrame?: SummaryTimeFrame;
  dashboardName?: string,
  dashboardKey?: DashboardKey,
  groupName?: string, // overruled by dashboardName
  groupId?: number,
}
const { props, setHeader } = useBcDialog<Props>(undefined)

const isLoading = ref(false)
const filter = ref('')
const data = ref<VDBSummaryValidatorsData[]>([])

watch(props, async (p) => {
  if (p) {
    let text = 'Validators'
    switch (p.context) {
      case 'attestation':
        text = $t('dashboard.validator.summary.row.attestations')
        break
      case 'sync':
        text = $t('dashboard.validator.summary.row.sync_committee')
        break
      case 'slashings':
        text = $t('dashboard.validator.summary.row.slashings')
        break
      case 'proposal':
        text = $t('dashboard.validator.summary.row.proposals')
        break
      case 'group':
        text = $t('dashboard.validator.col.validators')
        break
    }

    setHeader(
      text
    )

    isLoading.value = true
    let duty = ''
    switch (p.context) {
      case 'sync':
        duty = 'sync'
        break
      case 'proposal':
        duty = 'proposal'
        break
      case 'slashings':
        duty = 'slashed'
        break
    }

    const res = await fetch<InternalGetValidatorDashboardSummaryValidatorsResponse>(API_PATH.DASHBOARD_VALIDATOR_INDICES, { query: { period: p?.timeFrame?.replace('last_', ''), duty, group_id: p?.groupId } }, { dashboardKey: `${p?.dashboardKey}` })
    data.value = res.data
    isLoading.value = false
  }
}, { immediate: true })

const mapped = computed<ValidatorSubset[]>(() => {
  const sortAndFilter = (validators:VDBSummaryValidator[]):VDBSummaryValidator[] => {
    if (!filter.value) {
      return sortSummaryValidators(validators)
    } else {
      const index = parseInt(filter.value)
      if (isNaN(index)) {
        return []
      }
      const vali = validators.find(v => v.index === index)
      if (vali) {
        return [vali]
      }
    }
    return []
  }

  const filtered:ValidatorSubset[] = data.value.map(sub => ({
    category: sub.category,
    validators: sortAndFilter(sub.validators)
  })).filter(s => !!s.validators.length)
  if (filtered.length && !filter.value) {
    const all:ValidatorSubset = {
      category: 'all',
      validators: []
    }
    all.validators = sortSummaryValidators(uniqBy(filtered.reduce((list, sub) => list.concat(sub.validators), all.validators), 'index'))
    filtered.splice(0, 0, all)
    return filtered
  }
  return filtered
})

</script>

<template>
  <div class="validator_subset_modal_container">
    <div class="top_line_container">
      <span class="subtitle_text">
        {{ props?.groupName }}
      </span>
      <BcContentFilter v-model="filter" class="content_filter" :search-placeholder="$t('common.index')" @filter-changed="(f:string)=>filter=f" />
    </div>

    <Accordion :active-index="0" class="accordion basic">
      <AccordionTab v-for="subset in mapped" :key="subset.category">
        <template #headericon>
          <FontAwesomeIcon :icon="faCaretRight" />
        </template>
        <template #header>
          <DashboardValidatorSubsetHeader :category="subset.category" :validators="subset.validators" />
        </template>
        <DashboardValidatorSubsetList :category="subset.category" :validators="subset.validators" />
      </AccordionTab>
      <BcLoadingSpinner :loading="isLoading" alignment="center" class="spinner" />
    </Accordion>
  </div>
</template>

<style lang="scss" scoped>
.validator_subset_modal_container {
  width: 410px;
  height: 489px;
  display: flex;
  flex-direction: column;
  flex-grow: 1;

  @media screen and (max-width: 500px) {
    width: unset;
    height: unset;
  }

  .copy_button {
    position: absolute;
    bottom: calc(var(--padding-large) + var(--padding));
    right: calc(var(--padding-large) + var(--padding));
  }

  .top_line_container {
    padding: var(--padding) 0 14px 0;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .content_filter {
    width: 169px;
  }

  .spinner {
    position: absolute;
  }

  .accordion {
    position: relative;
    flex-grow: 1;
    height: 453px;
    overflow-y: auto;
    overflow-x: hidden;
    word-break: break-all;

    &:not(.has_more) span:last-child {
      display: none;
    }
  }
}
</style>
