import { provide } from 'vue'
import type { BcToastProvider, ToastData } from '~/types/toast'

const TOAST_TIME = 3000

export function useBcToastProvider() {
  const toast = useToast()
  const { t: $t } = useTranslation()

  const { bounce, instant, temp, value } = useDebounceValue<ToastData[]>([], TOAST_TIME)

  const showError = (data: ToastData) => {
    bounce([...temp.value, data], true)
  }

  const showInfo = (data: ToastData) => {
    toast.add({ summary: data.summary, detail: data.detail, severity: 'info', life: TOAST_TIME })
  }

  const showSuccess = (data: ToastData) => {
    toast.add({ summary: data.summary, detail: data.detail, severity: 'success', life: TOAST_TIME })
  }

  watch(value, (toasts) => {
    if (toasts.length) {
      if (toasts.length === 1) {
        const hasGroup = toasts[0].group
        toast.add({
          summary: toasts[0].summary,
          detail: hasGroup ? `${toasts[0].group}: ${toasts[0].detail}` : toasts[0].detail,
          severity: 'error',
          life: TOAST_TIME,
        })
      }
      else {
        const groups: Record<string, ToastData[]> = {}
        const mapped = toasts.reduce((m, t) => {
          const group = t.group ?? ''
          if (!m[group]) {
            m[group] = []
          }
          m[group].push(t)
          return m
        }, groups)
        for (const key in mapped) {
          const list = mapped[key]
          const summary = list[0].summary
          const detail = list.length === 1 ? `${key}: ${list[0].detail}` : $t('error.multiple_times', { error: key }, list.length)

          toast.add({ summary, detail, severity: 'error', life: TOAST_TIME })
        }
      }
      instant([])
    }
  })

  provide<BcToastProvider>('bcToast', { showError, showInfo, showSuccess })
}
