<script lang="ts" setup>
const { variant } = useDashboard()
const { validatorDashboards } = usePrivateDashboards()
const { t: $t } = useTranslation()
const route = useRoute()

const dismissed = ref(false)
const visible = computed(
  () =>
    variant.value === 'shared-dashboard' && !dismissed.value,
)

const text = computed(() => {
  const userHasOwnDashboard = validatorDashboards.value?.length > 0
  const textRoot = userHasOwnDashboard
    ? 'dashboard.shared_modal_with_own'
    : 'dashboard.shared_modal_without_own'

  const caption = $t(textRoot + '.text')
  const button = $t(textRoot + '.button')

  return {
    button,
    caption,
  }
})
</script>

<template>
  <Dialog
    v-model:visible="visible"
    :dismissable-mask="false"
    :draggable="false"
    :close-on-escape="false"
    position="bottom"
  >
    <div class="dialog-container">
      {{ text.caption }}
      <div class="button-row">
        <div
          class="dismiss"
          @click="dismissed = true"
        >
          {{ $t("navigation.dismiss") }}
        </div>
        <BcLink
          :to="`/dashboard`"
          :replace="route.path.startsWith('/dashboard')"
        >
          <Button>
            {{ text.button }}
          </Button>
        </BcLink>
      </div>
    </div>
  </Dialog>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";

.dialog-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--padding);

  .button-row {
    display: flex;
    align-items: center;
    gap: var(--padding-large);

    .dismiss {
      cursor: pointer;
      color: var(--text-color-disabled);
    }
  }
}
</style>
