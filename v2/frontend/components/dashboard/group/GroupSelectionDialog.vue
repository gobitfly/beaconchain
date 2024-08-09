<script setup lang="ts">
import {
  onMounted, ref,
} from 'vue'

interface Props {
  groupId?: number
  selectedValidators?: number
  totalValidators?: number
}
const {
  dialogRef, props, setHeader,
} = useBcDialog<Props>()
const { t: $t } = useTranslation()

const selectedGroupId = ref<number>()

onMounted(() => {
  setHeader($t('dashboard.group.selection.dialog.title'))
})

const closeDialog = (groupId?: number) => {
  dialogRef?.value.close(groupId)
}

watch(
  () => props.value?.groupId,
  (groupId) => {
    if (groupId !== undefined) {
      selectedGroupId.value = groupId
    }
  },
)
</script>

<template>
  <div class="content">
    <div class="form">
      {{ $t("dashboard.group.selection.dialog.assign_group") }}
      <DashboardGroupSelection
        v-model="selectedGroupId"
        class="group-selection"
      />
    </div>
    <div class="footer">
      <b v-if="props?.totalValidators">
        {{
          $t(
            "dashboard.group.selection.dialog.validators_selected",
            {
              total: props.totalValidators,
            },
            props.selectedValidators ?? 0,
          )
        }}</b>
      <Button
        :disabled="selectedGroupId === undefined"
        type="button"
        :label="$t('navigation.save')"
        @click="closeDialog(selectedGroupId)"
      />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.content {
  display: flex;
  flex-direction: column;
  width: 350px;

  .form {
    flex-grow: 1;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: var(--padding);
    margin: 25px 0;

    .group-selection {
      width: 195px;
    }
  }

  .footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: var(--padding);
  }
}
</style>
