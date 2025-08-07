<script setup lang="ts">
import type {
  DashboardValidatorContext,
  SummaryTimeFrame,
} from '~/types/dashboard/summary'
import { DashboardValidatorSubsetModal } from '#components'
import { getGroupLabel } from '~/utils/dashboard/group'
import type { DashboardKey } from '~/types/dashboard'
import type {
  VDBGroupSummaryData,
  VDBSummaryTableRow,
} from '~/types/api/validator_dashboard'

interface Props {
  context: DashboardValidatorContext,
  dashboardKey?: DashboardKey,
  data?: VDBGroupSummaryData,
  groupId?: number,
  row: VDBSummaryTableRow,
  timeFrame?: SummaryTimeFrame,
  validatorCount: number,
  validators: number[],
}
const props = defineProps<Props>()

const { t: $t } = useTranslation()
const { groups } = useValidatorDashboardGroups()

const dialog = useDialog()

const openValidatorModal = () => {
  dialog.open(DashboardValidatorSubsetModal, {
    data: {
      context: props.context,
      dashboardKey: props.dashboardKey,
      groupId: props.groupId,
      groupName: groupName.value,
      summary: {
        data: props.data,
        row: props.row,
      },
      timeFrame: props.timeFrame,
      validators: props.validators,
    },
  })
}

const groupName = computed(() => {
  return getGroupLabel($t, props.groupId, groups.value, $t('common.total'))
})
const v1Domain = useV1Domain()
</script>

<template>
  <div class="validator_column">
    <span
      v-if="validators.length && validatorCount <= 3"
      class="validators"
    >
      <template
        v-for="validator in validators"
        :key="validator"
      >
        <BcLink
          :to="`${v1Domain}/validator/${validator}`"
          target="_blank"
          class="link validator_link"
        >
          {{ validator }}
        </BcLink>
        <span>, </span>
      </template>
    </span>
    <span v-else>
      {{ validatorCount }} {{ $t('common.validator', validatorCount) }}
    </span>
    <BcButtonIcon
      v-if="validators?.length"
      screenreader-text="common.action.open_details"
      class="link popout"
      name="arrow-upright-from-square"
      @click="openValidatorModal"
    />
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/utils.scss";

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
    flex-shrink: 0;
  }
}
</style>
