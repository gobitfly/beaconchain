import { ChainIDs } from '~/types/networks'

export enum Categories {
  Tokens = 'tokens',
  NFTs = 'nfts',
  Protocol = 'protocol',
  Addresses = 'addresses',
  Validators = 'validators'
}

export enum ResultTypes {
  Tokens = 'tokens',
  NFTs = 'nfts',
  Epochs = 'epochs',
  Slots = 'slots',
  Blocks = 'blocks',
  BlockRoots = 'block_roots',
  StateRoots = 'state_roots',
  Transactions = 'transactions',
  TransactionBatches = 'transaction_batches',
  StateBatches = 'state_batches',
  Addresses = 'addresses',
  Ens = 'ens_names',
  EnsOverview = 'ens_overview',
  Graffiti = 'graffiti',
  ValidatorsByIndex = 'validators_by_index',
  ValidatorsByPubkey = 'validators_by_pubkey',
  ValidatorsByDepositAddress = 'count_validators_by_deposit_address',
  ValidatorsByDepositEnsName = 'count_validators_by_deposit_ens_name',
  ValidatorsByWithdrawalCredential = 'count_validators_by_withdrawal_credential',
  ValidatorsByWithdrawalAddress = 'count_validators_by_withdrawal_address',
  ValidatorsByWithdrawalEnsName = 'count_validators_by_withdrawal_ens_name',
  ValidatorsByGraffiti = 'validators_by_graffiti',
  ValidatorsByName = 'validators_by_name'
}

interface CategoryInfoFields {
  filterLabel : string
}

export const CategoryInfo: Record<Categories, CategoryInfoFields> = {
  [Categories.Tokens]: { filterLabel: 'Tokens' },
  [Categories.NFTs]: { filterLabel: 'NFTs' },
  [Categories.Protocol]: { filterLabel: 'Protocol' },
  [Categories.Addresses]: { filterLabel: 'Addresses' },
  [Categories.Validators]: { filterLabel: 'Validators' }
}

interface TypeInfoFields {
  title: string,
  preLabels: string, // to be displayed before the ahead-result returned by the API,
  midLabels: string, // between the two values in the result...
  postLabels: string, // and after. These labels can be text, or HTML code for icons
  category: Categories,
  priority: number
}

export const TypeInfo: Record<ResultTypes, TypeInfoFields> = {
  [ResultTypes.Tokens]: {
    title: 'Tokens',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Tokens,
    priority: 3
  },
  [ResultTypes.NFTs]: {
    title: 'NFTs',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.NFTs,
    priority: 4
  },
  [ResultTypes.Epochs]: {
    title: 'Epochs',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Protocol,
    priority: 12
  },
  [ResultTypes.Slots]: {
    title: 'Slots',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Protocol,
    priority: 11
  },
  [ResultTypes.Blocks]: {
    title: 'Blocks',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Protocol,
    priority: 10
  },
  [ResultTypes.BlockRoots]: {
    title: 'Block roots',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Protocol,
    priority: 18
  },
  [ResultTypes.StateRoots]: {
    title: 'State roots',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Protocol,
    priority: 19
  },
  [ResultTypes.Transactions]: {
    title: 'Transactions',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Protocol,
    priority: 17
  },
  [ResultTypes.TransactionBatches]: {
    title: 'Transaction batches',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Protocol,
    priority: 14
  },
  [ResultTypes.StateBatches]: {
    title: 'State batches',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Protocol,
    priority: 13
  },
  [ResultTypes.Addresses]: {
    title: 'Addresses',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Addresses,
    priority: 2
  },
  [ResultTypes.Ens]: {
    title: 'ENS addresses',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Addresses,
    priority: 1
  },
  [ResultTypes.EnsOverview]: {
    title: 'Overview of an ENS domain',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Addresses,
    priority: 15
  },
  [ResultTypes.Graffiti]: {
    title: 'Graffiti',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Protocol,
    priority: 16
  },
  [ResultTypes.ValidatorsByIndex]: {
    title: 'Validators by index',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Validators,
    priority: 9
  },
  [ResultTypes.ValidatorsByPubkey]: {
    title: 'Validators by public key',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Validators,
    priority: 9
  },
  [ResultTypes.ValidatorsByDepositAddress]: {
    title: 'Validators by deposit address',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Validators,
    priority: 6
  },
  [ResultTypes.ValidatorsByDepositEnsName]: {
    title: 'Validators by ENS of the deposit address',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Validators,
    priority: 5
  },
  [ResultTypes.ValidatorsByWithdrawalCredential]: {
    title: 'Validators by withdrawal credential',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Validators,
    priority: 8
  },
  [ResultTypes.ValidatorsByWithdrawalAddress]: {
    title: 'Validators by withdrawal address',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Validators,
    priority: 8
  },
  [ResultTypes.ValidatorsByWithdrawalEnsName]: {
    title: 'Validators by ENS of the withdrawal address',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Validators,
    priority: 7
  },
  [ResultTypes.ValidatorsByGraffiti]: {
    title: 'Validators by graffiti',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Validators,
    priority: 9999
  },
  [ResultTypes.ValidatorsByName]: {
    title: 'Validators by name',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Validators,
    priority: 9999
  }
}

export interface SearchAheadSingleResult {
  chain_id: number,
  type: string,
  str_value?: string,
  num_value?: number,
  hash_value?: string
}

export interface SearchAheadResults {
  data?: SearchAheadSingleResult[],
  error?: string
}

export interface OrganizedResults {
  networks: {
    chainId: ChainIDs,
    types: {
      type: ResultTypes,
      found: {
        main: string, // data corresponding to the input, like the address, the complete ens name, graffito ...
        complement: string // optional additional information, like the number of findings matching the input ...
      }[]
    }[]
  }[]
}

// This function executes a guessing procedure, hopefully correct. It is the best doable until the API specification
// is published to settle the structure of the response in every case. It will be improved that day if it is incorrect.
// This function takes a single result element returned by the API and
// organizes/standardizes it into information ready to be displayed in the drop-down of the search bar.
// If the data from the API is empty or unexpected then the function returns '' in field `main`,
// otherwise `main` contains result data. Field `complement` is '' if the API did not give 2 informations.
export function organizeAPIinfo (apiResponseElement : SearchAheadSingleResult) : { main: string, complement: string } {
  const SearchAheadResultFields : (keyof SearchAheadSingleResult)[] = ['str_value', 'num_value', 'hash_value']
  let mainField : keyof SearchAheadSingleResult
  let complement = ''
  const emptyResult = { main: '', complement }

  switch (apiResponseElement.type as ResultTypes) {
    case ResultTypes.Tokens :
    case ResultTypes.NFTs :
    case ResultTypes.Ens :
    case ResultTypes.EnsOverview :
    case ResultTypes.Graffiti :
    case ResultTypes.ValidatorsByDepositEnsName :
    case ResultTypes.ValidatorsByWithdrawalEnsName :
    case ResultTypes.ValidatorsByGraffiti :
    case ResultTypes.ValidatorsByName :
      mainField = SearchAheadResultFields[0]
      break
    case ResultTypes.Epochs :
    case ResultTypes.Slots :
    case ResultTypes.Blocks :
    case ResultTypes.TransactionBatches :
    case ResultTypes.StateBatches :
    case ResultTypes.ValidatorsByIndex :
      mainField = SearchAheadResultFields[1]
      break
    case ResultTypes.BlockRoots :
    case ResultTypes.StateRoots :
    case ResultTypes.Transactions :
    case ResultTypes.Addresses :
    case ResultTypes.ValidatorsByPubkey :
    case ResultTypes.ValidatorsByDepositAddress :
    case ResultTypes.ValidatorsByWithdrawalCredential :
    case ResultTypes.ValidatorsByWithdrawalAddress :
      mainField = SearchAheadResultFields[2]
      break
    default:
      return emptyResult
  }
  if (!(mainField in apiResponseElement) || apiResponseElement[mainField] === undefined || String(apiResponseElement[mainField]) === '') {
    return emptyResult
  }

  // fills the optional (second) field of OrganizedResults if the API gave a second information
  for (const optField of SearchAheadResultFields) {
    if (optField !== mainField &&
        optField in apiResponseElement && apiResponseElement[optField] !== undefined && String(apiResponseElement[optField]) !== '') {
      complement = String(apiResponseElement[optField])
      break
    }
  }

  return { main: String(apiResponseElement[mainField]), complement }
}

export function getListOfCategories () : Categories[] {
  const list : Categories[] = []

  for (const cat in Categories) {
    list.push(Categories[cat as keyof typeof Categories])
  }
  return list
}

export function getListOfResultTypes () : ResultTypes[] {
  const list : ResultTypes[] = []

  for (const type in ResultTypes) {
    list.push(ResultTypes[type as keyof typeof ResultTypes])
  }
  return list
}
