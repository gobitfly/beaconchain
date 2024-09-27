<script lang="ts" setup>
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faCaretRight } from '@fortawesome/pro-solid-svg-icons'
import { uniqBy } from 'lodash-es'
import type {
  DashboardValidatorContext,
  SummaryTimeFrame,
} from '~/types/dashboard/summary'
import type { DashboardKey } from '~/types/dashboard'
import type {
  ValidatorSubset,
  ValidatorSubsetCategory,
} from '~/types/validator'
import { sortSummaryValidators } from '~/utils/dashboard/validator'
import { API_PATH } from '~/types/customFetch'
import type {
  GetValidatorDashboardSummaryValidatorsResponse,
  VDBGroupSummaryData,
  VDBSummaryTableRow,
  VDBSummaryValidator,
  VDBSummaryValidatorsData,
} from '~/types/api/validator_dashboard'

const { t: $t } = useTranslation()
const { fetch } = useCustomFetch()

interface Props {
  context: DashboardValidatorContext,
  dashboardKey?: DashboardKey,
  dashboardName?: string,
  groupId?: number,
  groupName?: string,
  summary?: {
    data?: VDBGroupSummaryData,
    row: VDBSummaryTableRow,
  },
  timeFrame?: SummaryTimeFrame,
}
const {
  props, setHeader,
} = useBcDialog<Props>(undefined)

const isLoading = ref(false)
const filter = ref('')
const data = ref<VDBSummaryValidatorsData[]>([])

watch(
  props,
  async (p) => {
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

      setHeader(text)

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

      const res
        = await fetch<GetValidatorDashboardSummaryValidatorsResponse>(
          API_PATH.DASHBOARD_VALIDATOR_INDICES,
          {
            query: {
              duty,
              group_id: p?.groupId,
              period: p?.timeFrame,
            },
          },
          { dashboardKey: `${p?.dashboardKey}` },
        )
      data.value = res.data
      isLoading.value = false
    }
  },
  { immediate: true },
)

const subsets = computed<ValidatorSubset[]>(() => {
  const sortAndFilter = (
    validators: VDBSummaryValidator[],
  ): VDBSummaryValidator[] => {
    if (!filter.value) {
      return sortSummaryValidators(validators)
    }
    else {
      const index = parseInt(filter.value)
      if (isNaN(index)) {
        return []
      }
      const vali = validators.find(v => v.index === index)
      if (vali) {
        return [ vali ]
      }
    }
    return []
  }

  const filtered: ValidatorSubset[] = data.value
    .map(sub => ({
      category: sub.category,
      validators: sortAndFilter(sub.validators),
    }))
    .filter(s => !!s.validators.length)

  // Let's combine what needs to be combined
  if (filtered.length > 1) {
    if (
      props.value?.context === 'group'
      || props.value?.context === 'dashboard'
    ) {
      const all: ValidatorSubset = {
        category: 'all',
        validators: [],
      }
      all.validators = sortSummaryValidators(
        uniqBy(
          filtered.reduce(
            (list, sub) =>
              list.concat(
                sub.validators.map(v => ({
                  duty_objects: [],
                  index: v.index,
                })),
              ),
            all.validators,
          ),
          'index',
        ),
      )
      filtered.splice(0, 0, all)
    }

    // we need to split up the withdrawn and withrawing categories into exited
    // and slashed and not show them individually
    const withdrawnIndex = filtered.findIndex(
      s => s.category === 'withdrawn',
    )
    const withdrawn
      = withdrawnIndex >= 0 ? filtered.splice(withdrawnIndex, 1)[0] : undefined
    const withdrawingIndex = filtered.findIndex(
      s => s.category === 'withdrawing',
    )
    const withdrawing
      = withdrawingIndex >= 0
        ? filtered.splice(withdrawingIndex, 1)[0]
        : undefined
    if (withdrawn?.validators.length || withdrawing?.validators.length) {
      // a withrawn/withrawing validator can either be in the exited or slashed group
      const categories: ValidatorSubsetCategory[] = [
        'exited',
        'slashed',
      ]
      categories.forEach((category) => {
        const index = filtered.findIndex(s => s.category === category)
        if (index >= 0) {
          const baseSubset = filtered[index]

          const xWithdrawn: ValidatorSubset = {
            category: `${category}_withdrawn` as ValidatorSubsetCategory,
            validators: [],
          }
          const xWithdrawing: ValidatorSubset = {
            category: `${category}_withdrawing` as ValidatorSubsetCategory,
            validators: [],
          }

          const subsets = [
            [
              withdrawn,
              xWithdrawn,
            ],
            [
              withdrawing,
              xWithdrawing,
            ],
          ]
          baseSubset.validators.forEach((v) => {
            subsets.forEach(([
              origin,
              merged,
            ]) => {
              if (origin?.validators.find(sV => v.index === sV.index)) {
                merged?.validators.push({
                  ...v,
                  duty_objects: [],
                })
              }
            })
          })

          subsets.forEach(([
            _origin,
            merged,
          ]) => {
            if (merged?.validators.length) {
              filtered.splice(index + 1, 0, merged)
            }
          })
        }
      })
    }
  }
  return filtered
})
</script>

<template>
  <div class="validator_subset_modal_container">
    <div class="top_line_container">
      <DashboardValidatorSubsetSubHeader
        v-if="props"
        :context="props.context"
        :sub-title="props.groupName || props.dashboardName"
        :summary="props.summary"
        :subsets
      />
      <BcContentFilter
        v-model="filter"
        :search-placeholder="$t('common.index')"
        @filter-changed="(f: string) => (filter = f)"
      />
    </div>

    <Accordion
      :active-index="-1"
      class="accordion basic"
    >
      <AccordionTab
        v-for="subset in subsets"
        :key="subset.category"
      >
        <template #headericon>
          <FontAwesomeIcon :icon="faCaretRight" />
        </template>
        <template #header>
          <DashboardValidatorSubsetListHeader
            :category="subset.category"
            :validators="subset.validators"
          />
        </template>
        <DashboardValidatorSubsetList
          :category="subset.category"
          :validators="subset.validators"
        />
      </AccordionTab>
    </Accordion>
    <BcLoadingSpinner
      :loading="isLoading"
      alignment="center"
      class="spinner"
    />
  </div>
</template>

<style lang="scss" scoped>
.validator_subset_modal_container {
  width: 410px;
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
    gap: var(--padding);
    overflow: hidden;
  }

  .spinner {
    position: absolute;
  }

  .accordion {
    position: relative;
    flex-grow: 1;
    max-height: 453px;
    min-height: 453px;
    overflow-y: auto;
    overflow-x: hidden;
    word-break: break-all;

    &:not(.has_more) span:last-child {
      display: none;
    }
  }
}
</style>
