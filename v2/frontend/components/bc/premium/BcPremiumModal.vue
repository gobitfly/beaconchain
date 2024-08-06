<script lang="ts" setup>
interface Props {
  description?: string,
  dismissLabel?: string
}

const { props, dialogRef, setHeader } = useBcDialog<Props>({ contentClass: 'premium-modal' })
const { t: $t } = useTranslation()

const hide = () => {
  dialogRef?.value.close()
}

onMounted(() => {
  setHeader($t('premium.title'), true)
})

</script>

<template>
  <div class="text">
    {{ props?.description || $t('premium.description') }}
  </div>
  <div class="footer">
    <div class="dismiss" @click="hide()">
      {{ props?.dismissLabel || $t('navigation.dismiss') }}
    </div>
    <BcLink to="/pricing" target="_blank" @click="hide()">
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

:global(.p-dialog:has(.premium-modal)>.p-dialog-header) {
  color: var(--primary-color);
  font-size: var(--subtitle_font_size);
}
</style>
