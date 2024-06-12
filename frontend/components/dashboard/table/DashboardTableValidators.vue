<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faArrowUpRightFromSquare
} from '@fortawesome/pro-solid-svg-icons'
import type { DashboardValidatorContext } from '~/types/dashboard/summary'
import { DashboardValidatorSubsetModal } from '#components'
import type { TimeFrame } from '~/types/value'
import { getGroupLabel } from '~/utils/dashboard/group'
import { sortValidatorIds } from '~/utils/dashboard/validator'
import type { DashboardKey } from '~/types/dashboard'

interface Props {
  validators: number[],
  groupId?: number,
  timeFrame?: TimeFrame
  context: DashboardValidatorContext,
  dashboardKey?: DashboardKey,
}
const props = defineProps<Props>()

const { t: $t } = useI18n()
const { groups } = useValidatorDashboardGroups()

const dialog = useDialog()

const openValidatorModal = () => {
  dialog.open(DashboardValidatorSubsetModal, {
    data: {
      context: props.context,
      timeFrame: props.timeFrame,
      groupName: groupName.value,
      validators: props.validators,
      groupId: props.groupId,
      dashboardKey: props.dashboardKey
    }
  })
}

const groupName = computed(() => {
  return getGroupLabel($t, props.groupId, groups.value)
})

const cappedValidators = computed(() => sortValidatorIds(props.validators).slice(0, 10) || [])

</script>
<template>
  <div class="validator_column">
    <div class="validators">
      <template v-for="v in cappedValidators" :key="v">
        <BcLink :to="`/validator/${v}`" target="_blank" class="link validator_link">
          {{ v }}
        </BcLink>
        <span>, </span>
      </template>
    </div>
    <FontAwesomeIcon
      v-if="validators?.length"
      class="link popout"
      :icon="faArrowUpRightFromSquare"
      @click="openValidatorModal"
    />
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/utils.scss';

.validator_column {
  display: flex;
  justify-content: space-between;
  align-items: center;

  .validators {
    @include utils.truncate-text;

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
