import { type ModelRef, type WatchStopHandle } from 'vue'

type vRef<T> = ModelRef<T|undefined> | Ref<T|undefined>
interface ConverterCallback<Tx, Ty> { (x: Tx) : Ty}

/** This composable solves 3 difficulties that arise when performing a two way binding of reactive variables:
 *  1. Different types. This is the main reason you need to bridge reactive variables: values of different types
 *     that must stay synchronized.
 *  2. Infinite loops. When one variable changes, the goal is to update the other one and you do not want that
 *     the update triggers the first one again and so on.
 *  3. At the moment the binding is created, it is typical that one variable has initial data and the other is
 *     empty. You do not want the empty one to erase the initial data when the binding starts. You want this data
 *     to update the other variable at first. */

export function useRefBridge () {
  const stoppers: WatchStopHandle[] = []

  /** Bridges two reactive variables of primitive types (string, boolean, number, enum, etc).
   *
   * Caution: Altough the bridge between the refs is two-way (ensuring that the values will always update each other), the order of the parameters here is very important. At the moment the bridge is created, the values of the refs are not yet equal (for example one ref has data and the other ref is still undefined). To make sure that the empty value does not cross the bridge to erase your initial data, the first parameter must be the ref containing the initial data. Then, at the creation of the bridge, the initial data fills the second ref.
   * @param AtoB (optional) Callback/Arrow function that converts the first value into the type of the second value. If not provided, a basic conversion is performed, which is safe only between strings and numbers.
   * @param BtoA (optional) Callback/Arrow function that converts the second value into the type of the first value. If not provided, a basic conversion is performed, which is safe only between strings and numbers. */
  function bridgePrimitiveRefs<Ta, Tb> (refA: vRef<Ta>, refB: vRef<Tb>, AtoB?: ConverterCallback<Ta, Tb>, BtoA?: ConverterCallback<Tb, Ta>) : void {
    stoppers.push(watch(refA, () => {
      const AasB = (refA.value !== undefined) ? (AtoB ? AtoB(refA.value) : <Tb>(refA.value as unknown)) : undefined
      if (AasB !== refB.value) {
        refB.value = AasB
      }
    }, { immediate: true }))
    stoppers.push(watch(refB, () => {
      const BasA = (refB.value !== undefined) ? (BtoA ? BtoA(refB.value) : <Ta>(refB.value as unknown)) : undefined
      if (BasA !== refA.value) {
        refA.value = BasA
      }
    }))
  }

  /** Bridges two reactive arrays.
   *
   * Caution: Altough the bridge between the refs is two-way (ensuring that the values will always update each other), the order of the parameters here is very important. At the moment the bridge is created, the values of the refs are not yet equal (for example one ref has data and the other ref is still undefined). To make sure that the empty value does not cross the bridge to erase your initial data, the first parameter must be the ref containing the initial data. Then, at the creation of the bridge, the initial data fills the second ref.
   * @param AtoB (optional) Callback/Arrow function that converts the elements of the first array into elements compatibles with the second array. If not provided, a basic conversion is performed, which is safe only between strings and numbers.
   * @param BtoA (optional) Callback/Arrow function that converts the elements of the second array into elements compatibles with the first array. If not provided, a basic conversion is performed, which is safe only between strings and numbers. */
  function bridgeArrayRefs<Ta, Tb> (refA: vRef<Ta[]>, refB: vRef<Tb[]>, AtoB?: ConverterCallback<Ta, Tb>, BtoA?: ConverterCallback<Tb, Ta>) : void {
    stoppers.push(watch(refA, () => {
      const AasB = refA.value ? refA.value.map(el => AtoB ? AtoB(el) : <Tb>(el as unknown)) : undefined
      if (JSON.stringify(AasB) !== JSON.stringify(refB.value)) {
        refB.value = AasB
      }
    }, { immediate: true }))
    stoppers.push(watch(refB, () => {
      const BasA = refB.value ? refB.value.map(el => BtoA ? BtoA(el) : <Ta>(el as unknown)) : undefined
      if (JSON.stringify(BasA) !== JSON.stringify(refA.value)) {
        refA.value = BasA
      }
    }))
  }

  /** Bridges two reactive objects.
   *
   * Caution: Altough the bridge between the refs is two-way (ensuring that the values will always update each other), the order of the parameters here is very important. At the moment the bridge is created, the values of the refs are not yet equal (for example one ref has data and the other ref is still undefined). To make sure that the empty value does not cross the bridge to erase your initial data, the first parameter must be the ref containing the initial data. Then, at the creation of the bridge, the initial data fills the second ref.
   * @param AtoB Callback/Arrow function that converts the first object into the type of the second object.
   * @param BtoA Callback/Arrow function that converts the second object into the type of the first object. */
  function bridgeObjectRefs<Ta, Tb> (refA: vRef<Ta>, refB: vRef<Tb>, AtoB: ConverterCallback<Ta, Tb>, BtoA: ConverterCallback<Tb, Ta>) : void {
    stoppers.push(watch(refA, () => {
      const AasB = (refA.value !== undefined) ? AtoB(refA.value) : undefined
      if (JSON.stringify(AasB) !== JSON.stringify(refB.value)) {
        refB.value = AasB
      }
    }, { immediate: true }))
    stoppers.push(watch(refB, () => {
      const BasA = (refB.value !== undefined) ? BtoA(refB.value) : undefined
      if (JSON.stringify(BasA) !== JSON.stringify(refA.value)) {
        refA.value = BasA
      }
    }))
  }

  onUnmounted(() => {
    stoppers.forEach(stopper => stopper())
  })

  return { bridgePrimitiveRefs, bridgeArrayRefs, bridgeObjectRefs }
}
