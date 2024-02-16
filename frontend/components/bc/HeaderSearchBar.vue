<script setup lang="ts">
import { warn } from 'vue'
import { Categories, ResultTypes } from '~/types/search'
import { ChainIDs, ChainInfo } from '~/types/networks'

function redirectToRelevantPage (searched : string, type : ResultTypes, chain : ChainIDs) {
  let path : string
  let q = ''
  const networkPath = '/networks' + ChainInfo[chain].path

  warn('Search: ', searched, type, ChainInfo[chain].path)

  switch (type) {
    case ResultTypes.Tokens :
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
    id="main-bar"
    :searchable="Categories.Everything"
    width="460"
    height="34"
    @enter="redirectToRelevantPage"
    @select="redirectToRelevantPage"
  />
</template>

<style lang="scss" scoped>
#main-bar{

}
</style>
