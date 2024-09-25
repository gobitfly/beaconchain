<script setup lang="ts">
import { useNotificationsClientStore } from '~/stores/notifications/useNotificationsClientsStore'
import type { Cursor } from '~/types/datatable'

defineEmits<{ (e: 'openDialog'): void }>()

const cursor = ref<Cursor>()
const pageSize = ref<number>(10)
const { t: $t } = useTranslation()

const {
  clientsNotifications,
  isLoading,
  onSort,
  query,
  setCursor,
  setPageSize,
  setSearch,
} = useNotificationsClientStore()

const colsVisible = computed(() => {
  return {
    footer: 1024,
  }
})

const { overview } = useNotificationsDashboardOverviewStore()
</script>

<template>
  <div>
    <BcTableControl
      :title="$t('notifications.clients.title')"
      :search-placeholder="$t('notifications.clients.search_placeholder')"
      @set-search="setSearch"
    >
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            :data="clientsNotifications"
            data-key="client_name"
            :cursor
            :page-size
            :selected-sort="query?.sort"
            :loading="isLoading"
            @set-cursor="setCursor"
            @sort="onSort"
            @set-page-size="setPageSize"
          >
            <Column
              sortable
              header-class="col-client-name"
              body-class="col-client-name"
              :header="$t('notifications.clients.col.client_name')"
            >
              <template #body="slotProps">
                {{ slotProps.data.client_name }}
              </template>
            </Column>
            <Column
              header-class="col-version"
              body-class="col-version"
              :header="$t('notifications.clients.col.version')"
            >
              <template #body="slotProps">
                <BcLink
                  :to="`${slotProps.data.url}`"
                  class="link"
                  target="_blank"
                  external
                >
                  {{ slotProps.data.version }}
                </BcLink>
              </template>
            </Column>
            <Column
              field="timestamp"
              sortable
              header-class="col-age"
              body-class="col-age"
            >
              <template #header>
                <BcTableAgeHeader />
              </template>
              <template #body="slotProps">
                <BcFormatTimePassed
                  :value="slotProps.data.timestamp"
                />
              </template>
            </Column>
            <template #empty>
              <NotificationsDashboardsTableEmpty
                v-if="!clientsNotifications?.data.length"
                @open-dialog="$emit('openDialog')"
              />
            </template>
            <!-- TODO: implement number of clients subscriptions -->
            <template #bc-table-footer-right>
              <template v-if="colsVisible">
                {{ $t('notifications.clients.footer.subscriptions', { count: overview?.clients_subscription_count }) }}
              </template>
            </template>
          </BcTable>
        </ClientOnly>
      </template>
    </BcTableControl>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/main.scss";
@use "~/assets/css/utils.scss";

$breakpoint-sm: 630px;
$breakpoint-md: 780px;
$breakpoint-lg: 1024px;

:deep(.col-client-name) {

  @include utils.truncate-text;

  @media (max-width: $breakpoint-lg) {
    @include utils.set-all-width(240px);
  }

  @media (max-width: $breakpoint-md) {
    @include utils.set-all-width(200px);
  }

  @media (max-width: $breakpoint-sm) {
    @include utils.set-all-width(106px);
  }
}

:deep(.col-version) {

  @include utils.truncate-text;

  @media (max-width: $breakpoint-lg) {
    @include utils.set-all-width(140px);
  }

  @media (max-width: $breakpoint-sm) {
    @include utils.set-all-width(78px);
  }

  // *:not([data-pc-section="sort"]) {
  // }
}

:deep(.col-age) {
  @media (max-width: $breakpoint-lg) {
    @include utils.set-all-width(140px);
  }

  @media (max-width: $breakpoint-sm) {
    @include utils.set-all-width(78px);
  }

  *:not([data-pc-section="sort"]) {
    @include utils.truncate-text;
  }
}

.management-table {
  @include main.container;
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  overflow-y: hidden;
  justify-content: space-between;

  :deep(.p-datatable-wrapper) {
    flex-grow: 1;
  }
}

td > .col-client-name > a{
  text-overflow: ellipsis !important;
}

td > .col-client-name {
  text-overflow: ellipsis !important;
}
</style>
