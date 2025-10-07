import pricing from './pricing.json'

export default defineEventHandler(() => {
  type ApiPlan = 'business' | 'hobbyist' | 'scale'

  type ApiPrice = {
    [key in ApiPlan]: {
      monthly_price: number,
      requests_per_second: number,
      yearly_price: number,
      yearly_price_with_discount?: number,
    }
  } & { free: {
    requests_per_second: number,
  }, }
  const priceList: ApiPrice = pricing

  return priceList
})
