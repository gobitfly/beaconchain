import { boolean, object, string, type AnySchema } from 'yup'

export const createSchemaObject = (schema: Record<string, AnySchema>) => {
  return object({
    ...schema,
  })
}

export const validation = {
  // exopose thirdparty validation here, when needed
  boolean,
  url: (message: string) => string().url(message),
}
