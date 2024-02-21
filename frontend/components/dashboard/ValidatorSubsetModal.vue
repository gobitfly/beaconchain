<script lang="ts" setup>
import { warn } from 'vue'
import type { DashboardValidatorContext, SummaryDetail } from '~/types/dashboard/summary'

const { t: $t } = useI18n()

interface Props {
  context: DashboardValidatorContext;
  timeFrame: SummaryDetail;
  dashboardName: string,
  groupName?: string,
  validators: number[],
}
const props = defineProps<Props>()

const visible = defineModel<boolean>()
const shownValidators = ref<number[]>(props.validators)

const header = computed(() => {
  if (props.groupName) {
    return $t('dashboard.validator.summary.col.group') + ` "${props.groupName}"`
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
  }

  switch (props.timeFrame) {
    case 'details_24h':
      return text + ' ' + $t('statistics.24h')
    case 'details_7d':
      return text + ' ' + $t('statistics.7d')
    case 'details_31d':
      return text + ' ' + $t('statistics.31d')
    case 'details_all':
      return text + ' ' + $t('statistics.all')
  }
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
      <BcContentFilter class="content_filter" @filter-changed="handleEvent" />
    </div>
    <div class="link_container">
      <span v-for="(v) in shownValidators" :key="v" class="link_list">
        <NuxtLink :to="`/validator/${v}`" target="blank" class="link">
          {{ v }}
        </NuxtLink>
        <span>, </span>
      </span>
    </div>
    <Button class="p-button-icon-only copy_button" @click="copyValidatorsToClipboard">
      <i class="fas fa-copy" />
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

    .link_list:last-child span:last-child {
      display: none;
    }
  }
</style>
