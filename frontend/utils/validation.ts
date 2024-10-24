import {
  type AnySchema, boolean, number, object, string,
} from 'yup'

export const createSchemaObject = (schema: Record<string, AnySchema>) => {
  return object({ ...schema })
}

export const validation = {
  // expose thirdparty validation here, when needed
  boolean,
  number,
  url: (message: string) => string().url(message),
}
