<script lang="ts" setup>
import type { DashboardType } from '~/types/dashboard'
import { API_PATH } from '~/types/customFetch'
import type { ValidatorDashboard } from '~/types/api/dashboard'

const { t: $t } = useTranslation()
const { fetch } = useCustomFetch()

const name = defineModel<string>('name', { default: '' })
const isLoading = ref(false)

interface Props {
  dashboard: ValidatorDashboard
  dashboardType: DashboardType
}
const {
  dialogRef, props, setHeader,
} = useBcDialog<Props>({ pt: { header: { class: 'dashboard-rename-modal-header' } } })

watch(props, (p) => {
  let title = $t('dashboard.rename.title')
  if (p) {
    name.value = p.dashboard.name
    title += ' - ' + p.dashboard.name
  }
  setHeader(title)
}, { immediate: true })

const renameDisabled = computed(() => {
  return (
    !name.value?.length
    || isLoading.value
    || !REGEXP_VALID_NAME.test(name.value)
  )
})

const rename = async () => {
  name.value = name.value.trim()

  if (renameDisabled.value) {
    return
  }

  isLoading.value = true
  const path
    = props.value?.dashboardType === 'validator'
      ? API_PATH.DASHBOARD_RENAME_VALIDATOR
      : API_PATH.DASHBOARD_RENAME_ACCOUNT
  await fetch(
    path,
    { body: { name: name.value } },
    { dashboardKey: `${props.value?.dashboard.id}` },
  )

  isLoading.value = false

  dialogRef?.value.close(true)
}
</script>

<template>
  <div class="dashboard_rename_modal_container">
    <InputText
      v-model="name"
      :placeholder="$t('dashboard.creation.type.placeholder')"
      :maxlength="50"
      class="input-field"
      @keypress.enter="rename"
    />
    <div class="footer">
      <Button
        :disabled="renameDisabled"
        @click="rename"
      >
        {{ $t("navigation.save") }}
      </Button>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/utils.scss";

:global(.dashboard-rename-modal-header .p-dialog-title) {
  @include utils.truncate-text;
  max-width: 600px;
}
.dashboard_rename_modal_container {
  width: 620px;
  display: flex;
  flex-direction: column;
  gap: var(--padding);
  margin-top: var(--padding);

  @media screen and (max-width: 640px) {
    width: unset;
  }

  .input-field {
    width: 100%;
  }

  .footer {
    display: flex;
    justify-content: flex-end;
  }
}
</style>
