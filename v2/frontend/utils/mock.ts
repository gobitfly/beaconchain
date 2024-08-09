import type { InternalGetLatestStateResponse } from '~/types/api/latest_state'
import type { ApiDataResponse } from '~/types/api/common'
import { isMainNet } from '~/types/network'
import {
  Indirect,
  type ResultType,
  type SearchAheadAPIresponse,
  TypeInfo,
} from '~/types/searchbar'
import { type InternalGetUserNotificationSettingsResponse } from '~/types/api/notifications'

const probabilityOfNoResultOrError = 0.0

export function simulateAPIresponseForTheSearchBar(
  body?: Record<string, any>,
): SearchAheadAPIresponse {
  const searched = body?.input as string
  const searchableTypes = body?.types as ResultType[]
  const searchableNetworks = body?.networks as number[]
  const response: SearchAheadAPIresponse = {} as SearchAheadAPIresponse
  response.data = []

  if (Math.random() < probabilityOfNoResultOrError / 2) {
    return {} as SearchAheadAPIresponse
  }
  if (Math.random() < probabilityOfNoResultOrError / 2) {
    return response
  }

  const n = Math.floor(Number(searched))
  const searchedIsPositiveInteger
    = n !== Infinity && n >= 0 && String(n) === searched

  let ordinal = searched
  if (Number(searched) < 11 || Number(searched) > 13) {
    const last = searched.slice(-1)
    ordinal
      += last === '1' ? 'st' : last === '2' ? 'nd' : last === '3' ? 'rd' : 'th'
  }
  else {
    ordinal += 'th'
  }

  response.data.push(
    {
      chain_id: 1,
      hash_value: '0x06e523CD06A0cF68DaA6D8EB5ad672B5ADad0AD4',
      str_value: searched + 'Coin',
      type: 'tokens',
    },
    {
      chain_id: 1,
      hash_value: '0xc9f2d4D703d5B14bdb0FF261e308F88306DfF47b',
      type: 'accounts',
    },
    {
      chain_id: 1,
      str_value: searched + ' tutta la vita',
      type: 'graffiti',
    },
    {
      chain_id: 1,
      hash_value: '0xEB84C94dCBBceF74bf6CEB74Bc9bBf418939202D',
      str_value: 'C' + searched,
      type: 'contracts',
    },
    {
      chain_id: 1,
      hash_value: '0x2b7290a54aD073bB3963DDEb538b630e8ff10aD7',
      str_value: '',
      type: 'contracts',
    },
  )
  if (searchedIsPositiveInteger) {
    response.data.push(
      {
        chain_id: 1,
        num_value: Number(searched),
        type: 'epochs',
      },
      {
        chain_id: 1,
        hash_value:
          '0x910e0f2ee77c80bc506a1cefc90751b919cc612d42f17bb0acc49b546f42f0ce',
        num_value: Number(searched),
        type: 'slots',
      },
      {
        chain_id: 1,
        hash_value:
          '0x910e0f2ee77c80bc506a1cefc90751b919cc612d42f17bb0acc49b546f42f0ce',
        num_value: Number(searched),
        type: 'blocks',
      },
      {
        chain_id: 1,
        hash_value:
          '0xa525497ec3116c1310be8d73d2efd536dc0ce6bd4b0163dffddf94dad3d91d154c061b9a3bfd1b704a5ba67fc443974a',
        num_value: Number(searched),
        type: 'validator_by_index',
      },
      {
        chain_id: 1,
        hash_value: '0x3bfCb296F2d28FaDE20a7E53A508F73557Ca938',
        str_value: searched + 'kETH-hodler-club.eth',
        type: 'validators_by_deposit_ens_name',
      },
      {
        chain_id: 1,
        str_value: ordinal + ' proposal!',
        type: 'validators_by_graffiti',
      },
    )
  }
  else {
    response.data.push(
      {
        chain_id: 1,
        hash_value: '0xb794f5ea0ba39494ce839613fffba74279579268',
        str_value: searched,
        type: 'tokens',
      },
      {
        chain_id: 1,
        hash_value: '0x3bfCb296F2d28FaDE20a7E53A508F73557Ca938',
        str_value: searched + '.bitfly.eth',
        type: 'ens_addresses',
      },
      {
        chain_id: 1,
        hash_value: '0x3bfCb296F2d28FaDE20a7E53A508F73557Ca938',
        str_value: searched + '.bitfly.eth',
        type: 'ens_overview',
      },
      {
        chain_id: 1,
        hash_value: '0xEB84C94dCBBceF74bf6CEB74Bc9bBf418939202D',
        str_value: searched + '.bitfly.eth',
        type: 'validators_by_withdrawal_ens_name',
      },
      {
        chain_id: 1,
        hash_value:
          '0x8000300c7607886b7e6f1030f833162f81b02e702ff9cea045e5a1d4a13bc7010e277f077533c7899334df2d51d65660',
        num_value: Math.floor(Math.random() * 1000000),
        type: 'validator_by_public_key',
      },
      {
        chain_id: 1,
        hash_value: '0xc2c89C217d256b060e6b3Ae567B6b213ad9954B2',
        type: 'validators_by_withdrawal_address',
      },
    )
  }
  response.data.push(
    {
      chain_id: 1,
      hash_value: '0x1c6bC968f5Be2410e98f0CB5Fad7363Fac875351',
      str_value: 'Uniswap',
      type: 'contracts',
    },
    {
      chain_id: 42161,
      hash_value: '0xF2C5B60e03cb38cA0e568b847BB8773c93E31e12',
      str_value: 'Tormato Cash',
      type: 'contracts',
    },
    {
      chain_id: 42161,
      hash_value:
        '0xacd47cc7a30b4273aadec96b4e5aa06d1cfa627b751f358069b9a7febdba3c30',
      type: 'transactions',
    },
    {
      chain_id: 8453,
      hash_value: '0x2b7290a54aD073bB3963DDEb538b630e8ff10aD7',
      type: 'accounts',
    },
    {
      chain_id: 8453,
      hash_value: '0x2b7290a54aD073bB3963DDEb538b630e8ff10aD7',
      type: 'validators_by_deposit_address',
    },
  )
  if (searchedIsPositiveInteger) {
    response.data.push(
      {
        chain_id: 8453,
        num_value: Number(searched),
        type: 'epochs',
      },
      {
        chain_id: 8453,
        hash_value:
          '0xb47b779916e7b1517863ec60c372abf0dc255180ed6b47dd6f93e77f2dd6b9ca',
        num_value: Number(searched),
        type: 'slots',
      },
      {
        chain_id: 8453,
        hash_value:
          '0xb47b779916e7b1517863ec60c372abf0dc255180ed6b47dd6f93e77f2dd6b9ca',
        num_value: Number(searched),
        type: 'blocks',
      },
      {
        chain_id: 8453,
        hash_value:
          '0x99f9ec412465e15243a5996205928ef1461fd4ef6b6a0c642748c6f85de72c801751facda0c96454a8c2ad3bd19f91ee',
        num_value: Number(searched),
        type: 'validator_by_index',
      },
      {
        chain_id: 100,
        num_value: Number(searched),
        type: 'epochs',
      },
      {
        chain_id: 100,
        hash_value:
          '0xd13eb040661d8d8de07d154985be5f4332f57141948a9d67b87bb7a2cae29b81',
        num_value: Number(searched),
        type: 'slots',
      },
      {
        chain_id: 100,
        hash_value:
          '0xd13eb040661d8d8de07d154985be5f4332f57141948a9d67b87bb7a2cae29b81',
        num_value: Number(searched),
        type: 'blocks',
      },
      {
        chain_id: 100,
        hash_value:
          '0x85e5ac15a728a2bf0b0b4f22312dad780d4e27856e30997ee11f73d74d86682800046a86a01d134dbdf171326cd7cc54',
        num_value: Number(searched),
        type: 'validator_by_index',
      },
      {
        chain_id: 100,
        hash_value: '0x2b7290a54aD073bB3963DDEb538b630e8ff10aD7',
        type: 'validators_by_deposit_address',
      },
      {
        chain_id: 100,
        hash_value: '0xc2c89C217d256b060e6b3Ae567B6b213ad9954B2',
        type: 'validators_by_withdrawal_address',
      },
      {
        chain_id: 100,
        hash_value:
          '0x0100000000000000000000000c6bd499ef02a44031ffe8336f59c82d81333f2a',
        type: 'validators_by_withdrawal_credential',
      },
      {
        chain_id: 100,
        hash_value: '0x06e523CD06A0cF68DaA6D8EB5ad672B5ADad0AD4',
        str_value: '',
        type: 'contracts',
      },
    )
  }
  else {
    response.data.push(
      {
        chain_id: 8453,
        hash_value: '0xb794f5ea0ba39494ce839613fffba74279579268',
        str_value: searched + 'USD',
        type: 'tokens',
      },
      {
        chain_id: 8453,
        hash_value: '0x0701BF988309bf45a6771afaa6B8802Ba3E24090',
        str_value: searched + 'Plus',
        type: 'tokens',
      },
      {
        chain_id: 100,
        hash_value: '0x71C7656EC7ab88b098defB751B7401B5f6d8976F',
        str_value: searched + 'Coin',
        type: 'tokens',
      },
      {
        chain_id: 100,
        hash_value: '0xF2C5B60e03cb38cA0e568b847BB8773c93E31e12',
        str_value: searched,
        type: 'contracts',
      },
      {
        chain_id: 100,
        hash_value: '0xA9Bc41b63fCb29F2d2FaDE0a7E3A50F7357Ca938',
        str_value: searched + '-futureoffinance.eth',
        type: 'ens_addresses',
      },
      {
        chain_id: 100,
        hash_value: '0x3bfCb296F2d28FaDE20a7E53A508F73557CaBdF',
        str_value: searched + '.bitfly.eth',
        type: 'ens_addresses',
      },
      {
        chain_id: 100,
        hash_value: '0xA9Bc41b63fCb29F2d2FaDE0a7E3A50F7357Ca938',
        str_value: searched + '-green.eth',
        type: 'ens_overview',
      },
      {
        chain_id: 100,
        hash_value: '0x3bfCb296F2d28FaDE20a7E53A508F73557CaBdF',
        str_value: searched + '.bitfly.eth',
        type: 'ens_overview',
      },
      {
        chain_id: 100,
        hash_value: '0x0701BF988309bf45a6771afaa6B8802Ba3E24090',
        str_value: searched + '-futureoffinance.eth',
        type: 'validators_by_withdrawal_ens_name',
      },
      {
        chain_id: 100,
        hash_value: '0xEB84C94dCBBceF74bf6CEB74Bc9bBf418939202D',
        str_value: searched + '.bitfly.eth',
        type: 'validators_by_withdrawal_ens_name',
      },
    )
  }

  // keeping only the results that the API is asked for
  if (searchableTypes.length) {
    response.data = response.data.filter(singleRes =>
      searchableTypes.includes(singleRes.type as ResultType),
    )
  }
  if (searchableNetworks.length) {
    response.data = response.data.filter(
      singleRes =>
        searchableNetworks.includes(singleRes.chain_id)
        || TypeInfo[singleRes.type as ResultType].belongsToAllNetworks,
    )
  }
  // adding fake numbers of identical results where it is possible
  for (const singleRes of response.data) {
    const batchSize = 2 + Math.floor(30 * Math.random())
    switch (TypeInfo[singleRes.type as ResultType].countSource) {
      case Indirect.APInum_value:
        singleRes.num_value = batchSize
        break
      // add cases here in the future if new fields can hold batches or counts
    }
  }

  return response
}

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
    },
  }
}

interface ApiChainInfo {
  chain_id: number
  name: string
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
  return result
}

export function mockManageNotificationsGeneral(): InternalGetUserNotificationSettingsResponse {
  return {
    data: {
      general_settings: {
        do_not_disturb_timestamp: 9000,
        is_email_notifications_enabled: false,
        is_machine_offline_subscribed: true,
        is_push_notifications_enabled: true,
        is_rocket_pool_new_reward_round_subscribed: true,
        machine_cpu_usage_threshold: 40,
        machine_memory_usage_threshold: 50,
        machine_storage_usage_threshold: 80,
        rocket_pool_max_collateral_threshold: 29823,
        rocket_pool_min_collateral_threshold: 123,
        subscribed_clients: [],
      },
      networks: [],
      paired_devices: [
        {
          id: 'ABC-test',
          is_notifications_enabled: true,
          name: 'My device',
          paired_timestamp: 1620000000,
        },
        {
          id: 'DEF-test',
          is_notifications_enabled: false,
          name: 'My other device',
          paired_timestamp: 1700000000,
        },
      ],
    },
  }
}
