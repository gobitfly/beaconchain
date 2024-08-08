export type ToastData = {
  detail?: string
  group?: string
  summary: string
}
type ToastFunction = (data: ToastData) => void

export interface BcToastProvider {
  showError: ToastFunction
  showInfo: ToastFunction
  showSuccess: ToastFunction
}
