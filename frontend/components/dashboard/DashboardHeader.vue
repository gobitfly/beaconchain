<script lang="ts" setup>
import type {
  MenuBarButton, MenuBarEntry,
} from '~/types/menuBar'

const { t: $t } = useTranslation()
const {
  totalValidatorDashboards,
  validatorDashboards,
} = usePrivateDashboards()

const emit = defineEmits<{ (e: 'showCreation'): void }>()

const {
  key,
  // navigateToDashboard,
  variant,
} = useDashboard()

const items = computed<MenuBarButton[]>(() => {
  if (totalValidatorDashboards.value) {
    return validatorDashboards.value.map((dashboard) => {
      return {
        active: key.value === `${dashboard.id}`,
        command: () => navigateTo({
          name: 'dashboard-id',
          params: {
            id: dashboard.id,
          },
        }),
        label: dashboard.name,
      }
    })
  }
  return []
})

const buttons: MenuBarEntry[] = [ {
  dropdown: true,
  items: items.value,
  label: `${$t('dashboard.header.validator', items.value.length)}`,
} ]
</script>

<template>
  <div class="header-container">
    <BcMenuBar
      v-if="items.length"
      class="menu-bar"
      :buttons
    />
    <BcButtonIcon
      v-if="variant !== 'shared-dashboard'"
      name="plus"
      screenreader-text="dashboard.title"
      variant="flat"
      @click="emit('showCreation')"
    />
  </div>
</template>

<style lang="scss" scoped>
.header-container {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  margin-top: var(--padding);
  padding-bottom: var(--padding-medium);
  margin-bottom: var(--padding-medium);
  min-width: 1px;
  gap: var(--padding);
  border-bottom: var(--container-border);

  .edit_button {
    border-color: var(--container-border-color);
    background-color: var(--container-background);
    color: var(--container-color);
    flex-shrink: 0;
  }

  .menu-bar {
    display: flex;
    flex-shrink: 1;
    overflow: hidden;
  }

  :deep(.p-menubar-root-list >.p-menuitem ) {
    min-width: 102px;
    >.p-menuitem-content:not(:has(.toggle)) {
      .button-content {
        justify-content: center;
      }
    }
  }

  :deep(.p-menubar-root-list .p-menuitem .p-submenu-list) {
    position: fixed;
  }

  @media (max-width: 519px) {
    gap: var(--padding-small);

    :deep(.p-menubar-root-list) {
      gap: var(--padding-small);
    }

    :deep(.p-menubar-root-list > .p-menuitem > .p-menuitem-content) {
      padding: var(--padding-small) var(--padding);
    }
  }
}
</style>
