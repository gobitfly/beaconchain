<script lang="ts" setup>
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faDesktop, faUser,
} from '@fortawesome/pro-solid-svg-icons'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'

const { isLoggedIn } = useUserStore()
const { t: $t } = useTranslation()
const {
  overview,
  refreshOverview,
} = useNotificationsDashboardOverviewStore()

if (isLoggedIn.value) {
  refreshOverview()
}

const hasEmail = computed(() => overview.value?.is_email_notifications_enabled)
const hasPushNotifications = computed(() => overview.value?.is_push_notifications_enabled)
const vdbMostNotifiedGroups = computed(() => overview.value?.vdb_most_notified_groups || [])
const adbMostNotifiedGroups = computed(() => overview.value?.adb_most_notified_groups || [])
const last24hEmailsCount = computed(() => overview.value?.last_24h_email_count ?? 0)
const last24hPushCount = computed(() => overview.value?.last_24h_push_count ?? 0)
const last24hWebhookCount = computed(() => overview.value?.last_24h_webhook_count ?? 0)
const notificationsTotal = computed(() => {
  return last24hEmailsCount.value + last24hWebhookCount.value + last24hPushCount.value
})

const { user } = useUserStore()
const mailLimit = computed(() => user.value?.premium_perks.email_notifications_per_day ?? 0)

const resetHours = computed(
  () => getRelativeTime(overview.value?.next_email_count_reset_timestamp ?? 0),
)

const tooltipEmail = computed(() => {
  return $t('notifications.overview.email_tooltip', {
    in_x_hours: resetHours.value,
    limit: mailLimit.value,
  })
})
const emit = defineEmits<{
  (e: 'openDialog'): void,
}>()
</script>

<template>
  <div class="container">
    <div class="box">
      <section class="box-item">
        <h3 class="big_text_label">
          {{ $t('notifications.overview.headers.email_notifications') }}
        </h3>
        <div
          class="big_text"
        >
          {{ hasEmail ? $t('common.active') : $t('common.inactive') }}
        </div>
        <div v-if="hasEmail" class="inline-items">
          <span
            class="small_text"
            :class="{ is_limit_reached: last24hEmailsCount >= mailLimit }"
          >{{ last24hEmailsCount }}/{{ mailLimit }} {{ $t('common.units.per_day') }}</span>
          <BcTooltip
            tooltip-width="220px"
            :text="tooltipEmail"
          >
            <FontAwesomeIcon :icon="faInfoCircle" />
          </BcTooltip>
          <BcPremiumGem class="gem" />
        </div>
        <div v-else class="premium-invitation small_text">
          <BcTranslation
            keypath="notifications.overview.notifications_activate_premium.template"
            linkpath="notifications.overview.notifications_activate_premium._link"
            to="https://discord.com/developers/docs/resources/webhook"
          >
            <template #_link>
              <BcButtonText
                class="link"
                :aria-label="$t('notifications.overview.email_activate')"
                @click="emit('openDialog')"
              >
                {{ $t('notifications.overview.notifications_activate_premium._link') }}
              </BcButtonText>
            </template>
          </BcTranslation>
        </div>
      </section>
      <section class="box-item">
        <h3 class="overwrite-h3 big_text_label">
          {{ $t('notifications.overview.headers.push_notifications') }}
        </h3>
        <div class="big_text">
          {{ hasPushNotifications ? $t('common.active') : $t('common.inactive') }}
        </div>
        <div v-if="!hasPushNotifications" class="push-invitation small_text">
          <BcTranslation
            keypath="notifications.overview.notifications_download_app.template"
            linkpath="notifications.overview.notifications_download_app._link"
            to="/mobile"
          />
        </div>
      </section>
      <section class="box-item">
        <h3 class="overwrite-h3 big_text_label">
          {{ $t('notifications.overview.headers.most_notifications_30d') }}
        </h3>
        <div class="lists-container">
          <div class="lists-container-column">
            <BcScreenreaderOnly tag="h4">
              {{ $t('notifications.overview.headers.validator_groups') }}
            </BcScreenreaderOnly>
            <ol class="icon-list">
              <li
                v-for="(group, index) in vdbMostNotifiedGroups"
                :key="group"
                class="small_text list-item"
              >
                <FontAwesomeIcon :icon="faDesktop" />
                <span class="list-text">
                  {{ index + 1 }}. {{ group || '-' }}
                </span>
              </li>
            </ol>
          </div>
          <BcFeatureFlag feature="feature-account_dashboards">
            <div class="lists-container-column">
              <BcScreenreaderOnly tag="h4">
                {{ $t('notifications.overview.headers.account_groups') }}
              </BcScreenreaderOnly>
              <ol class="icon-list">
                <li
                  v-for="(group, index) in adbMostNotifiedGroups"
                  :key="group"
                  class="small_text list-item"
                >
                  <FontAwesomeIcon :icon="faUser" />
                  <span class="list-text">
                    {{ index + 1 }}. {{ group || '-' }}
                  </span>
                </li>
              </ol>
            </div>
          </BcFeatureFlag>
        </div>
      </section>
      <section class="box-item">
        <h3 class="overwrite-h3 big_text_label">
          {{ $t('notifications.overview.headers.notifications_24h') }}
        </h3>
        <div class="big_text">
          {{ notificationsTotal }}
        </div>
        <div class="small_text">
          {{ last24hEmailsCount }} {{ $t('common.email') }} | {{ last24hWebhookCount }} {{ $t('common.webhook') }} | {{ last24hPushCount }} {{ $t('notifications.overview.push') }}
        </div>
      </section>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';
@use "~/assets/css/utils.scss";

.container {
  @include main.container;
  padding: 1.0625rem 1.25rem;
  position: relative;
}
.info-section, .action-section {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: .625rem;
}
.icon {
  font-size: 1.5rem;
}
.text {
  font-size: 1.125rem;
  font-weight: 500;
}
.list-item {
  display: flex;
  gap: .625rem;
  .list-text {
    @include utils.truncate-text;
  }
}
.box {
  display: flex;
  justify-content: space-between;
  overflow: auto;
  scrollbar-width: none;
  gap: 1.25rem;

  &::-webkit-scrollbar {
    display: none;
  }
  .box-item {
    flex-shrink: 0;
    max-width: 17rem;

  }
}
.box-item {
  display: flex;
  flex-direction: column;
  gap: .625rem;
}
.inline-items {
  display: flex;
  align-items: center;
  gap: .625rem;
}
.is_limit_reached {
  color: var(--negative-color);
}
a:hover {
  color: var(--light-blue);
}
.lists-container {
  display: flex;
  gap: 1.25rem;
}
.lists-container-column {
  flex: 1;
  min-width: 0;
}
.icon-list {
  min-width: 0;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: .625rem;
}
.icon {
  font-size: 1rem;
}
.inline-link,
.gem {
  display: inline-block;
}
.premium-invitation {
  display: flex;
  align-items: center;
  gap: .3125rem;
}
.push-invitation {
  display: flex;
  align-items: center;
  gap: .3125rem;
  flex-wrap: wrap;
}
@media (max-width: 600px) {
  .box {
    flex-direction: row;
    gap: 1.25rem;
  }
  .box-item {
    min-width: 15.625rem;
  }
}
</style>
