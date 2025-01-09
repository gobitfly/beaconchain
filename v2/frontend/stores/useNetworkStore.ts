import {
  type ChainId,
  ChainInfo,
} from '~/types/network'

export function useNetworkStore() {
  const { chainIdByDefault } = useRuntimeConfig().public
  if (!chainIdByDefault) throw createError(
    {
      statusMessage: 'NUXT_PUBLIC_CHAIN_ID_BY_DEFAULT has to be set',
    })
  const currentNetwork = computed(() => (Number(chainIdByDefault)) as ChainId)
  const networkInfo = computed(() => ChainInfo[currentNetwork.value])
  const secondsPerEpoch = computed(() => networkInfo.value.slotsPerEpoch * networkInfo.value.secondsPerSlot)
  const epochsPerDay = computed(() => (24 * 60 * 60) / secondsPerEpoch.value)

  const getTimestampFromSlot = (slot: number) =>
    networkInfo.value.timeStampSlot0 + slot * networkInfo.value.secondsPerSlot

  const getSlotFromTimestamp = (timestamp: number) =>
    Math.floor((timestamp - networkInfo.value.timeStampSlot0) / networkInfo.value.secondsPerSlot)

  const getEpochFromSlot = (slot: number) => Math.floor(slot / networkInfo.value.slotsPerEpoch)

  const getEpochFromTimestamp = (timestamp: number) => {
    const slot = getSlotFromTimestamp(timestamp)
    const epoch = getEpochFromSlot(slot)
    return epoch
  }

  const getTimestampFromEpoch = (epoch: number) => {
    return networkInfo.value.timeStampSlot0 + epoch * networkInfo.value.slotsPerEpoch * networkInfo.value.secondsPerSlot
  }

  return {
    currentNetwork,
    epochsPerDay,
    getEpochFromSlot,
    getEpochFromTimestamp,
    getSlotFromTimestamp,
    getTimestampFromEpoch,
    getTimestampFromSlot,
    networkInfo,
    secondsPerEpoch,
  }
}
