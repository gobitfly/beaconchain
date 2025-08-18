import type { PostValidatorDashboardValidatorsRequest } from '~/types/api/validator_dashboard'

export const useValidators = () => {
  const validators = useFetchedData('validators')
  const refresh = () => {
    refreshNuxtData('validators')
  }
  const { $api } = useNuxtApp()
  const add = async (dashboardKey: string, body: PostValidatorDashboardValidatorsRequest) => {
    return await $api(`/api/bff/validator-dashboards/${dashboardKey}/validators`, {
      body,
      method: 'post',
    })
      .then(() => {
        refresh()
      })
  }
  const remove = async (dashboardKey: string, validators: number[]) => {
    // console.log('Removing validators:', validators, dashboardKey)
    // return
    return await $api(`/api/bff/validator-dashboards/${dashboardKey}/validators/bulk-deletions`, {
      body: { validators },
      method: 'post',
    })
      .then(() => {
        refresh()
      })
  }
  return {
    add,
    refresh,
    remove,
    validators,
  }
}
