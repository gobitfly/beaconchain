<script setup lang="ts">
import type { CryptoCurrency, Currency, FiatCurrency } from '~/types/currencies'

interface Props {
  currency?: Currency
}

const props = defineProps<Props>()

const mapped = computed(() => {
  if (isFiat(props.currency)) {
    return {
      fiat: props.currency as FiatCurrency
    }
  } else if (isCrypto(props.currency)) {
    return {
      crypto: props.currency as CryptoCurrency
    }
  }
})

</script>
<template>
  <IconFiat v-if="mapped?.fiat" :currency="mapped.fiat" />
  <IconCrypto v-else-if="mapped?.crypto" :currency="mapped.crypto" />
  <IconNetworkEthereum v-else class="monochromatic" />
</template>
