export type ToastData = {
  summary: string
  detail?: string
  group?: string
}
type ToastFunction = (data: ToastData) => void

export interface BcToastProvider {
  showError: ToastFunction
  showInfo: ToastFunction
  showSuccess: ToastFunction
}
