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
const { props, setHeader } = useBcDialog<Props>()

const visible = defineModel<boolean>()
const shownValidators = ref<number[]>([])

watch(props, (p) => {
  if (p) {
    shownValidators.value = p.validators
    setHeader(
      p?.groupName
        ? $t('dashboard.validator.col.group') + ` "${p.groupName}"`
        : $t('dashboard.title') + (p.dashboardName ? ` "${p.dashboardName}"` : '')
    )
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
    case 'last_31d':
      return text + ' ' + $t('statistics.last_31d')
    case 'all_time':
      return text + ' ' + $t('statistics.all')
  }
  return text
})

const handleEvent = (filter: string) => {
  if (filter === '') {
    shownValidators.value = props.value?.validators ?? []
    return
  }

  shownValidators.value = []

  const index = parseInt(filter)
  if (props.value?.validators?.includes(index)) {
    shownValidators.value = [index]
  }
}

watch(visible, (value) => {
  if (!value) {
    shownValidators.value = props.value?.validators ?? []
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
  <div class="validator_subset_modal_container">
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
}
</style>
