import { inject, warn } from 'vue'
import type { BcToastProvider, ToastData } from '~/types/toast'

export function useBcToast() {
  const toast = inject<BcToastProvider>('bcToast')

  const showError = (data: ToastData) => {
    if (!toast) {
      warn('bcToast provider not injected')
      return
    }
    toast.showError(data)
  }

  const showInfo = (data: ToastData) => {
    if (!toast) {
      warn('bcToast provider not injected')
      return
    }
    toast.showInfo(data)
  }

  const showSuccess = (data: ToastData) => {
    if (!toast) {
      warn('bcToast provider not injected')
      return
    }
    toast.showSuccess(data)
  }

  return { showError, showInfo, showSuccess }
}
