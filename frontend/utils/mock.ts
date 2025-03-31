import type { InternalGetLatestStateResponse } from '~/types/api/latest_state'

let mockSlot = 10000

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
