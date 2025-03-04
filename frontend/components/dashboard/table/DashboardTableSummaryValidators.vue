<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faArrowUpRightFromSquare } from '@fortawesome/pro-solid-svg-icons'
import type {
  DashboardValidatorContext,
  SummaryTimeFrame,
} from '~/types/dashboard/summary'
import { LazyDashboardValidatorSubsetModal } from '#components'
import { getGroupLabel } from '~/utils/dashboard/group'
import type { DashboardKey } from '~/types/dashboard'
import type {
  VDBSummaryTableRow, VDBSummaryValidators,
} from '~/types/api/validator_dashboard'

const props = defineProps<{
  context: DashboardValidatorContext,
  dashboardKey?: DashboardKey,
  groupId?: number,
  isAbsolute: boolean,
  row: VDBSummaryTableRow,
  timeFrame?: SummaryTimeFrame,
  validators: VDBSummaryValidators,
}>()

const { t: $t } = useTranslation()
const { groups } = useValidatorDashboardGroups()

const dialog = useDialog()

const openValidatorModal = () => {
  dialog.open(LazyDashboardValidatorSubsetModal, {
    data: {
      context: props.context,
      dashboardKey: props.dashboardKey,
      groupId: props.groupId,
      groupName: groupName.value,
      summary: { row: props.row },
      timeFrame: props.timeFrame,
    },
  })
}

const groupName = computed(() => {
  return getGroupLabel($t, props.groupId, groups.value, $t('common.total'))
})

const hasValidators = computed(
  () => props.validators.online || props.validators.offline || props.validators.exited,
)
</script>

<template>
  <div
    v-if="hasValidators"
    class="validator-status-column"
  >
    <BcTooltip fit-content>
      <section class="validator-status-container">
        <template
          v-if="!isAbsolute"
        >
          <DashboardTableSummaryValidatorsStatus
            v-if="validators.online"
            color="green"
            :value="validators.online"
            :base="validators.online + validators.offline"
          />
          <DashboardTableSummaryValidatorsStatus
            v-if="validators.offline"
            color="red"
            :value="validators.offline"
            :base="validators.online + validators.offline"
          />
          <DashboardTableSummaryValidatorsStatus
            v-if="validators.exited"
            color="gray"
            :value="validators.exited"
            :base="validators.online + validators.offline + validators.exited"
          />
        </template>
        <template v-else>
          <DashboardTableSummaryValidatorsStatus
            v-if="validators.online"
            color="green"
            :value="validators.online"
          />
          <DashboardTableSummaryValidatorsStatus
            v-if="validators.offline"
            color="red"
            :value="validators.offline"
          />
          <DashboardTableSummaryValidatorsStatus
            v-if="validators.exited"
            color="gray"
            :value="validators.exited"
          />
        </template>
      </section>
      <template #tooltip>
        <div class="validator-tooltip">
          <template v-if="!isAbsolute">
            <DashboardTableSummaryValidatorsStatus
              v-if="validators.online"
              color="green"
              :value="validators.online"
            />
            <DashboardTableSummaryValidatorsStatus
              v-if="validators.offline"
              color="red"
              :value="validators.offline"
            />
            <DashboardTableSummaryValidatorsStatus
              v-if="validators.exited"
              color="gray"
              :value="validators.exited"
            />
          </template>
          <template v-else>
            <DashboardTableSummaryValidatorsStatus
              v-if="validators.online"
              color="green"
              :value="validators.online"
              :base="validators.online + validators.offline"
            />
            <DashboardTableSummaryValidatorsStatus
              v-if="validators.offline"
              color="red"
              :value="validators.offline"
              :base="validators.online + validators.offline"
            />
            <DashboardTableSummaryValidatorsStatus
              v-if="validators.exited"
              color="gray"
              :value="validators.exited"
              :base="validators.online + validators.offline + validators.exited"
            />
          </template>
        </div>
      </template>
    </BcTooltip>
    <BcButtonIcon
      :screenreader-text="$t('dashboard.validator.summary.validator_status_popout')"
    >
      <FontAwesomeIcon
        class="link popout"
        :icon="faArrowUpRightFromSquare"
        @click="openValidatorModal"
      />
    </BcButtonIcon>
  </div>
  <div v-else>
    -
  </div>
</template>

<style lang="scss" scoped>
.validator-status-column {
  display: flex;
  align-items: center;
  flex-wrap: nowrap;

  .popout {
    width: 14px;
    height: auto;
    margin-left: var(--padding-small);
    flex-shrink: 0;
  }
}
.validator-status-container {
  display: flex;
  justify-content: left;
  flex-wrap: wrap;
  gap: var(--padding-small)
}
.validator-tooltip {
  display: flex;
  flex-direction: column;
  gap: var(--padding-small)
}
</style>
