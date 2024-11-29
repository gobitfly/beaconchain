import {
  type AnySchema,
  boolean,
  // mixed,
  number,
  object,
  string,
} from 'yup'
import * as valibot from 'valibot'
import { toTypedSchema } from '@vee-validate/valibot'

export const createSchemaObject = (schema: Record<string, AnySchema>) => {
  return object({ ...schema })
}

export const validation = {
  // expose thirdparty validation here, when needed
  boolean,
  number,
  // mixed,
  numberRange: (options: { max: number, min: number }) => {
    return toTypedSchema(
      valibot.pipe(
        valibot.string(),
        valibot.transform(input => Number(input)),
        valibot.number(),
        valibot.toMinValue(options.min),
        valibot.toMaxValue(options.max),
        valibot.transform(input => input.toString()),
      ),
    )
  },
  participationRate: number().integer().max(100).positive().default(10),
  url: (message: string) => string().url(message),
}
