import type { DialogProps } from 'primevue/dialog'
import type { DynamicDialogInstance } from 'primevue/dynamicdialogoptions'

export function useBcDialog <T> (dialogProps?: DialogProps) {
  const { width } = useWindowSize()

  const props = ref<T>()
  const dialogRef = inject<Ref<DynamicDialogInstance>>('dialogRef')

  const position = computed(() => width.value <= 430 ? 'bottom' : 'center')

  const setHeader = (header: string, show: boolean = true) => {
    if (dialogRef?.value?.options?.props) {
      if (show) {
        dialogRef.value.options.props.showHeader = true
        dialogRef.value.options.props!.header = header
      } else {
        dialogRef.value.options.props.showHeader = false
      }
    }
  }

  onBeforeMount(() => {
    if (dialogRef?.value?.options) {
      if (!dialogRef.value.options.props) {
        dialogRef.value.options.props = { }
      }
      if (dialogProps) {
        dialogRef.value.options.props = { ...dialogRef.value.options.props, ...dialogProps }
      }
      dialogRef.value.options.props.dismissableMask = true
      dialogRef.value.options.props.modal = true
      dialogRef.value.options.props.draggable = false
      dialogRef.value.options.props.position = position.value
    }
    props.value = dialogRef?.value?.data
  })

  watch(position, (pos) => {
    if (dialogRef?.value?.options?.props) {
      dialogRef.value.options.props.position = pos
    }
  }, { immediate: true })

  return { props, dialogRef, setHeader }
}
