<script lang="ts" setup>
import type { DashboardType } from '~/types/dashboard'
import { API_PATH } from '~/types/customFetch'
import type { ValidatorDashboard } from '~/types/api/dashboard'

const { t: $t } = useI18n()
const { fetch } = useCustomFetch()

const name = defineModel<string>('name', { required: true })
const isLoading = ref(false)

interface Props {
  dashboard: ValidatorDashboard,
  dashboardType: DashboardType
}
const { props, setHeader, dialogRef } = useBcDialog<Props>()

watch(props, (p) => {
  setHeader($t('dashboard.rename.title'))
  if (p) {
    name.value = p.dashboard.name
  }
}, { immediate: true })

const rename = async () => {
  isLoading.value = true
  const path = props.value?.dashboardType === 'validator' ? API_PATH.DASHBOARD_EDIT_VALIDATOR : API_PATH.DASHBOARD_EDIT_ACCOUNT
  // TODO: validate params once backend is done.
  await fetch(path, { body: { id: props.value?.dashboard.id, name: name.value } }, { dashboardKey: `${props.value?.dashboard.id}` })

  isLoading.value = false
  dialogRef?.value.close(true)
}

</script>

<template>
  <div class="dashboard_rename_modal_container">
    <InputText v-model="name" :placeholder="$t('dashboard.creation.type.placeholder')" class="input-field" />
    <div class="footer">
      <Button :disabled="!name?.length || isLoading" @click="rename">
        {{ $t('navigation.save') }}
      </Button>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.dashboard_rename_modal_container {
  width: 620px;
  display: flex;
  flex-direction: column;
  gap: var(--padding);
  margin-top: var(--padding);

  @media screen and (max-width: 640px) {
    width: unset;
  }

  .input-field{
    width: 100%;
  }

  .footer {
    display: flex;
    justify-content: flex-end;
  }
}
</style>
