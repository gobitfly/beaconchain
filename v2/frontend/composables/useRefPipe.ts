import { type ModelRef, type WatchStopHandle } from 'vue'

type MetaRef<T> = ModelRef<T|undefined> | Ref<T|undefined>

export function useRefPipe () {
  const stoppers: WatchStopHandle[] = []

  /** Caution: Altough the bridge between the refs is two-way (ensuring that both values will always be equal), the order of the parameters here is very important. At the moment the bridge is created, the values of the refs are not yet equal (for example one ref has data and the other ref is still undefined). To make sure that the empty value does not cross the bridge to erase your initial data, the first parameter must be the ref containing the initial data. Then, at the creation of the bridge, the initial data fills the empty ref. */
  function pipePrimitiveRefs<T> (refA: MetaRef<T>, refB: MetaRef<T>) {
    stoppers.push(watch(refA, () => {
      if (refA.value !== refB.value) {
        refB.value = refA.value
      }
    }, { immediate: true }))
    stoppers.push(watch(refB, () => {
      if (refB.value !== refA.value) {
        refA.value = refB.value
      }
    }))
  }

  /** Caution: Altough the bridge between the refs is two-way (ensuring that both values will always be equal), the order of the parameters here is very important. At the moment the bridge is created, the values of the refs are not yet equal (for example one ref has data and the other ref is still undefined). To make sure that the empty value does not cross the bridge to erase your initial data, the first parameter must be the ref containing the initial data. Then, at the creation of the bridge, the initial data fills the empty ref. */
  function pipePrimitiveRefsOfDifferentTypes<Ta, Tb> (refA: MetaRef<Ta>, refB: MetaRef<Tb>) {
    stoppers.push(watch(refA, () => {
      if (JSON.stringify(refA.value) !== JSON.stringify(refB.value)) {
        refB.value = (refA.value !== undefined) ? <Tb>(refA.value as unknown) : undefined
      }
    }, { immediate: true }))
    stoppers.push(watch(refB, () => {
      if (JSON.stringify(refB.value) !== JSON.stringify(refA.value)) {
        refA.value = (refB.value !== undefined) ? <Ta>(refB.value as unknown) : undefined
      }
    }))
  }

  /** Caution: Altough the bridge between the refs is two-way (ensuring that both values will always be equal), the order of the parameters here is very important. At the moment the bridge is created, the values of the refs are not yet equal (for example one ref has data and the other ref is still undefined). To make sure that the empty value does not cross the bridge to erase your initial data, the first parameter must be the ref containing the initial data. Then, at the creation of the bridge, the initial data fills the empty ref. */
  function pipeObjectRefs<T> (refA: MetaRef<T>, refB: MetaRef<T>) {
    stoppers.push(watch(refA, () => {
      if (JSON.stringify(refA.value) !== JSON.stringify(refB.value)) {
        refB.value = refA.value
      }
    }, { immediate: true }))
    stoppers.push(watch(refB, () => {
      if (JSON.stringify(refB.value) !== JSON.stringify(refA.value)) {
        refA.value = refB.value
      }
    }))
  }

  /** Caution: Altough the bridge between the refs is two-way (ensuring that both values will always be equal), the order of the parameters here is very important. At the moment the bridge is created, the values of the refs are not yet equal (for example one ref has data and the other ref is still undefined). To make sure that the empty value does not cross the bridge to erase your initial data, the first parameter must be the ref containing the initial data. Then, at the creation of the bridge, the initial data fills the empty ref. */
  function pipeArraysRefsOfDifferentTypes<Ta, Tb> (refA: MetaRef<Ta[]>, refB: MetaRef<Tb[]>) {
    stoppers.push(watch(refA, () => {
      const AasB = refA.value ? refA.value.map(el => <Tb>(el as unknown)) : undefined
      if (JSON.stringify(AasB) !== JSON.stringify(refB.value)) {
        refB.value = AasB
      }
    }, { immediate: true }))
    stoppers.push(watch(refB, () => {
      const BasA = refB.value ? refB.value.map(el => <Ta>(el as unknown)) : undefined
      if (JSON.stringify(BasA) !== JSON.stringify(refA.value)) {
        refA.value = BasA
      }
    }))
  }

  onUnmounted(() => {
    stoppers.forEach(stopper => stopper())
  })

  return { pipePrimitiveRefs, pipeObjectRefs, pipePrimitiveRefsOfDifferentTypes, pipeArraysRefsOfDifferentTypes }
}
