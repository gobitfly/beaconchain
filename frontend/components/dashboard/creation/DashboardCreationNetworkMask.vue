<script lang="ts" setup>
import { IconNetwork } from '#components'
import { type ChainIDs, ChainInfo, isL1 } from '~/types/network'
import { useNetworkStore } from '~/stores/useNetworkStore'

const { availableNetworks, currentNetwork, isNetworkDisabled }
  = useNetworkStore()
const { t: $t } = useTranslation()

const emit = defineEmits<{ (e: 'back' | 'next'): void }>()
const network = defineModel<ChainIDs>('network', { required: true })
const selection = usePrimitiveRefBridge<ChainIDs, '' | `${ChainIDs}`>(
  network,
  net => `${net || ''}`,
  sel => Number(sel || 0),
)

const buttonList = shallowRef<any[]>([])

watch(currentNetwork, (id) => {
  network.value = id
})

watch(
  availableNetworks,
  () => {
    buttonList.value = [] as any[]
    availableNetworks.value.forEach((chainId) => {
      if (isL1(chainId)) {
        buttonList.value.push({
          component: IconNetwork,
          componentClass: 'dashboard-creation-button-network-icon',
          componentProps: {
            chainId,
            colored: false,
            harmonizePerceivedSize: true,
          },
          disabled: isNetworkDisabled(chainId),
          subText: isNetworkDisabled(chainId)
            ? $t('common.coming_soon')
            : ChainInfo[chainId].nameParts[1],
          text: ChainInfo[chainId].nameParts[0],
          value: String(chainId),
        })
      }
    })
  },
  { immediate: true },
)

const continueDisabled = computed(() => {
  return !selection.value
})
</script>

<template>
  <div class="mask-container">
    <div class="element-container">
      <div class="big_text">
        {{ $t("dashboard.creation.title") }}
      </div>
      <div class="subtitle_text">
        {{ $t("dashboard.creation.network.subtitle") }}
      </div>
      <BcToggleSingleBar
        v-model="selection"
        class="single-bar"
        :buttons="buttonList"
        :are-buttons-networks="true"
        layout="gaudy"
      />
      <div class="row-container">
        <Button @click="emit('back')">
          {{ $t("navigation.back") }}
        </Button>
        <Button
          :disabled="continueDisabled"
          @click="emit('next')"
        >
          {{ $t("navigation.continue") }}
        </Button>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.mask-container {
  width: 100%;
  .element-container {
    display: flex;
    flex-direction: column;
    gap: var(--padding);

    .single-bar {
      height: 100px;
    }

    .row-container {
      display: flex;
      justify-content: flex-end;
      gap: var(--padding);
    }

    button {
      width: 90px;
    }
  }
}
</style>

<style lang="scss">
.dashboard-creation-button-network-icon {
  width: 100%;
  height: 100%;
}
</style>
