import { type SearchAheadResult } from '~/types/searchbar'

export function simulateAPIresponseForTheSearchBar (body? : Record<string, string>) : SearchAheadResult {
  const searched = body?.input
  // const searchable = body?.searchable as ResultType[]
  const response : SearchAheadResult = {}; response.data = []

  if (Math.random() < 1 / 10) {
    // 10% of the time, we simulate an error
    return {}
  }
  // results are found 85% of the time
  if (Math.random() < 1 / 7.0) {
    return response
  }

  const n = Math.floor(Number(searched))
  const searchedIsPositiveInteger = (n !== Infinity && n >= 0 && String(n) === searched)

  response.data.push(
    {
      chain_id: 1,
      type: 'tokens',
      str_value: searched + 'Coin',
      hash_value: '0xb794f5ea0ba39494ce839613fffba74279579268'
    },
    {
      chain_id: 1,
      type: 'accounts',
      hash_value: '0x' + searched + '00bfCb29F2d2FaDE0a7E3A50F7357Ca938'
    },
    {
      chain_id: 1,
      type: 'graffiti',
      str_value: searched + ' tutta la vita'
    },
    {
      chain_id: 1,
      type: 'contracts',
      hash_value: '0x' + searched + 'a0ba39494ce839613fffba74279579260',
      str_value: 'Uniswap'
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
        type: 'validators_by_index',
        num_value: Number(searched),
        hash_value: '0xa525497ec3116c1310be8d73d2efd536dc0ce6bd4b0163dffddf94dad3d91d154c061b9a3bfd1b704a5ba67fc443974a'
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
        str_value: searched + '.bitfly.eth'
      }
    )
  }
  response.data.push(
    {
      chain_id: 1,
      type: 'contracts',
      hash_value: '0x' + searched + 'a0ba39494ce839613fffba74279579260',
      str_value: 'Uniswap'
    },
    {
      chain_id: 1,
      type: 'validators_by_withdrawal_address',
      hash_value: '0x' + searched + '00bfCb29F2d2FaDE0a7E3A5357Ca938'
    },
    {
      chain_id: 42161,
      type: 'contracts',
      hash_value: '0x' + searched + '00000000000000000000000000CAFFE',
      str_value: 'Tormato Cash'
    },
    {
      chain_id: 42161,
      type: 'transactions',
      hash_value: '0x' + searched + 'a297ab886723ecfbc2cefab2ba385792058b344fbbc1f1e0a1139b2'
    },
    {
      chain_id: 8453,
      type: 'accounts',
      hash_value: '0x' + searched + '00b29F2d2FaDE0a7E3AAaaAAa'
    },
    {
      chain_id: 8453,
      type: 'validators_by_deposit_address',
      hash_value: '0x' + searched + '00b29F2d2FaDE0a7E3AAaaAAa'
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
        type: 'validators_by_index',
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
        type: 'validators_by_index',
        num_value: Number(searched),
        hash_value: '0x85e5ac15a728a2bf0b0b4f22312dad780d4e27856e30997ee11f73d74d86682800046a86a01d134dbdf171326cd7cc54'
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
        str_value: searched + ' Coin',
        hash_value: '0x71C7656EC7ab88b098defB751B7401B5f6d8976F'
      },
      {
        chain_id: 100,
        type: 'ens_addresses',
        str_value: searched + 'hallo.eth',
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
        str_value: searched + 'hallo.eth',
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
        str_value: searched + 'hallo.eth'
      },
      {
        chain_id: 100,
        type: 'validators_by_withdrawal_ens_name',
        str_value: searched + '.bitfly.eth'
      }
    )
  }

  return response
}
