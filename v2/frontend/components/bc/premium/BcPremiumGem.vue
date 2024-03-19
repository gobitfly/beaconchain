<script lang="ts" setup>
import {
  faGem
} from '@fortawesome/pro-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

interface Props {
  description?: string,
  dismissLabel?: string
}
defineProps<Props>()

const visible = ref<boolean>()

</script>

<template>
  <BcTooltip :text="$t('premium.subscribe')">
    <div @click.stop.prevent="visible = true">
      <FontAwesomeIcon :icon="faGem" class="gem" />
    </div>
    <BcDialog
      v-model="visible"
      :header="$t('premium.title')"
      :pt="{
        header: {
          class: 'premium-header'
        }
      }"
      class="bc-premium-gem-dialog"
    >
      <div class="text">
        {{ description || $t('premium.description') }}
      </div>
      <template #footer>
        <div class="footer">
          <div class="dismiss" @click="visible = false">
            {{ dismissLabel || $t('navigation.dismiss') }}
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
:global(.bc-premium-gem-dialog) {
   width: 620px;
 }
.dismiss {
  cursor: pointer;
  color: var(--text-color-disabled);
}

.gem {
  color: var(--primary-color);
  cursor: pointer;
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
}

:global(.p-dialog .p-dialog-header.premium-header .p-dialog-title) {
  color: var(--primary-color);
  font-size: var(--subtitle_font_size);
}
</style>
