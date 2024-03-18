<script lang="ts" setup>
import {
  faGem
} from '@fortawesome/pro-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

const visible = ref<boolean>()

</script>

<template>
  <BcTooltip :text="$t('premium.subscribe')">
    <FontAwesomeIcon :icon="faGem" class="gem" @click.stop.prevent="visible = true" />
    <BcDialog
      v-model="visible"
      :header="$t('premium.title')"
      :pt="{
        header: {
          class: 'premium-header'
        }
      }"
      class="validator-managment-modal-container"
    >
      <div class="text">
        {{ $t('premium.description') }}
      </div>
      <template #footer>
        <div class="footer">
          <div class="dismiss" @click="visible = false">
            {{ $t('navigation.dismiss') }}
          </div>
          <NuxtLink to="/premium/subscription">
            <Button :label="$t('premium.unlock')" />
          </NuxtLink>
        </div>
      </template>
    </BcDialog>
  </BcTooltip>
</template>

<style lang="scss" scoped>
.dismiss {
  cursor: pointer;
  color: var(--text-color-disabled);
}

.gem {
  color: var(--primary-color);
  cursor: pointer;
}

.text {
  font-size: var(--big_text_label_font_size);
  font-weight: var(-big_text_label_bold_font_weight);
  padding: 15px 0 28px 0;
}

.footer {
  display: flex;
  gap: 18px;
  align-items: center;
  justify-content: flex-end;
}

:global(.p-dialog .p-dialog-header.premium-header .p-dialog-title) {
  color: var(--primary-color);
  font-size: var(--subtitle_font_size);
}
</style>
