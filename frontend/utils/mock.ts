import type { InternalGetLatestStateResponse } from '~/types/api/latest_state'
import type { ApiDataResponse } from '~/types/api/common'
import { isMainNet } from '~/types/network'
import { type SearchAheadAPIresponse, type ResultType, TypeInfo, Indirect } from '~/types/searchbar'
import { type InternalGetUserNotificationSettingsResponse } from '~/types/api/notifications'

const probabilityOfNoResultOrError = 0.0

export function simulateAPIresponseForTheSearchBar (body? : Record<string, any>) : SearchAheadAPIresponse {
  const searched = body?.input as string
  const searchableTypes = body?.types as ResultType[]
  const searchableNetworks = body?.networks as number[]
  const response : SearchAheadAPIresponse = {} as SearchAheadAPIresponse
  response.data = []

  if (Math.random() < probabilityOfNoResultOrError / 2) {
    return {} as SearchAheadAPIresponse
  }
  if (Math.random() < probabilityOfNoResultOrError / 2) {
    return response
  }

  const n = Math.floor(Number(searched))
  const searchedIsPositiveInteger = (n !== Infinity && n >= 0 && String(n) === searched)

  let ordinal = searched
  if (Number(searched) < 11 || Number(searched) > 13) {
    const last = searched.slice(-1)
    ordinal += (last === '1') ? 'st' : (last === '2' ? 'nd' : (last === '3' ? 'rd' : 'th'))
  } else {
    ordinal += 'th'
  }

  response.data.push(
    {
      chain_id: 1,
      type: 'tokens',
      str_value: searched + 'Coin',
      hash_value: '0x06e523CD06A0cF68DaA6D8EB5ad672B5ADad0AD4'
    },
    {
      chain_id: 1,
      type: 'accounts',
      hash_value: '0xc9f2d4D703d5B14bdb0FF261e308F88306DfF47b'
    },
    {
      chain_id: 1,
      type: 'graffiti',
      str_value: searched + ' tutta la vita'
    },
    {
      chain_id: 1,
      type: 'contracts',
      hash_value: '0xEB84C94dCBBceF74bf6CEB74Bc9bBf418939202D',
      str_value: 'C' + searched
    },
    {
      chain_id: 1,
      type: 'contracts',
      hash_value: '0x2b7290a54aD073bB3963DDEb538b630e8ff10aD7',
      str_value: ''
    }
  )
  if (searchedIsPositiveInteger) {
    response.data.push(
      {
        chain_id: 1,
        type: 'epochs',
        num_value: Number(searched)
      },
      {
        chain_id: 1,
        type: 'slots',
        num_value: Number(searched),
        hash_value: '0x910e0f2ee77c80bc506a1cefc90751b919cc612d42f17bb0acc49b546f42f0ce'
      },
      {
        chain_id: 1,
        type: 'blocks',
        num_value: Number(searched),
        hash_value: '0x910e0f2ee77c80bc506a1cefc90751b919cc612d42f17bb0acc49b546f42f0ce'
      },
      {
        chain_id: 1,
        type: 'validator_by_index',
        num_value: Number(searched),
        hash_value: '0xa525497ec3116c1310be8d73d2efd536dc0ce6bd4b0163dffddf94dad3d91d154c061b9a3bfd1b704a5ba67fc443974a'
      },
      {
        chain_id: 1,
        type: 'validators_by_deposit_ens_name',
        str_value: searched + 'kETH-hodler-club.eth',
        hash_value: '0x3bfCb296F2d28FaDE20a7E53A508F73557Ca938'
      },
      {
        chain_id: 1,
        type: 'validators_by_graffiti',
        str_value: ordinal + ' proposal!'
      }
    )
  } else {
    response.data.push(
      {
        chain_id: 1,
        type: 'tokens',
        str_value: searched,
        hash_value: '0xb794f5ea0ba39494ce839613fffba74279579268'
      },
      {
        chain_id: 1,
        type: 'ens_addresses',
        str_value: searched + '.bitfly.eth',
        hash_value: '0x3bfCb296F2d28FaDE20a7E53A508F73557Ca938'
      },
      {
        chain_id: 1,
        type: 'ens_overview',
        str_value: searched + '.bitfly.eth',
        hash_value: '0x3bfCb296F2d28FaDE20a7E53A508F73557Ca938'
      },
      {
        chain_id: 1,
        type: 'validators_by_withdrawal_ens_name',
        str_value: searched + '.bitfly.eth',
        hash_value: '0xEB84C94dCBBceF74bf6CEB74Bc9bBf418939202D'
      },
      {
        chain_id: 1,
        type: 'validator_by_public_key',
        num_value: Math.floor(Math.random() * 1000000),
        hash_value: '0x8000300c7607886b7e6f1030f833162f81b02e702ff9cea045e5a1d4a13bc7010e277f077533c7899334df2d51d65660'
      },
      {
        chain_id: 1,
        type: 'validators_by_withdrawal_address',
        hash_value: '0xc2c89C217d256b060e6b3Ae567B6b213ad9954B2'
      }
    )
  }
  response.data.push(
    {
      chain_id: 1,
      type: 'contracts',
      hash_value: '0x1c6bC968f5Be2410e98f0CB5Fad7363Fac875351',
      str_value: 'Uniswap'
    },
    {
      chain_id: 42161,
      type: 'contracts',
      hash_value: '0xF2C5B60e03cb38cA0e568b847BB8773c93E31e12',
      str_value: 'Tormato Cash'
    },
    {
      chain_id: 42161,
      type: 'transactions',
      hash_value: '0xacd47cc7a30b4273aadec96b4e5aa06d1cfa627b751f358069b9a7febdba3c30'
    },
    {
      chain_id: 8453,
      type: 'accounts',
      hash_value: '0x2b7290a54aD073bB3963DDEb538b630e8ff10aD7'
    },
    {
      chain_id: 8453,
      type: 'validators_by_deposit_address',
      hash_value: '0x2b7290a54aD073bB3963DDEb538b630e8ff10aD7'
    }
  )
  if (searchedIsPositiveInteger) {
    response.data.push(
      {
        chain_id: 8453,
        type: 'epochs',
        num_value: Number(searched)
      },
      {
        chain_id: 8453,
        type: 'slots',
        num_value: Number(searched),
        hash_value: '0xb47b779916e7b1517863ec60c372abf0dc255180ed6b47dd6f93e77f2dd6b9ca'
      },
      {
        chain_id: 8453,
        type: 'blocks',
        num_value: Number(searched),
        hash_value: '0xb47b779916e7b1517863ec60c372abf0dc255180ed6b47dd6f93e77f2dd6b9ca'
      },
      {
        chain_id: 8453,
        type: 'validator_by_index',
        num_value: Number(searched),
        hash_value: '0x99f9ec412465e15243a5996205928ef1461fd4ef6b6a0c642748c6f85de72c801751facda0c96454a8c2ad3bd19f91ee'
      },
      {
        chain_id: 100,
        type: 'epochs',
        num_value: Number(searched)
      },
      {
        chain_id: 100,
        type: 'slots',
        num_value: Number(searched),
        hash_value: '0xd13eb040661d8d8de07d154985be5f4332f57141948a9d67b87bb7a2cae29b81'
      },
      {
        chain_id: 100,
        type: 'blocks',
        num_value: Number(searched),
        hash_value: '0xd13eb040661d8d8de07d154985be5f4332f57141948a9d67b87bb7a2cae29b81'
      },
      {
        chain_id: 100,
        type: 'validator_by_index',
        num_value: Number(searched),
        hash_value: '0x85e5ac15a728a2bf0b0b4f22312dad780d4e27856e30997ee11f73d74d86682800046a86a01d134dbdf171326cd7cc54'
      },
      {
        chain_id: 100,
        type: 'validators_by_deposit_address',
        hash_value: '0x2b7290a54aD073bB3963DDEb538b630e8ff10aD7'
      },
      {
        chain_id: 100,
        type: 'validators_by_withdrawal_address',
        hash_value: '0xc2c89C217d256b060e6b3Ae567B6b213ad9954B2'
      },
      {
        chain_id: 100,
        type: 'validators_by_withdrawal_credential',
        hash_value: '0x0100000000000000000000000c6bd499ef02a44031ffe8336f59c82d81333f2a'
      },
      {
        chain_id: 100,
        type: 'contracts',
        hash_value: '0x06e523CD06A0cF68DaA6D8EB5ad672B5ADad0AD4',
        str_value: ''
      }
    )
  } else {
    response.data.push(
      {
        chain_id: 8453,
        type: 'tokens',
        str_value: searched + 'USD',
        hash_value: '0xb794f5ea0ba39494ce839613fffba74279579268'
      },
      {
        chain_id: 8453,
        type: 'tokens',
        str_value: searched + 'Plus',
        hash_value: '0x0701BF988309bf45a6771afaa6B8802Ba3E24090'
      },
      {
        chain_id: 100,
        type: 'tokens',
        str_value: searched + 'Coin',
        hash_value: '0x71C7656EC7ab88b098defB751B7401B5f6d8976F'
      },
      {
        chain_id: 100,
        type: 'contracts',
        hash_value: '0xF2C5B60e03cb38cA0e568b847BB8773c93E31e12',
        str_value: searched
      },
      {
        chain_id: 100,
        type: 'ens_addresses',
        str_value: searched + '-futureoffinance.eth',
        hash_value: '0xA9Bc41b63fCb29F2d2FaDE0a7E3A50F7357Ca938'
      },
      {
        chain_id: 100,
        type: 'ens_addresses',
        str_value: searched + '.bitfly.eth',
        hash_value: '0x3bfCb296F2d28FaDE20a7E53A508F73557CaBdF'
      },
      {
        chain_id: 100,
        type: 'ens_overview',
        str_value: searched + '-green.eth',
        hash_value: '0xA9Bc41b63fCb29F2d2FaDE0a7E3A50F7357Ca938'
      },
      {
        chain_id: 100,
        type: 'ens_overview',
        str_value: searched + '.bitfly.eth',
        hash_value: '0x3bfCb296F2d28FaDE20a7E53A508F73557CaBdF'
      },
      {
        chain_id: 100,
        type: 'validators_by_withdrawal_ens_name',
        str_value: searched + '-futureoffinance.eth',
        hash_value: '0x0701BF988309bf45a6771afaa6B8802Ba3E24090'
      },
      {
        chain_id: 100,
        type: 'validators_by_withdrawal_ens_name',
        str_value: searched + '.bitfly.eth',
        hash_value: '0xEB84C94dCBBceF74bf6CEB74Bc9bBf418939202D'
      }
    )
  }

  // keeping only the results that the API is asked for
  if (searchableTypes.length) {
    response.data = response.data.filter(singleRes => searchableTypes.includes(singleRes.type as ResultType))
  }
  if (searchableNetworks.length) {
    response.data = response.data.filter(singleRes => searchableNetworks.includes(singleRes.chain_id) || TypeInfo[singleRes.type as ResultType].belongsToAllNetworks)
  }
  // adding fake numbers of identical results where it is possible
  for (const singleRes of response.data) {
    const batchSize = 2 + Math.floor(30 * Math.random())
    switch (TypeInfo[singleRes.type as ResultType].countSource) {
      case Indirect.APInum_value: singleRes.num_value = batchSize
        break
      // add cases here in the future if new fields can hold batches or counts
    }
  }

  return response
}

let mockSlot = 10000

export function mockLatestState (..._:any): InternalGetLatestStateResponse {
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
          symbol: '$',
          rate: randomize(2996.79)
        },
        {
          code: 'EUR',
          currency: 'Euro',
          symbol: '€',
          rate: randomize(2758.45)
        }
      ]
    }
  }
}

interface ApiChainInfo {
  chain_id: number,
  name: string
}

export function simulateAPIresponseAboutNetworkList () : ApiDataResponse<ApiChainInfo[]> {
  const result = { data: [] } as ApiDataResponse<ApiChainInfo[]>
  if (isMainNet(Number(useRuntimeConfig().public.chainIdByDefault))) {
    result.data.push({ chain_id: 1, name: 'ethereum' }, { chain_id: 100, name: 'gnosis' })
    if (useRuntimeConfig().public.showInDevelopment) {
      result.data.push({ chain_id: 42161, name: 'arbitrum' }, { chain_id: 8453, name: 'base' })
    }
  } else {
    result.data.push({ chain_id: 17000, name: 'holesky' }, { chain_id: 10200, name: 'chiado' })
    if (useRuntimeConfig().public.showInDevelopment) {
      result.data.push({ chain_id: 421614, name: 'arbitrum testnet' }, { chain_id: 84532, name: 'base testnet' })
    }
  }
  return result
}

export function mockManageNotificationsGeneral (): InternalGetUserNotificationSettingsResponse {
  return {
    data: {
      general_settings: {
        do_not_disturb_timestamp: 9000,
        is_email_notifications_enabled: false,
        is_push_notifications_enabled: true,
        is_machine_offline_subscribed: true,
        machine_storage_usage_threshold: 80,
        machine_cpu_usage_threshold: 40,
        machine_memory_usage_threshold: 50,
        subscribed_clients: [],
        is_rocket_pool_new_reward_round_subscribed: true,
        rocket_pool_max_collateral_threshold: 29823,
        rocket_pool_min_collateral_threshold: 123
      },
      networks: [],
      paired_devices: [
        {
          id: 'ABC-test',
          name: 'My device',
          is_notifications_enabled: true,
          paired_timestamp: 1620000000
        },
        {
          id: 'DEF-test',
          name: 'My other device',
          is_notifications_enabled: false,
          paired_timestamp: 1700000000
        }
      ]
    }
  }
}
