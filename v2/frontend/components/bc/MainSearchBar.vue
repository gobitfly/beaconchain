<script setup lang="ts">
import { Category, ResultType, type SearchBarStyle, type Matching } from '~/types/searchengine'
import { ChainIDs, ChainInfo } from '~/types/networks'
const props = defineProps({ location: { type: String, required: true } })

let searchBarStyle : SearchBarStyle

switch (props.location) {
  case 'header' :
    searchBarStyle = 'discreet'
    break
  case 'page' :
    searchBarStyle = 'gaudy'
    break
}

// picks a result by default when the user presses Enter instead of clicking a result in the drop-down
function pickSomethingByDefault (possibilities : Matching[]) : Matching {
  // SearchEngine.vue has sorted the possible results in `possibilities` by network and type priority (the order appearing in the drop-down).
  // Now we look for the possibility that matches the best with the user input (this is known through the field `Matching.closeness`).
  // If several possibilities with this best closeness value exist, we catch the first one (so the one having the highest priority). This
  // happens for example when the user input corresponds to both a validator index and a block number.
  let bestMatchWithHigherPriority = possibilities[0]
  for (const possibility of possibilities) {
    if (possibility.closeness < bestMatchWithHigherPriority.closeness) {
      bestMatchWithHigherPriority = possibility
    }
  }
  return bestMatchWithHigherPriority
}

function redirectToRelevantPage (searched : string, type : ResultType, chain : ChainIDs) {
  let path : string
  let q = ''
  const networkPath = '/networks' + ChainInfo[chain].path

  switch (type) {
    case ResultType.Tokens :
    case ResultType.NFTs :
      path = '/token/' + searched
      break
    case ResultType.Epochs :
      path = networkPath + '/epoch/' + searched
      break
    case ResultType.Slots :
      path = networkPath + '/slot/' + searched
      break
    case ResultType.Blocks :
      path = networkPath + '/block/' + searched
      break
    case ResultType.BlockRoots :
    case ResultType.StateRoots :
    case ResultType.Transactions :
      path = networkPath + '/tx/' + searched
      break
    case ResultType.TransactionBatches :
      path = networkPath + '/transactionbatch/' + searched
      break
    case ResultType.StateBatches :
      path = networkPath + '/batch/' + searched
      break
    case ResultType.Contracts :
    case ResultType.Accounts :
    case ResultType.EnsAddresses :
      path = '/address/' + searched
      break
    case ResultType.EnsOverview :
      path = '/ens/' + searched
      break
    case ResultType.Graffiti :
      path = networkPath + '/slots'
      q = searched
      break
    case ResultType.ValidatorsByIndex :
    case ResultType.ValidatorsByPubkey :
      path = networkPath + '/validator/' + searched
      break
    case ResultType.ValidatorsByDepositAddress :
    case ResultType.ValidatorsByDepositEnsName :
      path = networkPath + '/validators/deposits'
      q = searched
      break
    case ResultType.ValidatorsByWithdrawalCredential :
    case ResultType.ValidatorsByWithdrawalAddress :
    case ResultType.ValidatorsByWithdrawalEnsName :
      path = networkPath + '/validators/withdrawals'
      q = searched
      break
    default :
      return
  }

  if (q !== '') {
    navigateTo({ path, query: { q } })
  } else {
    navigateTo({ path })
  }
}
</script>

<template>
  <BcSearchEngine
    :searchable="[Category.Protocol, Category.Addresses, Category.Tokens, Category.NFTs, Category.Validators]"
    :bar-style="searchBarStyle"
    :pick-by-default="pickSomethingByDefault"
    @go="redirectToRelevantPage"
  />
</template>
