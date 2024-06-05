import type { DialogProps } from 'primevue/dialog'
import type { DynamicDialogInstance } from 'primevue/dynamicdialogoptions'

export function useBcDialog <T> (dialogProps?: DialogProps) {
  const { width } = useWindowSize()
  const { setTouchableElement, onSwipe } = useSwipe()

  const props = ref<T>()
  const dialogRef = inject<Ref<DynamicDialogInstance>>('dialogRef')
  const uuid = ref(generateUUID())

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
      dialogRef.value.options.props.pt = { ...dialogRef.value.options.props.pt, root: { uuid: uuid.value } }
    }
    props.value = dialogRef?.value?.data
  })

  onMounted(() => {
    const dialog = document.querySelector(`[uuid="${uuid.value}"]`)

    if (!dialog) {
      return
    }

    setTouchableElement(dialog as HTMLElement)
    onSwipe(() => {
      onClose()
      return true
    })
  })

  watch(position, (pos) => {
    if (dialogRef?.value?.options?.props) {
      dialogRef.value.options.props.position = pos
    }
  }, { immediate: true })

  const onClose = () => {
    if (dialogRef?.value) {
      dialogRef.value.close()
    }
  }
  return { props, dialogRef, setHeader }
}
