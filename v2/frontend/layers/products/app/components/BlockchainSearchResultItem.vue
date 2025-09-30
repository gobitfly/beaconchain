<script setup lang="ts">
import type { InternalPostSearchResponseWithChainId } from '~/layers/products/server/api/bff/search'

const { result } = defineProps<{
  result: InternalPostSearchResponseWithChainId['data'][number],
}>()

const {
  hoodi,
  mainnet,
} = useApiUrl()

const baseUrl = computed(() => {
  if (result.chain_id === 560048) return hoodi
  return mainnet
})
</script>

<template>
  <NuxtLink
    v-if="result.type === 'address'"
    class="flex gap-md"
    :to="`${baseUrl}/address/${result.value.address.hash}?chain_id=${result.chain_id}`"
  >
    <BaseNetworkIcon :chain-id="result.chain_id" />
    {{ result.value.address.hash }}
  </NuxtLink>

  <NuxtLink
    v-else-if="result.type === 'block'"
    class="flex gap-md"
    :to="`${baseUrl}/block/${result.value.block_number}`"
  >
    <BaseNetworkIcon :chain-id="result.chain_id" />
    {{ result.value.block_number }}
  </NuxtLink>

  <NuxtLink
    v-else-if="result.type === 'ens_name'"
    class="flex gap-md"
    :to="`${baseUrl}/ens/${result.value.ens_name}`"
  >
    <BaseNetworkIcon :chain-id="result.chain_id" />
    {{ result.value.ens_name }}
  </NuxtLink>

  <NuxtLink
    v-else-if="result.type === 'epoch'"
    class="flex gap-md"
    :to="`${baseUrl}/epoch/${result.value.epoch}`"
  >
    <BaseNetworkIcon :chain-id="result.chain_id" />
    {{ result.value.epoch }}
  </NuxtLink>

  <NuxtLink
    v-else-if="result.type === 'slot'"
    class="flex gap-md"
    :to="`${baseUrl}/slot/${result.value.slot}`"
  >
    <BaseNetworkIcon :chain-id="result.chain_id" />
    {{ result.value.slot }}
  </NuxtLink>

  <NuxtLink
    v-else-if="result.type === 'token'"
    class="flex flex-col"
    :to="`${baseUrl}/token/${result.value.address.hash}`"
  >
    <span class="flex gap-md">
      <BaseNetworkIcon :chain-id="result.chain_id" />
      <span>
        {{ `${getNetworkShortName(result.chain_id)} (${result.value.token})` }}
      </span>
    </span>
    {{ result.value.address.hash }}
  </NuxtLink>

  <NuxtLink
    v-else-if="result.type === 'transaction'"
    class="flex gap-md"
    :to="`${baseUrl}/tx/${result.value.transaction_hash}`"
  >
    <BaseNetworkIcon :chain-id="result.chain_id" />
    {{ result.value.transaction_hash }}
  </NuxtLink>

  <NuxtLink
    v-else-if="result.type === 'validator'"
    class="flex gap-md"
    :to="`${baseUrl}/validator/${result.value.index}`"
  >
    <BaseNetworkIcon :chain-id="result.chain_id" />
    {{ result.value.index || result.value.public_key }}
  </NuxtLink>

  <NuxtLink
    v-else-if="result.type === 'validators_by_deposit_address'"
    class="flex gap-md"
    :to="`${baseUrl}/validators/deposits?q=${result.value.deposit_address}&lala=jey`"
  >
    <BaseNetworkIcon :chain-id="result.chain_id" />
    {{ result.value.deposit_address }}
  </NuxtLink>

  <NuxtLink
    v-else-if="result.type === 'validators_by_graffiti'"
    class="flex gap-md"
    :to="`${baseUrl}/slots?q=${encodeURIComponent(result.value.graffiti)}&chain_id=${result.chain_id}`"
  >
    <BaseNetworkIcon :chain-id="result.chain_id" />
    {{ result.value.graffiti }}
  </NuxtLink>

  <NuxtLink
    v-else-if="result.type === 'validators_by_withdrawal_credential'"
    class="flex gap-md"
    :to="`${baseUrl}/address/0x${result.value.withdrawal_credential.substring(26)}`"
  >
    <BaseNetworkIcon :chain-id="result.chain_id" />
    {{ result.value.withdrawal_credential }}
  </NuxtLink>
</template>
