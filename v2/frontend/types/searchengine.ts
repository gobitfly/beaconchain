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
  priority: number,
  belongsToAllNetworks: boolean
}

export const TypeInfo: Record<ResultTypes, TypeInfoFields> = {
  [ResultTypes.Tokens]: {
    title: 'Tokens',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Tokens,
    priority: 3,
    belongsToAllNetworks: true
  },
  [ResultTypes.NFTs]: {
    title: 'NFTs',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.NFTs,
    priority: 4,
    belongsToAllNetworks: true
  },
  [ResultTypes.Epochs]: {
    title: 'Epochs',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Protocol,
    priority: 12,
    belongsToAllNetworks: false
  },
  [ResultTypes.Slots]: {
    title: 'Slots',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Protocol,
    priority: 11,
    belongsToAllNetworks: false
  },
  [ResultTypes.Blocks]: {
    title: 'Blocks',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Protocol,
    priority: 10,
    belongsToAllNetworks: false
  },
  [ResultTypes.BlockRoots]: {
    title: 'Block roots',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Protocol,
    priority: 18,
    belongsToAllNetworks: false
  },
  [ResultTypes.StateRoots]: {
    title: 'State roots',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Protocol,
    priority: 19,
    belongsToAllNetworks: false
  },
  [ResultTypes.Transactions]: {
    title: 'Transactions',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Protocol,
    priority: 17,
    belongsToAllNetworks: false
  },
  [ResultTypes.TransactionBatches]: {
    title: 'Transaction batches',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Protocol,
    priority: 14,
    belongsToAllNetworks: false
  },
  [ResultTypes.StateBatches]: {
    title: 'State batches',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Protocol,
    priority: 13,
    belongsToAllNetworks: false
  },
  [ResultTypes.Addresses]: {
    title: 'Addresses',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Addresses,
    priority: 2,
    belongsToAllNetworks: true
  },
  [ResultTypes.Ens]: {
    title: 'ENS addresses',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Addresses,
    priority: 1,
    belongsToAllNetworks: true
  },
  [ResultTypes.EnsOverview]: {
    title: 'Overview of an ENS domain',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Addresses,
    priority: 15,
    belongsToAllNetworks: true
  },
  [ResultTypes.Graffiti]: {
    title: 'Graffiti',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Protocol,
    priority: 16,
    belongsToAllNetworks: false
  },
  [ResultTypes.ValidatorsByIndex]: {
    title: 'Validators by index',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Validators,
    priority: 9,
    belongsToAllNetworks: false
  },
  [ResultTypes.ValidatorsByPubkey]: {
    title: 'Validators by public key',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Validators,
    priority: 9,
    belongsToAllNetworks: false
  },
  [ResultTypes.ValidatorsByDepositAddress]: {
    title: 'Validators by deposit address',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Validators,
    priority: 6,
    belongsToAllNetworks: false
  },
  [ResultTypes.ValidatorsByDepositEnsName]: {
    title: 'Validators by ENS of the deposit address',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Validators,
    priority: 5,
    belongsToAllNetworks: false
  },
  [ResultTypes.ValidatorsByWithdrawalCredential]: {
    title: 'Validators by withdrawal credential',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Validators,
    priority: 8,
    belongsToAllNetworks: false
  },
  [ResultTypes.ValidatorsByWithdrawalAddress]: {
    title: 'Validators by withdrawal address',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Validators,
    priority: 8,
    belongsToAllNetworks: false
  },
  [ResultTypes.ValidatorsByWithdrawalEnsName]: {
    title: 'Validators by ENS of the withdrawal address',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Validators,
    priority: 7,
    belongsToAllNetworks: false
  },
  [ResultTypes.ValidatorsByGraffiti]: {
    title: 'Validators by graffiti',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Validators,
    priority: 9999,
    belongsToAllNetworks: false
  },
  [ResultTypes.ValidatorsByName]: {
    title: 'Validators by name',
    preLabels: '',
    midLabels: '',
    postLabels: '',
    category: Categories.Validators,
    priority: 9999,
    belongsToAllNetworks: false
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

export function getListOfResultTypes (sortByPriority : boolean) : ResultTypes[] {
  const list : ResultTypes[] = []

  for (const type in ResultTypes) {
    list.push(ResultTypes[type as keyof typeof ResultTypes])
  }
  if (sortByPriority) {
    list.sort((a, b) => { return TypeInfo[a].priority - TypeInfo[b].priority })
  }
  return list
}
