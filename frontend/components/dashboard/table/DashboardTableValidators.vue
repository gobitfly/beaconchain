<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faArrowUpRightFromSquare
} from '@fortawesome/pro-solid-svg-icons'
import type { DashboardValidatorContext, SummaryDetail } from '~/types/dashboard/summary'

interface Props {
  validators: number[],
  groupId?: number,
  timeFrame?: SummaryDetail
  context: DashboardValidatorContext
}
const props = defineProps<Props>()

const modalVisibility = ref(false)

const { t: $t } = useI18n()
const { overview } = storeToRefs(useValidatorDashboardOverviewStore())

const openValidatorModal = () => {
  modalVisibility.value = true
}

const groupName = computed(() => {
  if (props.groupId === undefined) {
    return
  }
  if (props.groupId < 0) {
    return $t('dashboard.validator.summary.total_group_name')
  }
  const group = overview.value?.groups?.find(g => g.id === props.groupId)
  return group?.name || `${props.groupId}`
})

</script>
<template>
  <div class="validator_column">
    <div class="validators">
      <template v-for="v in props.validators" :key="v">
        <NuxtLink :to="`/validator/${v}`" target="_blank" class="link validator_link" :no-prefetch="true">
          {{ v }}
        </NuxtLink>
        <span>, </span>
      </template>
    </div>
    <FontAwesomeIcon
      v-if="validators?.length"
      class="link popout"
      :icon="faArrowUpRightFromSquare"
      @click="openValidatorModal"
    />
    <DashboardValidatorSubsetModal
      v-model="modalVisibility"
      :context="props.context"
      :time-frame="props.timeFrame"
      :group-name="groupName"
      :validators="props.validators"
    />
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';

.validator_column {
  display: flex;
  justify-content: space-between;
  align-items: center;

  .validators {
    @include main.truncate-text;

    span:last-child {
      display: none;
    }
  }

  .popout {
    width: 14px;
    height: auto;
    margin-left: var(--padding-small);
    flex-shrink: 0;
  }
}
</style>
