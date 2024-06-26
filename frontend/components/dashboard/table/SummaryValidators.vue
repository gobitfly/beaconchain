<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faArrowUpRightFromSquare,
  faPowerOff
} from '@fortawesome/pro-solid-svg-icons'
import type { DashboardValidatorContext, SummaryTimeFrame } from '~/types/dashboard/summary'
import { DashboardValidatorSubsetModal } from '#components'
import { getGroupLabel } from '~/utils/dashboard/group'
import type { DashboardKey } from '~/types/dashboard'
import type { VDBSummaryValidators } from '~/types/api/validator_dashboard'

interface Props {
  // TODO: replace with v2.5 structs
  validators: VDBSummaryValidators,
  absolute: boolean,
  groupId?: number,
  timeFrame?: SummaryTimeFrame
  context: DashboardValidatorContext,
  dashboardKey?: DashboardKey,
  isTooltip?: boolean
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
      groupId: props.groupId,
      dashboardKey: props.dashboardKey
    }
  })
}

const groupName = computed(() => {
  return getGroupLabel($t, props.groupId, groups.value)
})

const mapped = computed(() => {
  const list: { count: number, key: string }[] = []
  const addCount = (key: string, count?: number) => {
    if (count) {
      list.push({ count, key })
    }
  }

  addCount('online', props.validators?.online)
  if (props.absolute || props.isTooltip || !props.validators?.online) {
    addCount('offline', props.validators?.offline)
    addCount('exited', props.validators?.exited)
  }
  const total = props.validators?.offline + props.validators?.online

  return {
    list,
    total
  }
})

</script>
<template>
  <BcTooltip v-if="mapped.list.length" class="validator-status-column">
    <template v-if="!isTooltip" #tooltip>
      <DashboardTableSummaryValidators v-bind="props" :absolute="!props.absolute" :is-tooltip="true" />
    </template>
    <div class="status-list">
      <div v-for="status in mapped.list" :key="status.key" class="status" :class="status.key">
        <div class="icon">
          <FontAwesomeIcon :icon="faPowerOff" />
        </div>
        <BcFormatNumber v-if="absolute" :value="status.count" />
        <BcFormatPercent v-else :value="status.count" :base="mapped.total" />
      </div>
    </div>
    <FontAwesomeIcon v-if="!isTooltip" class="link popout" :icon="faArrowUpRightFromSquare" @click="openValidatorModal" />
  </BcTooltip>
  <div v-else>
    -
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/utils.scss';

.validator-status-column {
  display: flex;
  justify-content: space-between;
  align-items: center;

  .status-list {
    display: flex;
    align-items: center;
    gap: var(--padding-small);
    flex-wrap: wrap;

    .status {
      display: flex;
      align-items: center;
      gap: 3px;

      .icon {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 14px;
        height: 14px;
        border-radius: 50%;
        background-color: var(--text-color-disabled);

        svg {
          height: 8px;
          width: 8px;
        }
      }

      &.online {
        .icon {
          background-color: var(--positive-color);
          color: var(--positive-contrast-color);
        }

        span {
          color: var(--positive-color);
        }
      }

      &.offline {
        .icon {
          background-color: var(--negative-color);
          color: var(--negative-contrast-color);
        }

        span {

          color: var(--negative-color);
        }
      }
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
