<script setup lang="ts">
import { SearchbarStyle, SearchbarPurpose, ResultType, pickHighestPriorityAmongMostRelevantMatchings } from '~/types/searchbar'
import { ChainIDs, ChainInfo } from '~/types/networks'
const props = defineProps({ location: { type: String, required: true } })

let searchbarStyle : SearchbarStyle

switch (props.location) {
  case 'page' :
    searchbarStyle = SearchbarStyle.Gaudy
    break
  case 'header' :
    searchbarStyle = SearchbarStyle.Discreet
    break
}

function redirectToRelevantPage (wanted : string, type : ResultType, chain : ChainIDs) {
  let path : string
  let q = ''
  const networkPath = '/networks' + ChainInfo[chain].path

  switch (type) {
    case ResultType.Tokens :
    case ResultType.NFTs :
      path = '/token/' + wanted
      break
    case ResultType.Epochs :
      path = networkPath + '/epoch/' + wanted
      break
    case ResultType.Slots :
      path = networkPath + '/slot/' + wanted
      break
    case ResultType.Blocks :
      path = networkPath + '/block/' + wanted
      break
    case ResultType.BlockRoots :
    case ResultType.StateRoots :
    case ResultType.Transactions :
      path = networkPath + '/tx/' + wanted
      break
    case ResultType.TransactionBatches :
      path = networkPath + '/transactionbatch/' + wanted
      break
    case ResultType.StateBatches :
      path = networkPath + '/batch/' + wanted
      break
    case ResultType.Contracts :
    case ResultType.Accounts :
    case ResultType.EnsAddresses :
      path = '/address/' + wanted
      break
    case ResultType.EnsOverview :
      path = '/ens/' + wanted
      break
    case ResultType.Graffiti :
      path = networkPath + '/slots'
      q = wanted
      break
    case ResultType.ValidatorsByIndex :
    case ResultType.ValidatorsByPubkey :
      path = networkPath + '/validator/' + wanted
      break
    case ResultType.ValidatorsByDepositAddress :
    case ResultType.ValidatorsByDepositEnsName :
      path = networkPath + '/validators/deposits'
      q = wanted
      break
    case ResultType.ValidatorsByWithdrawalCredential :
    case ResultType.ValidatorsByWithdrawalAddress :
    case ResultType.ValidatorsByWithdrawalEnsName :
      path = networkPath + '/validators/withdrawals'
      q = wanted
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
  <BcSearchbarMain
    :bar-style="searchbarStyle"
    :bar-purpose="SearchbarPurpose.General"
    :pick-by-default="pickHighestPriorityAmongMostRelevantMatchings"
    @go="redirectToRelevantPage"
  />
</template>
