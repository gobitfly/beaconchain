<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faArrowUpRightFromSquare } from '@fortawesome/pro-solid-svg-icons'
import type {
  DashboardValidatorContext,
  SummaryTimeFrame,
} from '~/types/dashboard/summary'
import { DashboardValidatorSubsetModal } from '#components'
import { getGroupLabel } from '~/utils/dashboard/group'
import type { DashboardKey } from '~/types/dashboard'
import type { VDBSummaryTableRow } from '~/types/api/validator_dashboard'
import type {
  SummaryValidatorsIconRowInfo,
  ValidatorSummaryIconRowKey,
} from '~/types/validator'

interface Props {
  row: VDBSummaryTableRow
  absolute: boolean
  groupId?: number
  timeFrame?: SummaryTimeFrame
  context: DashboardValidatorContext
  dashboardKey?: DashboardKey
  isTooltip?: boolean
}
const props = defineProps<Props>()

const { t: $t } = useTranslation()
const { groups } = useValidatorDashboardGroups()

const dialog = useDialog()

const openValidatorModal = () => {
  dialog.open(DashboardValidatorSubsetModal, {
    data: {
      context: props.context,
      timeFrame: props.timeFrame,
      groupName: groupName.value,
      groupId: props.groupId,
      dashboardKey: props.dashboardKey,
      summary: {
        row: props.row,
      },
    },
  })
}

const groupName = computed(() => {
  return getGroupLabel($t, props.groupId, groups.value, $t('common.total'))
})

const mapped = computed(() => {
  const list: SummaryValidatorsIconRowInfo[] = []
  const validatorIcons: SummaryValidatorsIconRowInfo[] = []
  const addCount = (key: ValidatorSummaryIconRowKey, count?: number) => {
    if (count) {
      list.push({ count, key })
    }
  }

  addCount('online', props.row?.validators.online)
  if (props.absolute || props.isTooltip || !props.row?.validators.online) {
    addCount('offline', props.row?.validators.offline)
    addCount('exited', props.row?.validators.exited)
  }
  // for the total percentage we ignore the exited validators
  const total = props.row?.validators.offline + props.row?.validators.online

  return {
    list,
    total,
    validatorIcons,
  }
})
</script>

<template>
  <div
    v-if="mapped.list.length"
    class="validator-status-column"
  >
    <BcTooltip class="status-list">
      <template
        v-if="!isTooltip"
        #tooltip
      >
        <DashboardTableSummaryValidators
          v-bind="props"
          :absolute="!props.absolute"
          :is-tooltip="true"
        />
      </template>
      <DashboardTableSummaryValidatorsIconRow
        :icons="mapped.list"
        :total="mapped.total"
        :absolute="absolute"
      />
    </BcTooltip>
    <FontAwesomeIcon
      v-if="!isTooltip"
      class="link popout"
      :icon="faArrowUpRightFromSquare"
      @click="openValidatorModal"
    />
  </div>
  <div v-else>
    -
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/utils.scss";

.validator-status-column {
  display: flex;
  align-items: center;
  flex-wrap: nowrap;
  gap: var(--padding);

  @media (max-width: 729px) {
    justify-content: space-between;
    padding-right: 13px;
  }

  .status-list {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: var(--padding-small);
  }

  .popout {
    width: 14px;
    height: auto;
    margin-left: var(--padding-small);
    flex-shrink: 0;
  }
}
</style>
