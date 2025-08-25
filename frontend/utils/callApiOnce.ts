import type { NitroFetchRequest } from 'nitropack'

type CallApiOnceOptions = {
  key: Key,
  url: NitroFetchRequest,
}
export const callApiOnce = ({
  key,
  url,
}: CallApiOnceOptions,
) => {
  const { $api } = useNuxtApp()
  callOnce(key, () => $api(url), { mode: 'navigation' })
  return callOnce(key, () => $api(url), { mode: 'navigation' })
}
