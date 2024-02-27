<script setup lang="ts">
import { Categories, ResultTypes, type SearchBarStyle } from '~/types/searchengine'
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

function redirectToRelevantPage (searched : string, type : ResultTypes, chain : ChainIDs) {
  let path : string
  let q = ''
  const networkPath = '/networks' + ChainInfo[chain].path

  switch (type) {
    case ResultTypes.Tokens :
    case ResultTypes.NFTs :
      path = '/token/' + searched
      break
    case ResultTypes.Epochs :
      path = networkPath + '/epoch/' + searched
      break
    case ResultTypes.Slots :
      path = networkPath + '/slot/' + searched
      break
    case ResultTypes.Blocks :
      path = networkPath + '/block/' + searched
      break
    case ResultTypes.BlockRoots :
    case ResultTypes.StateRoots :
    case ResultTypes.Transactions :
      path = networkPath + '/tx/' + searched
      break
    case ResultTypes.TransactionBatches :
      path = networkPath + '/transactionbatch/' + searched
      break
    case ResultTypes.StateBatches :
      path = networkPath + '/batch/' + searched
      break
    case ResultTypes.Addresses :
    case ResultTypes.Ens :
      path = '/address/' + searched
      break
    case ResultTypes.EnsOverview :
      path = '/ens/' + searched
      break
    case ResultTypes.Graffiti :
      path = networkPath + '/slots'
      q = searched
      break
    case ResultTypes.ValidatorsByIndex :
    case ResultTypes.ValidatorsByPubkey :
      path = networkPath + '/validator/' + searched
      break
    case ResultTypes.ValidatorsByDepositAddress :
    case ResultTypes.ValidatorsByDepositEnsName :
      path = networkPath + '/validators/deposits'
      q = searched
      break
    case ResultTypes.ValidatorsByWithdrawalCredential :
    case ResultTypes.ValidatorsByWithdrawalAddress :
    case ResultTypes.ValidatorsByWithdrawalEnsName :
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
    :searchable="[Categories.Protocol, Categories.Addresses, Categories.Tokens, Categories.NFTs, Categories.Validators]"
    :bar-style="searchBarStyle"
    @enter="redirectToRelevantPage"
    @select="redirectToRelevantPage"
  />
</template>
