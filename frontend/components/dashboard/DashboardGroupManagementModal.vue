<script lang="ts" setup>
import { orderBy } from 'lodash-es'
import {
  BcDialogConfirm, BcPremiumModal,
} from '#components'
import type { ApiPagingResponse } from '~/types/api/common'
import type { VDBOverviewGroup } from '~/types/api/validator_dashboard'

const { t: $t } = useTranslation()
const {
  width,
} = useWindowSize()
const {
  groups,
  key,
  variant,
} = useDashboard()

const visible = defineModel<boolean>()

const { dashboards } = storeToRefs(useUserDashboardStore())

const query = useDefaultQuery({
  sort: 'name:asc',
})

const data = computed<ApiPagingResponse<VDBOverviewGroup>>(() => {
  let processedGroups = groups.value

  if (query.value.search?.length) {
    const search = query.value.search.toLowerCase()
    processedGroups = processedGroups.filter(
      group => group.name.toLowerCase().includes(search) || parseInt(search) === group.id,
    )
  }

  const [
    sortField,
    sortOrder,
  ] = (query.value.sort?.split(':') ?? []) as [string, 'asc' | 'desc']

  if (sortField === 'name') {
    // lodash needs some help when sorting strings alphabetically
    processedGroups = orderBy(
      processedGroups,
      [ g => g.name.toLowerCase() ],
      sortOrder,
    )
  }
  else {
    processedGroups = orderBy(
      processedGroups,
      sortField,
      sortOrder,
    )
  }

  return {
    data: processedGroups,
    paging: {
      total_count: processedGroups.length,
    },
  }
})

const showSubTitle = computed(() => {
  return width.value >= 760
})

const resetData = () => {
  query.value.search = ''
  newGroupName.value = ''
}

const onClose = () => {
  visible.value = false
  resetData()
}

const dialog = useDialog()
const toast = useBcToast()

const emit = defineEmits<{
  (e: 'change-groups', value: VDBOverviewGroup[]): void,
}>()

const newGroupName = ref<string>('')
const hasOpenDialogs = ref(false)
// Keep a local copy of the groups, which won't change when we optimistically update them,
// so we can use them to revert changes if the API call fails.
const oldGroups = ref<VDBOverviewGroup[]>([ ...groups.value ])

const newGroupDisabled = computed(
  () => !REGEXP_VALID_NAME.test(newGroupName.value),
)

const addGroup = async () => {
  newGroupName.value = newGroupName.value.trim()

  if (premiumLimit.value) {
    dialog.open(BcPremiumModal, {})
    return
  }

  const optimisticGroups: VDBOverviewGroup[] = [
    ...groups.value,
    {
      count: 0,
      id: -1,
      name: newGroupName.value,
    },
  ]
  // Optimistically update the dashboard groups to give immediate feedback
  emit('change-groups', optimisticGroups)
  // Keep a copy of the new group name and clear input field
  const tempNewGroupName = newGroupName.value
  newGroupName.value = ''

  await $api(`/api/bff/validator-dashboards/${key.value}/groups`, {
    body: { name: tempNewGroupName },
    method: 'POST',
  }).then((res) => {
    // We need to replace the optimistic group with the real one, as it contains the real ID
    const newGroups = [
      ...oldGroups.value,
      res,
    ]
    oldGroups.value = newGroups
    emit('change-groups', newGroups)
  }).catch((error) => {
    // If the API call fails, revert the optimistic response
    emit('change-groups', oldGroups.value)
    if (error.statusCode === 409) {
      toast.showError({
        summary: $t('dashboard.validator.group_management.errors.duplicate'),
      })
      return
    }
    toast.showError({
      summary: $t('dashboard.validator.group_management.errors.add'),
    })
  })
}

const editGroup = async (row: VDBOverviewGroup, newName: string) => {
  const optimisticGroups: VDBOverviewGroup[] = groups.value.map((group) => {
    if (group.id === row.id) {
      return {
        ...group, name: newName,
      }
    }
    return group
  })
  // Optimistically update the dashboard groups to give immediate feedback
  emit('change-groups', optimisticGroups)

  await $api(`/api/bff/validator-dashboards/${key.value}/groups/${row.id}`, {
    body: { name: newName },
    method: 'PUT',
  }).catch(() => {
    // If the API call fails, revert the optimistic response
    emit('change-groups', oldGroups.value)
    toast.showError({
      summary: $t('dashboard.validator.group_management.errors.edit'),
    })
  })
}

const removeGroupConfirmed = async (row: VDBOverviewGroup) => {
  const optimisticGroups: VDBOverviewGroup[] = groups.value.filter((group) => {
    return group.id !== row.id
  })
  // Optimistically update the dashboard groups to give immediate feedback
  emit('change-groups', optimisticGroups)
  await $api(`/api/bff/validator-dashboards/${key.value}/groups/${row.id}`, {
    method: 'DELETE',
  }).then(() => oldGroups.value = optimisticGroups,
  ).catch(() => {
    // If the API call fails, revert the optimistic response
    emit('change-groups', oldGroups.value)
    toast.showError({
      summary: $t('dashboard.validator.group_management.errors.remove'),
    })
  })
}

const removeGroup = (row: VDBOverviewGroup) => {
  hasOpenDialogs.value = true
  dialog.open(BcDialogConfirm, {
    data: {
      question: $t('dashboard.validator.group_management.remove_text', { group: row.name }),
      title: $t('dashboard.validator.group_management.remove_title'),
    },
    onClose: (response) => {
      hasOpenDialogs.value = false
      if (response?.data) {
        removeGroupConfirmed(row)
      }
    },
  })
}

const { $api } = useNuxtApp()

const dashboardName = computed(() => {
  return (
    dashboards.value?.validator_dashboards?.find(
      d => `${d.id}` === key.value,
    )?.name || $t('dashboard.validator.group_management.your_dashboard')
  )
})

const user = useFetchedData('user')

const maxGroupsPerDashboard = computed(() =>
  variant.value === 'guest-dashboard' || !user.value?.premium_perks?.validator_groups_per_dashboard
    ? 1
    : user.value.premium_perks.validator_groups_per_dashboard,
)
const premiumLimit = computed(
  () => (data.value?.paging?.total_count ?? 0) >= maxGroupsPerDashboard.value,
)

const isMobile = computed(() => {
  return (width.value ?? 0) <= 800
})
</script>

<template>
  <BcDialog
    v-model="visible"
    :close-on-escape="!hasOpenDialogs"
    :header="$t('dashboard.validator.group_management.title')"
    class="validator-group-managment-modal-container"
    @update:visible="(visible: boolean) => !visible && resetData()"
  >
    <template
      v-if="!showSubTitle"
      #header
    >
      <span />
    </template>
    <BcTableControl
      v-model:search="query.search"
      :search-placeholder="
        $t('dashboard.validator.group_management.search_placeholder')
      "
      :disabled-filter="variant === 'guest-dashboard'"
    >
      <template #header-left>
        <span v-if="showSubTitle">
          {{
            $t("dashboard.validator.group_management.sub_title", {
              dashboardName,
            })
          }}
        </span>
        <span
          v-else
          class="small-title"
        >
          {{
            $t("dashboard.validator.group_management.title")
          }}
        </span>
      </template>
      <template #bc-table-sub-header>
        <form
          class="add-row"
          @submit.prevent="addGroup"
        >
          <InputText
            v-model="newGroupName"
            class="search-input"
            maxlength="20"
            :placeholder="
              $t('dashboard.validator.group_management.new_group_placeholder')
            "
          />
          <Button
            style="display: inline"
            type="submit"
            :disabled="newGroupDisabled"
          >
            <BcIcon name="plus" />
          </Button>
        </form>
      </template>
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            :data
            class="management-table"
            :query
            hide-pager
            data-key="id"
          >
            <Column
              field="name"
              class="edit-group"
              :sortable="true"
              :header="$t('dashboard.validator.group_management.col.name')"
            >
              <template #body="slotProps">
                <DashboardValidatorManagementModalGroupLabel
                  class="edit-group truncate-text"
                  :value="slotProps.data.name"
                  :default="
                    slotProps.data.id === 0
                      ? $t('dashboard.group.selection.default')
                      : ''
                  "
                  :can-be-empty="false"
                  :disabled="variant === 'guest-dashboard'"
                  :pattern="REGEXP_VALID_NAME"
                  :maxlength="20"
                  @set-value="(name: string) => editGroup(slotProps.data, name)"
                />
              </template>
            </Column>
            <Column
              field="id"
              :sortable="!isMobile"
              :header="$t('dashboard.validator.group_management.col.id')"
            >
              <template #body="slotProps">
                <div class="id-cell">
                  {{ slotProps.data.id }}
                </div>
              </template>
            </Column>
            <Column
              field="count"
              :sortable="!isMobile"
              :header="$t('dashboard.validator.group_management.col.count')"
            >
              <template #body="slotProps">
                <BcFormatNumber
                  :value="slotProps.data.count"
                  default="0"
                />
              </template>
            </Column>
            <Column field="action">
              <template #body="slotProps">
                <div class="action-col">
                  <BcButtonIcon
                    v-if="slotProps.data.id"
                    :screenreader-text="{
                      interpolation: { groupName: slotProps.data.name },
                      key: 'dashboard.validator.group_management.remove_validator',
                    }"
                    name="trash"
                    class="link"
                    @click="removeGroup(slotProps.data)"
                  />
                </div>
              </template>
            </Column>

            <template #bc-table-footer-bottom>
              <div class="validator-group-managment-modal__footer">
                <div class="left">
                  <div
                    class="labels"
                    :class="{ premiumLimit }"
                  >
                    <span>
                      <BcFormatNumber
                        :value="groups.length"
                        default="0"
                      />
                      /
                      <BcFormatNumber :value="maxGroupsPerDashboard" />
                    </span>
                  </div>
                  <BcPremiumGem />
                </div>
                <Button
                  :label="$t('navigation.done')"
                  @click="onClose"
                />
              </div>
            </template>
          </BcTable>
        </ClientOnly>
      </template>
    </BcTableControl>
  </BcDialog>
</template>

<style lang="scss" scoped>
@use "~/assets/css/main.scss";
@use "~/assets/css/utils.scss";
@use "~/assets/css/fonts.scss";

:global(.validator-group-managment-modal-container) {
  width: 960px;
  height: 800px;
}

:global(.validator-group-managment-modal-container .p-dialog-content) {
  display: flex;
  flex-direction: column;
  flex-grow: 1;
}

:global(.validator-group-managment-modal-container .bc-table-header) {
  height: unset !important;
  padding: var(--padding) 0 !important;
  @include fonts.subtitle_text;
}

:global(
    .validator-group-managment-modal-container
      .bc-table-header
      .side:first-child
  ) {
  display: contents;
}

:global(.validator-group-managment-modal-container .edit-group) {
  max-width: 201px;
  height: 27px;
  width: 201px;
}

.edit-group {
  max-width: 180px;
}

.id-cell {
  @include utils.set-all-width(64px);
}

.small-title {
  @include utils.truncate-text;
  @include fonts.big_text;
}

.management-table {
  @include main.container;
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  overflow-y: hidden;

  :deep(.p-datatable-wrapper) {
    flex-grow: 1;
  }
}

.add-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--padding);

  .search-input {
    padding: var(--padding-small);
    flex-shrink: 1;
    flex-grow: 1;
    width: 50px;
    border-top-right-radius: 0;
    border-bottom-right-radius: 0;
    height: 100%;
  }

  button {
    border-top-left-radius: 0;
    border-bottom-left-radius: 0;
    padding: var(--padding-small) 8px;
  }
}

.left {
  display: flex;
  margin-top: 4px;
  gap: var(--padding-small);

  .labels {
    display: flex;
    gap: var(--padding-small);

    &.premiumLimit {
      color: var(--negative-color);
    }

    @media (max-width: 450px) {
      flex-direction: column;
    }
  }

  .gem {
    color: var(--primary-color);
  }
}

.edit-icon {
  margin-left: var(--padding-small);
}

.action-col {
  width: 100%;
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 560px) {
  .edit-group {
    max-width: 100px;
  }

  .action-col {
    width: 33px;
  }
}

:global(.validator-group-managment-modal__footer) {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--padding-medium);
}
</style>
