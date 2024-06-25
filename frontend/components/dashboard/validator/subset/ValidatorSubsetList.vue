<script lang="ts" setup>
import { warn } from 'vue'
import {
  faCopy
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import type { Paging } from '~/types/api/common'
import { type Cursor } from '~/types/datatable'
import type { ValidatorSubsetCategory } from '~/types/validator'
import type { VDBSummaryValidator } from '~/types/api/validator_dashboard'

interface Props {
  category: ValidatorSubsetCategory,
  validators: VDBSummaryValidator[],
}
const props = defineProps<Props>()

const { t: $t } = useI18n()

const paging = ref<Paging | null>(null)
const cursor = ref<Cursor>(undefined)
const VALIDATORS_PER_PAGE = 100

watch(props, (p) => {
  cursor.value = undefined
  if (p?.validators?.length) {
    if (p.validators.length > VALIDATORS_PER_PAGE) {
      paging.value = {
        total_count: p.validators.length
      }
    } else {
      paging.value = null
    }
  } else {
    paging.value = null
  }
}, { immediate: true })

const currentPage = computed<VDBSummaryValidator[]>(() => {
  if (!props.validators?.length) {
    return []
  }
  const start = (cursor.value as number) || 0
  return props.validators.slice(start, start + VALIDATORS_PER_PAGE)
})

function copyValidatorsToClipboard (): void {
  if (!props.validators?.length) {
    return
  }
  navigator.clipboard.writeText(props.validators.map(v => v.index).join(','))
    .catch((error) => {
      warn('Error copying text to clipboard:', error)
    })
}

function mapDutyLabel (dutyObjects?: number[]) {
  if (!dutyObjects) {
    return
  }
  switch (props.category) {
    case 'proposal_proposed':
      return $t('common.block', dutyObjects.length) + ':'
    case 'proposal_missed':
      return $t('common.slot', dutyObjects.length) + ':'
    case 'pending':
      return formatGoTimestamp(dutyObjects[0], undefined, 'relative', 'short', $t('locales.date'), true)
    case 'has_slashed':
      return $t('dashboard.validator.subset_dialog.slashed') + ':'
    case 'got_slashed':
      return $t('dashboard.validator.subset_dialog.got_slashed') + ':'
  }
}
function mapDutyLinks (dutyObjects?: number[]) {
  if (!dutyObjects) {
    return
  }
  const links: { to: string, label: string }[] = []
  let path = ''
  switch (props.category) {
    case 'proposal_proposed':
      path = '/block/'
      break
    case 'proposal_missed':
      path = '/slot/'
      break
    case 'has_slashed':
    case 'got_slashed':
      path = '/validator/'
      break
  }
  if (path) {
    return dutyObjects.map(o => ({
      to: `${path}${o}`,
      label: `${o}`
    }))
  }
  return links
}

</script>

<template>
  <div class="validator-list">
    <div class="list">
      <template v-for="v in currentPage" :key="v.index">
        <BcLink :to="`/validator/${v.index}`" target="_blank" class="link">
          {{ v.index }}
        </BcLink>
        <template v-if="mapDutyLabel(v.duty_objects)">
          <span class="round-brackets">
            <span class="label">{{ mapDutyLabel(v.duty_objects) }}</span>
            <template
              v-for="link in mapDutyLinks(v.duty_objects)"
              :key="link.label"
            >
              <BcLink
                :to="link.to"
                target="_blank"
                class="link"
              >
                {{ link.label }}
              </BcLink>
              <span>, </span>

            </template>
          </span>
        </template>
        <span>, </span>
      </template>
    </div>
    <div v-if="paging" class="page-row">
      <BcTablePager
        :cursor="cursor"
        :page-size="VALIDATORS_PER_PAGE"
        :paging="paging"
        :stepper-only="true"
        @set-cursor="(c: Cursor) => cursor = c"
      />
    </div>
    <Button class="p-button-icon-only copy_button" @click="copyValidatorsToClipboard">
      <FontAwesomeIcon :icon="faCopy" />
    </Button>
  </div>
</template>

<style lang="scss" scoped>
.validator-list {
  position: relative;
  flex-grow: 1;
  background-color: var(--subcontainer-background);
  padding: var(--padding) var(--padding) 7px var(--padding);
  border: 1px solid var(--container-border-color);
  border-radius: var(--border-radius);
  word-break: break-all;
  .list{
    min-height: 40px;
  }

  .round-brackets >span:last-child:not(.label),
  .list >span:last-child {
    display: none;
  }

  .round-brackets {
    margin-left: 3px;
    &:has(a){
      .label{
        margin-right: 3px;
      }
    }
  }

  .page-row {
    width: 100%;
    height: 52px;
    margin-top: var(--padding);
    border-top: var(--container-border);
    display: flex;
    justify-content: center;
    align-items: center;
  }

  .copy_button {
    position: absolute;
    bottom: var(--padding);
    right: var(--padding);
  }
}
</style>
