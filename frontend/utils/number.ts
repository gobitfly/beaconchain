export const isInteger = (value: string) => !Number.isNaN(value) && value === `${parseInt(value)}`
