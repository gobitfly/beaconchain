import type { GetTruncatedGuestValidatorDashboardResponse } from '~/types/api/validator_dashboard'

export default defineNuxtRouteMiddleware(async ({
  query,
}) => {
  if (isClientSide) return

  const validators = query?.validators || query?.index
  if (!validators) return
  const filteredValidators = Array.isArray(validators)
    ? validators
        .filter(id => isInt(`${id}`) || isPublicKey(`${id}`))
    : validators
        .split(',')
        .filter(id => isInt(id) || isPublicKey(id))
  if (!filteredValidators.length) return

  const { fetch } = useCustomFetch()
  try {
    /**
   * Get validators (ordered by index) eligible for the guest dashboard's free tier,
   * stopping once their total effective balance meets the tier's balance limit.
  */
    const truncatedValidators = await fetch<GetTruncatedGuestValidatorDashboardResponse>('DASHBOARD_TRUNCATED_LIST', {
      query: {
        validators: filteredValidators.join(','),
      },
    })
      .then(res => res.data)
      .then(data => data.validators)

    const isTruncated = truncatedValidators.length !== filteredValidators.length
    return navigateTo(
      {
        name: 'dashboard-id',
        params: {
          id: encodeBase64Url(truncatedValidators.join(',')),
        },
        query: {
          isTruncated: isTruncated ? null : undefined,
        },
      })
  }
  catch (error) {
    logError(`Failed to fetch truncated validator list: ${filteredValidators}\n${error}`)
    return navigateTo('/dashboard')
  }
})
