<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faInfoCircle
} from '@fortawesome/pro-regular-svg-icons'
import {
  faArrowUpRightFromSquare
} from '@fortawesome/pro-solid-svg-icons'
import { type OverviewTableData } from '~/types/dashboard/overview'
import { DashboardValidatorSubsetModal } from '#components'
interface Props {
  data: OverviewTableData
}
const props = defineProps<Props>()
const dialog = useDialog()

const { dashboardKey } = useDashboardKey()
const { getDashboardLabel } = useUserDashboardStore()

const openValidatorModal = () => {
  dialog.open(DashboardValidatorSubsetModal, {
    data: {
      context: 'dashboard',
      dashboardName: getDashboardLabel(dashboardKey.value, 'validator'),
      dashboardKey: dashboardKey.value
    }
  })
}

</script>
<template>
  <div class="box">
    <div class="main">
      <div class="big_text_label">
        {{ props.data.label }}
      </div>
      <div class="big_text">
        <BcTooltip :text="props.data.value?.fullLabel" :fit-content="true">
          <!-- eslint-disable-next-line vue/no-v-html -->
          <span v-html="props.data.value?.label" />
        </BcTooltip>
        <FontAwesomeIcon
          v-if="data?.addValidatorModal"
          class="link popout"
          :icon="faArrowUpRightFromSquare"
          @click="openValidatorModal"
        />
      </div>
    </div>
    <div v-for="(infos, index) in props.data.additonalValues" :key="index" class="additional">
      <div v-for="(addValue, subIndex) in infos" :key="subIndex" class="small_text">
        <BcTooltip :text="addValue.fullLabel" :fit-content="true">
          {{ addValue.label }}
        </BcTooltip>
      </div>
    </div>
    <div v-if="props.data.infos" class="info">
      <BcTooltip :fit-content="true">
        <FontAwesomeIcon :icon="faInfoCircle" />
        <template #tooltip>
          <div class="info-label-list">
            <div v-for="info in props.data.infos" :key="info.label">
              <div><b>{{ info.label }}:</b> {{ info.value }}</div>
            </div>
          </div>
        </template>
      </BcTooltip>
    </div>
  </div>
</template>
<style lang="scss" scoped>
.box {
  display: flex;
  align-items: center;

  .popout {
    width: 14px;
    margin-left: var(--padding);
    flex-shrink: 0;
  }

  .main,
  .additional {
    display: flex;
    flex-direction: column;

    div {
      display: inline-block;
      white-space: nowrap;
      text-wrap: nowrap;
    }
  }

  .additional {
    margin-left: 8px;

    &:nth-child(2) {
      margin-left: var(--padding);
    }
  }

}

.info-label-list {
  text-align: left;
}

.info {
  margin-left: var(--padding);
}
</style>
