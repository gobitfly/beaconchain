/* eslint-disable vue/max-len -- TODO:   plz fix this */
import { type ModelRef } from 'vue'

export interface BridgeRef<T> extends Ref<T> {
  pauseBridgeFromNowOn: () => void,
  wakeupBridgeAtNextTick: () => void,
}
interface ConverterCallback<Tx, Ty> {
  (x: Tx): Ty,
}

/** This composable creates a two-way pipe between reactive variables of 2 different types. The values circulate back
 *  and forth transparently from A to B and B to A, the reactivity is preserved on both ends, the types are converted
 *  when the values pass through, infinite loops (refs triggering each other) are prevented.
 * @param origRef Ref that you want to bridge with the new ref that this function will create for you.
 * @param origToCreated (optional) Callback/Arrow function that converts the value in the original ref into the type of the value in the created ref. For strings and numbers, the function can be omitted (then the bridge converts the strings to numbers and vice-versa).
 * @param createdToOrig (optional) Callback/Arrow function that converts the value in the created ref into the type of the value in the original ref. For strings and numbers, the function can be omitted (then the bridge converts the strings to numbers and vice-versa).
 * @returns a `BridgeRef` that is essentially a regular Vue `Ref` (you use it the same way, you can assign it to regular refs and v-models, no difference), containing two methods that you can call to control the bridge if needed (`.pauseBridgeFromNowOn()` and `.wakeupBridgeAtNextTick()`).
 * */
export function usePrimitiveRefBridge<Torig, Tcreated>(
  origRef: ModelRef<Torig> | Ref<Torig>,
  origToCreated?: ConverterCallback<Torig, Tcreated>,
  createdToOrig?: ConverterCallback<Tcreated, Torig>,
): BridgeRef<Tcreated> {
  return createBridge<Torig, Tcreated, Torig, Tcreated>(
    origRef,
    origToCreated ?? stringNumberConversion<Tcreated>,
    createdToOrig ?? stringNumberConversion<Torig>,
    false,
  )
}

/** This composable creates a two-way pipe between reactive arrays of 2 different types. The values circulate back
 *  and forth transparently from A to B and B to A, the reactivity is preserved on both ends, the types are converted
 *  when the values pass through, infinite loops (refs triggering each other) are prevented.
 * @param origRef Ref that you want to bridge with the new ref that this function will create for you.
 * @param origToCreated (optional) Callback/Arrow function that converts an element in the original array into the type of the elements in the created array. For strings and numbers, the function can be omitted (then the bridge converts the strings to numbers and vice-versa).
 * @param createdToOrig (optional) Callback/Arrow function that converts an element in the created array into the type of the elements in the original array. For strings and numbers, the function can be omitted (then the bridge converts the strings to numbers and vice-versa).
 * @returns a `BridgeRef` that is essentially a regular Vue `Ref` (you use it the same way, you can assign it to regular refs and v-models, no difference), containing two methods that you can call to control the bridge if needed (`.pauseBridgeFromNowOn()` and `.wakeupBridgeAtNextTick()`).
 * */
export function useArrayRefBridge<Torig, Tcreated>(
  origRef: ModelRef<Torig[]> | Ref<Torig[]>,
  origToCreated?: ConverterCallback<Torig, Tcreated>,
  createdToOrig?: ConverterCallback<Tcreated, Torig>,
): BridgeRef<Tcreated[]> {
  return createBridge<Torig, Tcreated, Torig[], Tcreated[]>(
    origRef,
    origToCreated ?? stringNumberConversion<Tcreated>,
    createdToOrig ?? stringNumberConversion<Torig>,
    true,
  )
}

/** This composable creates a two-way pipe between reactive objects of 2 different structures. The values circulate back
 *  and forth transparently from A to B and B to A, the reactivity is preserved on both ends, the structures are converted
 *  when the values pass through, infinite loops (refs triggering each other) are prevented.
 * @param origRef Ref that you want to bridge with the new ref that this function will create for you.
 * @param origToCreated Callback/Arrow function that converts the object in the original ref into the structure of the object in the created ref.
 * @param createdToOrig Callback/Arrow function that converts the object in the created ref into the structure of the object in the original ref.
 * @returns a `BridgeRef` that is essentially a regular Vue `Ref` (you use it the same way, you can assign it to regular refs and v-models, no difference), containing two methods that you can call to control the bridge if needed (`.pauseBridgeFromNowOn()` and `.wakeupBridgeAtNextTick()`).
 * */
export function useObjectRefBridge<Torig, Tcreated>(
  origRef: ModelRef<Torig> | Ref<Torig>,
  origToCreated: ConverterCallback<Torig, Tcreated>,
  createdToOrig: ConverterCallback<Tcreated, Torig>,
): BridgeRef<Tcreated> {
  return createBridge<Torig, Tcreated, Torig, Tcreated>(
    origRef,
    origToCreated,
    createdToOrig,
    false,
  )
}

function createBridge<Torig, Tcreated, TorigWhole, TcreatedWhole>(
  origRef: ModelRef<TorigWhole> | Ref<TorigWhole>,
  origToCreated: ConverterCallback<Torig, Tcreated>,
  createdToOrig: ConverterCallback<Tcreated, Torig>,
  bothEndsAreArrays: boolean,
): BridgeRef<TcreatedWhole> {
  const createdRef = ref<TcreatedWhole>() as BridgeRef<TcreatedWhole>
  let pauseBack = false
  let pauseForth = false

  createdRef.wakeupBridgeAtNextTick = () => {
    nextTick(() => {
      pauseBack = false
      pauseForth = false
    })
  }

  createdRef.pauseBridgeFromNowOn = () => {
    pauseBack = true
    pauseForth = true
  }

  watch(
    origRef,
    () => {
      if (pauseForth) {
        return
      }
      const OasC = (
        origRef.value !== undefined
          ? bothEndsAreArrays
            ? (origRef.value as Torig[]).map(el => origToCreated(el))
            : origToCreated(origRef.value as Torig)
          : undefined
      ) as TcreatedWhole
      createdRef.value = OasC
      pauseBack = true
      nextTick(() => {
        pauseBack = false
      })
    },
    {
      deep: true,
      immediate: true,
    },
  )

  watch(
    createdRef,
    () => {
      if (pauseBack) {
        return
      }
      const CasO = (
        createdRef.value !== undefined
          ? bothEndsAreArrays
            ? (createdRef.value as Tcreated[]).map(el => createdToOrig(el))
            : createdToOrig(createdRef.value as Tcreated)
          : undefined
      ) as TorigWhole
      origRef.value = CasO
      pauseForth = true
      nextTick(() => {
        pauseForth = false
      })
    },
    { deep: true },
  )

  return createdRef
}

function stringNumberConversion<TO>(from: any): TO {
  switch (typeof from) {
    case 'number':
      return String(from) as TO
    case 'string':
      return Number(from) as TO
    case 'undefined':
      return undefined as TO
    default:
      throw new TypeError(
        'Type '
        + typeof from
        + ' cannot be converted implicitely, please give the bridge a callback function achieving the conversion.',
      )
  }
}
