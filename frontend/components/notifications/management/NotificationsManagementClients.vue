<script lang="ts" setup>
const { t: $t } = useTranslation()

const notificationsManagementStore = useNotificationsManagementStore()
await notificationsManagementStore.getSettings()

const executionClients = computed(
  () => notificationsManagementStore.settings.clients.filter(client => client.category === 'execution_layer'),
)
const consensusClients = computed(
  () => notificationsManagementStore.settings.clients.filter(client => client.category === 'consensus_layer'),
)
const otherClients = computed(
  () => notificationsManagementStore.settings.clients.filter(client => client.category === 'other'),
)

const setNotificationForClient = ({
  id,
  value,
}: {
  id: number,
  value: boolean,
},
) => {
  notificationsManagementStore.setNotificationForClient({
    client_id: id,
    is_subscribed: value,
  })
}
</script>

<template>
  <BcTabPanel>
    <BcListSection>
      <span class="grid-span-2">
        {{ $t('notifications.clients.settings.execution_clients') }}
      </span>
      <BcDropdownToggle
        :options="executionClients"
        option-label="name"
        option-value="is_subscribed"
        option-identifier="id"
        screenreader-heading="notifications.clients.settings.execution_clients"
        :text="$t('notifications.clients.settings.clients', [executionClients.length])"
        @change="setNotificationForClient"
      />
      <span class="grid-span-2">
        {{ $t('notifications.clients.settings.consensus_clients') }}
      </span>
      <BcDropdownToggle
        :options="consensusClients"
        option-label="name"
        option-value="is_subscribed"
        option-identifier="id"
        screenreader-heading="notifications.clients.settings.consensus_clients"
        :text="$t('notifications.clients.settings.clients', [consensusClients.length])"
        @change="setNotificationForClient"
      />
      <span class="grid-span-2">
        {{ $t('notifications.clients.settings.other_clients') }}
      </span>
      <BcDropdownToggle
        :options="otherClients"
        option-label="name"
        option-value="is_subscribed"
        option-identifier="id"
        screenreader-heading="notifications.clients.settings.other_clients"
        :text="$t('notifications.clients.settings.clients', [otherClients.length])"
        @change="setNotificationForClient"
      />
    </BcListSection>
  </BcTabPanel>
</template>

<style lang="scss" scoped>
.toggle {
  justify-content: end;
  margin: 0;
}
.grid-span-2{
  grid-column: span 2;
}
</style>
