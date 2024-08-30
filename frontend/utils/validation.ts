import {
  type AnySchema, boolean, object, string,
} from 'yup'

export const createSchemaObject = (schema: Record<string, AnySchema>) => {
  return object({ ...schema })
}

export const validation = {
  // expose thirdparty validation here, when needed
  boolean,
  email: (message: string) => string().email(message),
  url: (message: string) => string().url(message),
}
