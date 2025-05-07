import {
  type ChainId,
  ChainInfo,
} from '~/types/network'

export function useNetwork() {
  const { chainIdByDefault } = useRuntimeConfig().public
  if (!chainIdByDefault) throw createError(
    {
      statusMessage: 'NUXT_PUBLIC_CHAIN_ID_BY_DEFAULT has to be set',
    })
  const currentNetwork = computed(() => (Number(chainIdByDefault)) as ChainId)
  const networkInfo = computed(() => ChainInfo[currentNetwork.value])
  const {
    clCurrency,
    displayCurrencyDefault,
    elCurrency,
    hasRocketPool,
    secondsPerSlot,
    slotsPerEpoch,
    timeStampSlot0,
  } = networkInfo.value

  const getNetworkName = (chainId: ChainId) => ChainInfo[chainId].name

  const secondsPerEpoch = computed(() => slotsPerEpoch * secondsPerSlot)
  const epochsPerDay = computed(() => (24 * 60 * 60) / secondsPerEpoch.value)

  const getTimestampFromSlot = (slot: number) =>
    timeStampSlot0 + slot * secondsPerSlot

  const getSlotFromTimestamp = (timestamp: number) =>
    Math.floor((timestamp - timeStampSlot0) / secondsPerSlot)

  const getEpochFromSlot = (slot: number) => Math.floor(slot / slotsPerEpoch)

  const getEpochFromTimestamp = (timestamp: number) => {
    const slot = getSlotFromTimestamp(timestamp)
    const epoch = getEpochFromSlot(slot)
    return epoch
  }

  /**
   *
   * @returns timestamp in seconds (backend also uses seconds instead of milliseconds like in js)
   */
  const getTimestampFromEpoch = (epoch: number) => {
    return timeStampSlot0 + epoch * slotsPerEpoch * secondsPerSlot
  }

  const numberOfEpochsTheNetworkIsConsideredToBeFinalized = 3
  const secondsUntilNetworkFinality = computed(
    () => secondsPerSlot * slotsPerEpoch * numberOfEpochsTheNetworkIsConsideredToBeFinalized,
  )

  return {
    clCurrency,
    currentNetwork,
    displayCurrencyDefault,
    elCurrency,
    epochsPerDay,
    getEpochFromSlot,
    getEpochFromTimestamp,
    getNetworkName,
    getSlotFromTimestamp,
    getTimestampFromEpoch,
    getTimestampFromSlot,
    hasRocketPool,
    networkInfo,
    secondsPerEpoch,
    secondsPerSlot,
    secondsUntilNetworkFinality,
  }
}
