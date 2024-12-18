import type { InternalGetLatestStateResponse } from '~/types/api/latest_state'
import type { ApiDataResponse } from '~/types/api/common'
import { isMainNet } from '~/types/network'

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
  if (isMainNet(Number(useRuntimeConfig().public.chainIdByDefault))) {
    result.data.push(
      {
        chain_id: 1,
        name: 'ethereum',
      },
      {
        chain_id: 100,
        name: 'gnosis',
      },
    )
    if (useRuntimeConfig().public.showInDevelopment) {
      result.data.push(
        {
          chain_id: 42161,
          name: 'arbitrum',
        },
        {
          chain_id: 8453,
          name: 'base',
        },
      )
    }
  }
  else {
    result.data.push(
      {
        chain_id: 17000,
        name: 'holesky',
      },
      {
        chain_id: 10200,
        name: 'chiado',
      },
    )
    if (useRuntimeConfig().public.showInDevelopment) {
      result.data.push(
        {
          chain_id: 421614,
          name: 'arbitrum testnet',
        },
        {
          chain_id: 84532,
          name: 'base testnet',
        },
      )
    }
  }
  result.data.push(
    {
      chain_id: 11155111,
      name: 'sepolia',
    },
  )
  return result
}
