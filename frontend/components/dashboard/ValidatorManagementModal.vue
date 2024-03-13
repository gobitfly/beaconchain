<script lang="ts" setup>
import {
  faAdd,
  faTrash
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import type { DataTableSortEvent } from 'primevue/datatable'
import { warn } from 'vue'
import type { InternalGetValidatorDashboardValidatorsResponse, VDBManageValidatorsTableRow } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { Cursor } from '~/types/datatable'

const { t: $t } = useI18n()

interface Props {
  dashboardKey: DashboardKey;
}
const props = defineProps<Props>()

const { width } = useWindowSize()

const visible = defineModel<boolean>()

const { value: query, bounce: setQuery } = useDebounceValue<PathValues | undefined>(undefined, 500)

const cursor = ref<Cursor>()
const pageSize = ref<number>(5)
const selectedGroup = ref<number>(-1)
const selectedValidator = ref<string>('')

const data = ref<InternalGetValidatorDashboardValidatorsResponse | undefined>()
const selected = ref<VDBManageValidatorsTableRow[]>()

const size = computed(() => {
  return {
    expandable: width.value < 960
  }
})

const onClose = () => {
  visible.value = false
}

const addValidator = () => {
  // TODO call API to add Validator
  alert(`Add validator ${selectedValidator.value} for ${selectedGroup.value}`)
}

const onSort = (sort: DataTableSortEvent) => {
  setQuery(setQuerySort(sort, query?.value))
}

const setCursor = (value: Cursor) => {
  cursor.value = value
  setQuery(setQueryCursor(value, query?.value))
}

const setPageSize = (value: number) => {
  pageSize.value = value
  setQuery(setQueryPageSize(value, query?.value))
}

const setSearch = (value?: string) => {
  setQuery(setQuerySearch(value, query?.value))
}

watch(selectedGroup, (value) => {
  setQuery({ ...query?.value, group_id: value })
})

watch(() => [props.dashboardKey, visible.value, query.value], async () => {
  if (props.dashboardKey && visible.value) {
    data.value = await useCustomFetch<InternalGetValidatorDashboardValidatorsResponse>(API_PATH.DASHBOARD_VALIDATOR_MANAGEMENT, undefined, { dashboardKey: props.dashboardKey }, query.value)
  }
}, { immediate: true })

const switchValidatorGroup = (row:VDBManageValidatorsTableRow, group: number) => {
  alert(`switchGroup ${group} for ${row.index}`)
}

const removeRow = (row:VDBManageValidatorsTableRow) => {
  alert(`remove val ${row.index}`)
}

</script>

<template>
  <BcDialog
    v-model="visible"
    :header="$t('dashboard.validator.management.title')"
    class="validator-managment-modal-container"
  >
    <BcTableControl :search-placeholder="$t('dashboard.validator.summary.search_placeholder')" @set-search="setSearch">
      <template #header-left>
        <span>{{ $t('dashboard.validator.management.sub_title') }}</span>
      </template>
      <template #bc-table-sub-header>
        <div class="add-row">
          <DashboardGroupSelection v-model="selectedGroup" :include-all="true" class="small" />
          <!-- TODO: replace input once Searchbar is finished -->
          <InputText v-model="selectedValidator" style="flex-grow: 1;" />
          <Button class="p-button-icon-only" style="display: inline;" @click="addValidator">
            <FontAwesomeIcon :icon="faAdd" />
          </Button>
          <!-- end of temp -->
        </div>
      </template>
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            v-model:selection="selected"
            :data="data"
            data-key="group_id"
            :expandable="size.expandable"
            selection-mode="multiple"
            class="management-table"
            :cursor="cursor"
            :page-size="pageSize"
            @set-cursor="setCursor"
            @sort="onSort"
            @set-page-size="setPageSize"
          >
            <Column
              field="index"
              :sortable="true"
              :header="$t('dashboard.validator.col.index')"
            />

            <Column
              field="public_key"
              :sortable="true"
              :header="$t('dashboard.validator.col.public_key')"
            >
              <template #body="slotProps">
                <BcFormatHash :hash="slotProps.data.public_key" type="public_key" />
              </template>
            </Column>
            <Column
              field="group_id"
              :sortable="true"
              :header="$t('dashboard.validator.col.group')"
            >
              <template #body="slotProps">
                <DashboardGroupSelection v-model="slotProps.data.group_id" class="small" @set-group="(id:number)=>switchValidatorGroup(slotProps.data, id)" />
              </template>
            </Column>

            <Column
              field="balance"
              :sortable="true"
              :header="$t('dashboard.validator.col.balance')"
            >
              <template #body="slotProps">
                <BcFormatValue :value="slotProps.data.balance" />
              </template>
            </Column>
            <Column
              field="status"
              :sortable="true"
              :header="$t('dashboard.validator.col.status')"
            >
              <template #body="slotProps">
                <ValidatorTableStatus :status="slotProps.data.status" :position="1" />
              </template>
            </Column>
            <Column
              field="withdrawal_credential"
              :sortable="true"
              :header="$t('dashboard.validator.col.withdrawal_credential')"
            >
              <template #body="slotProps">
                <BcFormatHash :hash="slotProps.data.withdrawal_credential" type="withdrawal_credentials" />
              </template>
            </Column>
            <Column
              field="action"
            >
              <template #body="slotProps">
                <FontAwesomeIcon :icon="faTrash" class="link" @click="removeRow(slotProps.data)" />
              </template>
            </Column>
          </BcTable>
        </ClientOnly>
      </template>
    </BcTableControl>
    <template #footer>
      <div class="footer">
        <span>TODO left</span>
        <Button :label="$t('navigation.done')" @click="onClose" />
      </div>
    </template>
  </BcDialog>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';

:global(.validator-managment-modal-container) {
  width: 960px;
  height: 800px;

}

:global(.validator-managment-modal-container .p-dialog-content) {
  display: flex;
  flex-direction: column;
  flex-grow: 1;
}

:global(.validator-managment-modal-container .bc-table-header) {
  height: unset;
  padding: var(--padding) 0;
}

/* TODO: set width for each col independently */
.management-table {
  --col-width: 160px;
  :deep(td) {
    &:not(.expander):not(:selection):not(:last-child) {
      max-width: var(--col-width);
      width: var(--col-width);
    }
  }
  :deep(th) {
    &:not(.expander):not(:selection):not(:last-child) {
      max-width: var(--col-width);
      width: var(--col-width);
    }
  }
}

.management-table {
  @include main.container;
  flex-grow: 1;
  display: flex;
  flex-direction: column;

  :deep(.p-datatable-wrapper) {
    flex-grow: 1;
  }
}

.add-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--padding);
  gap: var(--padding);
}

.footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: var(--padding);
  gap: var(--padding);
}
</style>
