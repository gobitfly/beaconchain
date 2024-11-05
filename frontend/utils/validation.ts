import {
  type AnySchema,
  boolean,
  lazy,
  mixed,
  // mixed,
  number,
  object,
  string,
} from 'yup'
import * as valibot from 'valibot'
import { toTypedSchema } from '@vee-validate/yup'

export const createSchemaObject = (schema: Record<string, AnySchema>) => {
  return object({ ...schema })
}

const appendSchemas = (
  schema: AnySchema,
  schemasToAppend: AnySchema[],
) => {
  const result = schema
  schemasToAppend.forEach((schemaToAppend) => {
    result.concat(schemaToAppend)
  })
  return result
}

export const validation = {
  // expose thirdparty validation here, when needed
  boolean,
  mixed,
  number,
  // numberRange: (options: { max: number, min: number }) => {
  //   return toTypedSchema(
  //     valibot.pipe(
  //       valibot.string(),
  //       valibot.transform(input => Number(input)),
  //       valibot.number(),
  //       valibot.toMinValue(options.min),
  //       valibot.toMaxValue(options.max),
  //       valibot.transform(input => `${input}`),
  //     ),
  //   )
  // },
  numberRange: (options: { isInteger?: boolean, max: number, min: number }) => {
    const schema = number()
      // .transform(value => !value ? 0 : Number(value))
      .min(options.min)
      .max(options.max)
    return toTypedSchema(schema)
  },
  participationRate: number().integer().max(100).positive().default(10),
  string,
  url: (message: string) => string().url(message),
}
