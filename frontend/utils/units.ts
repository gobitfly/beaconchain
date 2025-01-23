export const unitFactorCrypto = {
  base: 18,
  gwei: 9,
  wei: 0,
} as const

export type CryptoUnit = keyof typeof unitFactorCrypto
