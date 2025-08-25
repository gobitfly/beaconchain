import { FetchError } from 'ofetch'

// type Failure<E> = {
//   data: null,
//   error: E,
// }
// type Result<T, E = Error> = Failure<E> | Success<T>

// Types for the result object with discriminated union
// type Success<T> = {
//   data: T,
//   error: null,
// }
export async function tryCatch<T>(
  promise: Promise<T>,
) {
  try {
    const data = await promise
    return data
  }
  catch (error) {
    if (!(error instanceof FetchError)) {
      // Todo: think about throw createError
      throw createError({
        message: 'Error is of unexpected type',
        statusCode: 500,
      })
    }

    return null
  }
}
