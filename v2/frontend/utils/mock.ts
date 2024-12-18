import type { InternalGetLatestStateResponse } from '~/types/api/latest_state'
import type { ApiDataResponse } from '~/types/api/common'

let mockSlot = 10000

interface ApiChainInfo {
  chain_id: number,
  name: string,
}

export function mockLatestState(..._: any): InternalGetLatestStateResponse {
  const randomize = (num: number) => {
    return num + Math.random() * num
  }
  return {
    data: {
      current_slot: ++mockSlot,
      exchange_rates: [
        {
          code: 'USD',
          currency: 'Dollar',
          rate: randomize(2996.79),
          symbol: '$',
        },
        {
          code: 'EUR',
          currency: 'Euro',
          rate: randomize(2758.45),
          symbol: '€',
        },
      ],
      finalized_epoch: Math.floor(mockSlot / 32),
    },
  }
}

export function simulateAPIresponseAboutNetworkList(): ApiDataResponse<
  ApiChainInfo[]
> {
  const result = { data: [] } as ApiDataResponse<ApiChainInfo[]>
  result.data.push(
    {
      chain_id: Number(useRuntimeConfig().public.chainIdByDefault),
      name: 'ehtereum',
    },
  )
  result.data.push(
    {
      chain_id: 100,
      name: 'gnosis',
    },
  )

  return result
}
