<script lang="ts" setup>
import { warn } from 'vue'
import {
  faCopy
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import type { DashboardValidatorContext, SummaryDetail } from '~/types/dashboard/summary'

const { t: $t } = useI18n()

interface Props {
  context: DashboardValidatorContext;
  timeFrame?: SummaryDetail;
  dashboardName?: string,
  groupName?: string, // overruled by dashboardName
  validators: number[],
}
const props = defineProps<Props>()

const visible = defineModel<boolean>()
const shownValidators = ref<number[]>(props.validators)

const header = computed(() => {
  if (props.groupName) {
    return $t('dashboard.validator.col.group') + ` "${props.groupName}"`
  }

  return $t('dashboard.title') + (props.dashboardName ? ` "${props.dashboardName}"` : '')
})

const caption = computed(() => {
  let text = 'Validators'
  switch (props.context) {
    case 'attestation':
      text = $t('dashboard.validator.summary.row.attestations')
      break
    case 'sync':
      text = $t('dashboard.validator.summary.row.sync')
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

  switch (props.timeFrame) {
    case 'details_day':
      return text + ' ' + $t('statistics.day')
    case 'details_week':
      return text + ' ' + $t('statistics.week')
    case 'details_month':
      return text + ' ' + $t('statistics.month')
    case 'details_total':
      return text + ' ' + $t('statistics.all')
  }
  return text
})

const handleEvent = (filter: string) => {
  if (filter === '') {
    shownValidators.value = props.validators
    return
  }

  shownValidators.value = []

  const index = parseInt(filter)
  if (props.validators.includes(index)) {
    shownValidators.value = [index]
  }
}

watch(visible, (value) => {
  if (!value) {
    shownValidators.value = props.validators
  }
})

function copyValidatorsToClipboard (): void {
  if (shownValidators.value.length === 0) {
    return
  }

  let text = ''
  shownValidators.value.forEach((v, i) => {
    text += v
    if (i !== shownValidators.value.length - 1) {
      text += ','
    }
  })
  navigator.clipboard.writeText(text)
    .catch((error) => {
      warn('Error copying text to clipboard:', error)
    })
}
</script>

<template>
  <BcDialog v-model="visible" :header="header" class="validator_subset_modal_container">
    <div class="top_line_container">
      <span class="subtitle_text">
        {{ caption }}
      </span>
      <BcContentFilter class="content_filter" :search-placeholder="$t('common.index')" @filter-changed="handleEvent" />
    </div>
    <div class="link_container">
      <template v-for="v in shownValidators" :key="v">
        <NuxtLink :to="`/validator/${v}`" target="_blank" class="link" :no-prefetch="true">
          {{ v }}
        </NuxtLink>
        <span>, </span>
      </template>
    </div>
    <Button class="p-button-icon-only copy_button" @click="copyValidatorsToClipboard">
      <FontAwesomeIcon :icon="faCopy" />
    </Button>
  </BcDialog>
</template>

<style lang="scss" scoped>
 :global(.validator_subset_modal_container) {
    width: 450px;
    height: 569px;
  }

  :global(.validator_subset_modal_container .p-dialog-content) {
      display: flex;
      flex-direction: column;
      flex-grow: 1;
  }

  :global(.validator_subset_modal_container .p-dialog-content .copy_button) {
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

  .link_container {
    position: relative;
    flex-grow: 1;
    background-color: var(--subcontainer-background);
    padding: var(--padding) var(--padding) 7px var(--padding);
    border: 1px solid var(--container-border-color);
    border-radius: var(--border-radius);
    height: 453px;
    overflow-y: auto;
    word-break: break-all;

    span:last-child {
      display: none;
    }
  }
</style>
