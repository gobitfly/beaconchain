<script lang="ts" setup>
import type { BcButton } from '#components'

interface Props {
  description?: string,
  dismissLabel?: string,
}

const {
  dialogRef, props, setHeader,
} = useBcDialog<Props>({ contentClass: 'premium-modal' })
const { t: $t } = useTranslation()

const hide = () => {
  dialogRef?.value.close()
}

const buttonDismiss = ref<typeof BcButton>()
const lastActiveElement = ref<Element | null>(null)
onBeforeMount(() => {
  lastActiveElement.value = document.activeElement
})
onMounted(() => {
  setHeader($t('premium.title'), true)
  buttonDismiss.value?.$el?.focus()
})
onUnmounted(() => {
  if (lastActiveElement.value instanceof HTMLElement) {
    lastActiveElement.value.focus()
  }
})
</script>

<template>
  <div class="text">
    {{ props?.description || $t("premium.description") }}
  </div>
  <div
    class="footer"
    @keydown.esc.stop="hide()"
  >
    <BcButton
      ref="buttonDismiss"
      variant="secondary"
      class="dismiss"
      @click="hide()"
    >
      {{ props?.dismissLabel || $t("navigation.dismiss") }}
    </BcButton>
    <BcLink
      to="/pricing"
      target="_blank"
      @click="hide()"
    >
      <Button :label="$t('premium.unlock')" />
    </BcLink>
  </div>
</template>

<style lang="scss" scoped>
:global(.premium-modal) {
  width: 620px;
  max-width: 100%;
}

.text {
  font-family: var(--subtitle_font_family);
  font-weight: var(--subtitle_font_weight);
  font-size: var(--subtitle_font_size);
  padding: 15px 0 28px 0;
}

.footer {
  display: flex;
  gap: 18px;
  align-items: center;
  justify-content: flex-end;

  .dismiss {
    cursor: pointer;
    color: var(--text-color-disabled);
  }
}

:global(.p-dialog:has(.premium-modal) > .p-dialog-header) {
  color: var(--primary-color);
  font-size: var(--subtitle_font_size);
}
</style>
