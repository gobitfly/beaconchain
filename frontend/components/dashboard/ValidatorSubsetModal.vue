<script lang="ts" setup>
import { warn } from 'vue'
import {
  faCopy
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import type { DashboardValidatorContext, SummaryTimeFrame } from '~/types/dashboard/summary'
import type { DashboardKey } from '~/types/dashboard'
import { sortValidatorIds } from '~/utils/dashboard/validator'
import { API_PATH } from '~/types/customFetch'
import { type InternalGetValidatorDashboardValidatorIndicesResponse } from '~/types/api/validator_dashboard'

const { t: $t } = useI18n()
const { fetch } = useCustomFetch()

interface Props {
  context: DashboardValidatorContext;
  timeFrame?: SummaryTimeFrame;
  dashboardName?: string,
  dashboardKey?: DashboardKey,
  groupName?: string, // overruled by dashboardName
  groupId?: number,
  validators: number[],
}
const { props, setHeader } = useBcDialog<Props>(undefined)

const visible = defineModel<boolean>()
const isLoading = ref(false)
const shownValidators = ref<number[]>([])
const validators = ref<number[]>([])
const MAX_VALIDATORS = 1000

watch(props, async (p) => {
  if (p) {
    shownValidators.value = sortValidatorIds(p.validators)
    validators.value = p.validators
    setHeader(
      p?.groupName
        ? $t('dashboard.validator.col.group') + ` "${p.groupName}"`
        : $t('dashboard.title') + (p.dashboardName ? ` "${p.dashboardName}"` : '')
    )
    // we get a maximum of 10 validators in the table, so if it's 10 we try to get more
    if (p.validators.length === 10) {
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

      const res = await fetch<InternalGetValidatorDashboardValidatorIndicesResponse>(API_PATH.DASHBOARD_VALIDATOR_INDICES, { query: { period: p?.timeFrame, duty, group_id: p?.groupId } }, { dashboardKey: `${p?.dashboardKey}` })
      validators.value = sortValidatorIds(res.data)
      shownValidators.value = validators.value
      isLoading.value = false
    }
  }
}, { immediate: true })

const caption = computed(() => {
  let text = 'Validators'
  switch (props.value?.context) {
    case 'attestation':
      text = $t('dashboard.validator.summary.row.attestations')
      break
    case 'sync':
      text = $t('dashboard.validator.summary.row.sync')
      break
    case 'slashings':
      text = $t('dashboard.validator.summary.row.slashed')
      break
    case 'proposal':
      text = $t('dashboard.validator.summary.row.proposals')
      break
    case 'group':
      text = $t('dashboard.validator.col.validators')
      break
  }

  switch (props.value?.timeFrame) {
    case 'last_24h':
      return text + ' ' + $t('statistics.last_24h')
    case 'last_7d':
      return text + ' ' + $t('statistics.last_7d')
    case 'last_30d':
      return text + ' ' + $t('statistics.last_30d')
    case 'all_time':
      return text + ' ' + $t('statistics.all')
  }
  return text
})

const handleEvent = (filter: string) => {
  if (filter === '') {
    shownValidators.value = validators.value
    return
  }

  shownValidators.value = []

  const index = parseInt(filter)
  if (!isNaN(index) && validators.value?.includes(index)) {
    shownValidators.value = [index]
  }
}

watch(visible, (value) => {
  if (!value) {
    shownValidators.value = validators.value
  }
})

function copyValidatorsToClipboard (): void {
  if (validators.value?.length === 0) {
    return
  }
  navigator.clipboard.writeText(validators.value.join(','))
    .catch((error) => {
      warn('Error copying text to clipboard:', error)
    })
}

const cappedValidators = computed(() => {
  const list = shownValidators.value.length <= MAX_VALIDATORS ? shownValidators.value : shownValidators.value.slice(0, MAX_VALIDATORS)

  return {
    count: shownValidators.value.length - list.length,
    list
  }
})

</script>

<template>
  <div class="validator_subset_modal_container">
    <div class="top_line_container">
      <span class="subtitle_text">
        {{ caption }}
      </span>
      <BcContentFilter class="content_filter" :search-placeholder="$t('common.index')" @filter-changed="handleEvent" />
    </div>
    <div class="link_container" :class="{'has_more': !!cappedValidators.count}">
      <template v-for="v in cappedValidators.list" :key="v">
        <BcLink :to="`/validator/${v}`" target="_blank" class="link">
          {{ v }}
        </BcLink>
        <span>, </span>
      </template>
      <template v-if="cappedValidators.count">
        <span>... {{ $t('common.and_more', {count: trim(cappedValidators.count, 0, 0)}) }}</span>
      </template>
    </div>
    <BcLoadingSpinner :loading="isLoading" alignment="center" class="spinner" />
    <Button class="p-button-icon-only copy_button" @click="copyValidatorsToClipboard">
      <FontAwesomeIcon :icon="faCopy" />
    </Button>
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

  .link_container {
    position: relative;
    flex-grow: 1;
    background-color: var(--subcontainer-background);
    padding: var(--padding) var(--padding) 7px var(--padding);
    border: 1px solid var(--container-border-color);
    border-radius: var(--border-radius);
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
