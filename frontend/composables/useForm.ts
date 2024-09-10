import { useForm as useFormVeeValidate } from 'vee-validate'

type Params = Parameters< typeof useFormVeeValidate >[0]

// ^?

export function useForm(params?: Params) {
  return {
    ...useFormVeeValidate(params),
  }
}
