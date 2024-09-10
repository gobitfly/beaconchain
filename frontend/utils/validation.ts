import {
  type AnySchema, boolean, object, string,
} from 'yup'

export const createSchemaObject = (schema: Record<string, AnySchema>) => {
  return object({ ...schema })
}

export const validation = {
  // expose thirdparty validation here, when needed
  boolean,
  email: (message: string, { isRequiredMessage = '' } = {}) => {
    const baseValidation = string().email(message)
    if (isRequiredMessage) {
      return baseValidation.required(isRequiredMessage)
    }
    return baseValidation
  },
  // email: (message: string) => {
  //   const baseValidation = string().email(message)
  //   return baseValidation
  // },
  url: (message: string) => string().url(message),
}
