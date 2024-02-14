import { ChainIDs } from '~/types/networks'

export const enum Searchable {
    Anything = '',
    Validators = 'validators',
    Accounts = 'accounts'
}

export const enum ResultTypes {
    Tokens = 'tokens',
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

interface TypeInfoFields {
  title: string,
  preLabels: string, // to be displayed before the ahead-result returned by the API,
  midLabels: string, // between the two values in the result...
  postLabels: string // and after. These labels can be text, or HTML code for icons
}

export const TypeInfo: Record<ResultTypes, TypeInfoFields> = {
  [ResultTypes.Tokens]: {
    title: 'Tokens',
    preLabels: '',
    midLabels: '',
    postLabels: ''
  },
  [ResultTypes.Epochs]: {
    title: 'Epochs',
    preLabels: '',
    midLabels: '',
    postLabels: ''
  },
  [ResultTypes.Slots]: {
    title: 'Slots',
    preLabels: '',
    midLabels: '',
    postLabels: ''
  },
  [ResultTypes.Blocks]: {
    title: 'Blocks',
    preLabels: '',
    midLabels: '',
    postLabels: ''
  },
  [ResultTypes.BlockRoots]: {
    title: 'Block roots',
    preLabels: '',
    midLabels: '',
    postLabels: ''
  },
  [ResultTypes.StateRoots]: {
    title: 'State roots',
    preLabels: '',
    midLabels: '',
    postLabels: ''
  },
  [ResultTypes.Transactions]: {
    title: 'Transactions',
    preLabels: '',
    midLabels: '',
    postLabels: ''
  },
  [ResultTypes.TransactionBatches]: {
    title: 'Transaction batches',
    preLabels: '',
    midLabels: '',
    postLabels: ''
  },
  [ResultTypes.StateBatches]: {
    title: 'State batches',
    preLabels: '',
    midLabels: '',
    postLabels: ''
  },
  [ResultTypes.Addresses]: {
    title: 'Addresses',
    preLabels: '',
    midLabels: '',
    postLabels: ''
  },
  [ResultTypes.Ens]: {
    title: 'ENS addresses',
    preLabels: '',
    midLabels: '',
    postLabels: ''
  },
  [ResultTypes.EnsOverview]: {
    title: 'Overview of an ENS domain',
    preLabels: '',
    midLabels: '',
    postLabels: ''
  },
  [ResultTypes.Graffiti]: {
    title: 'Graffiti',
    preLabels: '',
    midLabels: '',
    postLabels: ''
  },
  [ResultTypes.ValidatorsByIndex]: {
    title: 'Validators by index',
    preLabels: '',
    midLabels: '',
    postLabels: ''
  },
  [ResultTypes.ValidatorsByPubkey]: {
    title: 'Validators by public key',
    preLabels: '',
    midLabels: '',
    postLabels: ''
  },
  [ResultTypes.ValidatorsByDepositAddress]: {
    title: 'Validators by deposit address',
    preLabels: '',
    midLabels: '',
    postLabels: ''
  },
  [ResultTypes.ValidatorsByDepositEnsName]: {
    title: 'Validators by ENS of the deposit address',
    preLabels: '',
    midLabels: '',
    postLabels: ''
  },
  [ResultTypes.ValidatorsByWithdrawalCredential]: {
    title: 'Validators by withdrawal credential',
    preLabels: '',
    midLabels: '',
    postLabels: ''
  },
  [ResultTypes.ValidatorsByWithdrawalAddress]: {
    title: 'Validators by withdrawal address',
    preLabels: '',
    midLabels: '',
    postLabels: ''
  },
  [ResultTypes.ValidatorsByWithdrawalEnsName]: {
    title: 'Validators by ENS of the withdrawal address',
    preLabels: '',
    midLabels: '',
    postLabels: ''
  },
  [ResultTypes.ValidatorsByGraffiti]: {
    title: 'Validators by graffiti',
    preLabels: '',
    midLabels: '',
    postLabels: ''
  },
  [ResultTypes.ValidatorsByName]: {
    title: 'Validators by name',
    preLabels: '',
    midLabels: '',
    postLabels: ''
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

// Takes a single result element returned by the API and
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
