import { type ModelRef, type WatchStopHandle } from 'vue'

interface ConverterCallback<Tx, Ty> { (x: Tx) : Ty}
interface BridgedRef<T> extends Ref<T> {
   deactivateBridge: () => void,
   reactivateBridge: (updateFromOriginalRef: boolean) => void
}

/** This composable creates a two-way pipe between reactive arrays of 2 different types. The values circulate back
 *  and forth transparently from A to B and B to A, the reactivity is preserved on both ends, the types are converted
 *  when the values pass through.
 *  It solves 3 difficulties:
 *  1. Different types. This is the main reason you need to bridge reactive variables: values of different types
 *     that must stay synchronized.
 *  2. Infinite loops. When one array changes, the goal is to update the other one and you do not want that
 *     the update triggers the first one again and so on.
 *  3. At the moment the binding is created, it is typical that one array has initial data and the other is
 *     empty. You do not want the empty one to erase the initial data when the binding starts.
 * @param origRef Ref that you want to bridge with the new ref that this function will create for you.
 * @param origToCreated (optional) Callback/Arrow function that converts an element in the original array into the type of the elements in the created array. If not provided, a basic conversion is performed, which is safe only between strings and numbers.
 * @param createdToOrig (optional) Callback/Arrow function that converts an element in the created array into the type of the elements in the original array. If not provided, a basic conversion is performed, which is safe only between strings and numbers.
 * */
export function useArrayRefBridge<Torig, Tcreated> (origRef: Ref<Torig[]>|ModelRef<Torig[]>, origToCreated?: ConverterCallback<Torig, Tcreated>, createdToOrig?: ConverterCallback<Tcreated, Torig>) : BridgedRef<Tcreated[]> {
  const createdRef = ref<Tcreated[]>() as BridgedRef<Tcreated[]>
  let pauseBack = false
  let pauseForth = false
  let stopperBack: WatchStopHandle
  let stopperForth: WatchStopHandle

  function startBridge (updateFromOriginalRef: boolean) {
    pauseBack = pauseForth = false
    stopperForth = watch(origRef, () => {
      if (pauseForth) { return }
      const OasC = origRef.value ? origRef.value.map(el => origToCreated ? origToCreated(el) : stringNumberConversion<Tcreated>(el)) : undefined as unknown as Tcreated[]
      createdRef.value = OasC
      pauseBack = true
      nextTick(() => { pauseBack = false })
    }, { immediate: updateFromOriginalRef, deep: true })
    stopperBack = watch(createdRef, () => {
      if (pauseBack) { return }
      const CasO = createdRef.value ? createdRef.value.map(el => createdToOrig ? createdToOrig(el) : stringNumberConversion<Torig>(el)) : undefined as unknown as Torig[]
      origRef.value = CasO
      pauseForth = true
      nextTick(() => { pauseForth = false })
    }, { deep: true })
  }

  function reactivateBridge (updateFromOriginalRef: boolean) {
    nextTick(() => startBridge(updateFromOriginalRef))
  }

  function deactivateBridge () {
    pauseBack = true
    pauseForth = true
    stopperBack()
    stopperForth()
  }

  onUnmounted(() => {
    stopperBack?.()
    stopperForth?.()
  })

  createdRef.deactivateBridge = deactivateBridge
  createdRef.reactivateBridge = reactivateBridge
  startBridge(true)
  return createdRef
}

/** This composable creates a two-way pipe between reactive objects of 2 different structures. The values circulate back
 *  and forth transparently from A to B and B to A, the reactivity is preserved on both ends, the structures are converted
 *  when the values pass through.
 *  It solves 3 difficulties:
 *  1. Different structures. This is the main reason you need to bridge reactive objects: different objects
 *     that must stay synchronized.
 *  2. Infinite loops. When one object changes, the goal is to update the other one and you do not want that
 *     the update triggers the first one again and so on.
 *  3. At the moment the binding is created, it is typical that one object has initial data and the other is
 *     empty. You do not want the empty one to erase the initial data when the binding starts.
 * @param origRef Ref that you want to bridge with the new ref that this function will create for you.
 * @param origToCreated Callback/Arrow function that converts the object in the original ref into the structure of the object in the created ref.
 * @param createdToOrig Callback/Arrow function that converts the object in the created ref into the structure of the object in the original ref.
 * */
export function useObjectRefBridge<Torig, Tcreated> (origRef: Ref<Torig>|ModelRef<Torig>, origToCreated: ConverterCallback<Torig, Tcreated>, createdToOrig: ConverterCallback<Tcreated, Torig>) : BridgedRef<Tcreated> {
  const createdRef = ref<Tcreated>() as BridgedRef<Tcreated>
  let pauseBack = false
  let pauseForth = false
  let stopperBack: WatchStopHandle
  let stopperForth: WatchStopHandle

  function startBridge (updateFromOriginalRef: boolean) {
    pauseBack = pauseForth = false
    stopperForth = watch(origRef, () => {
      if (pauseForth) { return }
      const OasC = (origRef.value !== undefined) ? origToCreated(origRef.value) : undefined as unknown as Tcreated
      createdRef.value = OasC
      pauseBack = true
      nextTick(() => { pauseBack = false })
    }, { immediate: updateFromOriginalRef, deep: true })
    stopperBack = watch(createdRef, () => {
      if (pauseBack) { return }
      const CasO = (createdRef.value !== undefined) ? createdToOrig(createdRef.value) : undefined as unknown as Torig
      origRef.value = CasO
      pauseForth = true
      nextTick(() => { pauseForth = false })
    }, { deep: true })
  }

  function reactivateBridge (updateFromOriginalRef: boolean) {
    nextTick(() => startBridge(updateFromOriginalRef))
  }

  function deactivateBridge () {
    pauseBack = true
    pauseForth = true
    stopperBack()
    stopperForth()
  }

  onUnmounted(() => {
    stopperBack?.()
    stopperForth?.()
  })

  createdRef.deactivateBridge = deactivateBridge
  createdRef.reactivateBridge = reactivateBridge
  startBridge(true)
  return createdRef
}

/** This composable creates a two-way pipe between reactive variables of 2 different types. The values circulate back
 *  and forth transparently from A to B and B to A, the reactivity is preserved on both ends, the types are converted
 *  when the values pass through.
 *  It solves 3 difficulties:
 *  1. Different types. This is the main reason you need to bridge reactive variables: values of different types
 *     that must stay synchronized.
 *  2. Infinite loops. When one variable changes, the goal is to update the other one and you do not want that
 *     the update triggers the first one again and so on.
 *  3. At the moment the binding is created, it is typical that one variable has initial data and the other is
 *     empty. You do not want the empty one to erase the initial data when the binding starts.
 * @param origRef Ref that you want to bridge with the new ref that this function will create for you.
 * @param origToCreated (optional) Callback/Arrow function that converts the value in the original ref into the type of the value in the created ref. If not provided, a basic conversion is performed, which is safe only between strings and numbers.
 * @param createdToOrig (optional) Callback/Arrow function that converts the value in the created ref into the type of the value in the original ref. If not provided, a basic conversion is performed, which is safe only between strings and numbers.
 * */
export function usePrimitiveRefBridge<Torig, Tcreated> (origRef: Ref<Torig>|ModelRef<Torig>, origToCreated?: ConverterCallback<Torig, Tcreated>, createdToOrig?: ConverterCallback<Tcreated, Torig>) : BridgedRef<Tcreated> {
  return useObjectRefBridge<Torig, Tcreated>(origRef, origToCreated ?? stringNumberConversion, createdToOrig ?? stringNumberConversion)
}

function stringNumberConversion<TO> (from: any) : TO {
  switch (typeof from) {
    case 'number' : return String(from) as TO
    case 'string' : return Number(from) as TO
    case 'undefined' : return undefined as TO
    default : throw new TypeError('Type ' + typeof from + ' cannot be converted implicitely, please give the bridge a callback function achieving the conversion.')
  }
}
