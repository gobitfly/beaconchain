import type { GetTruncatedGuestValidatorDashboardResponse } from '~/types/api/validator_dashboard'
import { isGuestDashboardKey } from '~/utils/dashboard/key'

export default defineNuxtRouteMiddleware(async (to, from) => {
  if (isClientSide) return
  if ('validators' in to.query) return
  if ('validators' in from.query) return
  if ('isTruncated' in to.query) return
  if ('isTruncated' in from.query) return

  const dashboardKey = to.params.id
  if (!dashboardKey || typeof dashboardKey !== 'string' || !isGuestDashboardKey(dashboardKey)) return
  const validators
  = decodeBase64Url(dashboardKey)
    .split(',')
    .filter(id => isInt(id) || isPublicKey(id))

  const { fetch } = useCustomFetch()
  try {
    /**
   * Get validators (ordered by index) eligible for the guest dashboard's free tier,
   * stopping once their total effective balance meets the tier's balance limit.
  */
    const truncatedValidators = await fetch<GetTruncatedGuestValidatorDashboardResponse>('DASHBOARD_TRUNCATED_LIST', {
      query: {
        validators: validators.join(','),
      },
    })
      .then(res => res.data)
      .then(data => data.validators)

    const isTruncated = truncatedValidators.length !== validators.length

    if (!isTruncated) return

    return navigateTo(
      {
        name: 'dashboard-id',
        params: {
          id: encodeBase64Url(truncatedValidators.join(',')),
        },
        query: {
          isTruncated: null,
        },
      })
  }
  catch (error) {
    logError(`Failed to fetch truncated validator list: ${validators}\n${error}`)
    return navigateTo('/dashboard')
  }
})
