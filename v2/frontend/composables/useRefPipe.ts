import { type WatchStopHandle } from 'vue'

export interface ConverterCallBack<Tx, Ty> { (x: Tx) : Ty}

export function useRefPipe () {
  const stoppers: WatchStopHandle[] = []

  function bindPrimitiveRefs<T> (refA: Ref<T>, refB: Ref<T>) {
    stoppers.push(watch(refA, (a: T) => {
      if (a !== refB.value) {
        refB.value = a
      }
    }, { immediate: true }))
    stoppers.push(watch(refB, (b: T) => {
      if (b !== refA.value) {
        refA.value = b
      }
    }, { immediate: true }))
  }

  function bindPrimitiveRefsOfDifferentTypes<Ta, Tb> (refA: Ref<Ta>, refB: Ref<Tb>) {
    stoppers.push(watch(refA, (a: Ta) => {
      if (JSON.stringify(a) !== JSON.stringify(refB.value)) {
        refB.value = <Tb>(a as unknown)
      }
    }, { immediate: true }))
    stoppers.push(watch(refB, (b: Tb) => {
      if (JSON.stringify(b) !== JSON.stringify(refA.value)) {
        refA.value = <Ta>(b as unknown)
      }
    }, { immediate: true }))
  }

  function bindObjectRefs<T> (refA: Ref<T>, refB: Ref<T>) {
    stoppers.push(watch(refA, (a: T) => {
      if (JSON.stringify(a) !== JSON.stringify(refB.value)) {
        refB.value = a
      }
    }, { immediate: true }))
    stoppers.push(watch(refB, (b: T) => {
      if (JSON.stringify(b) !== JSON.stringify(refA.value)) {
        refA.value = b
      }
    }, { immediate: true }))
  }

  function bindArraysRefsOfDifferentTypes<Ta, Tb> (refA: Ref<Ta[]>, refB: Ref<Tb[]>, AtoB: ConverterCallBack<Ta, Tb>, BtoA: ConverterCallBack<Tb, Ta>) {
    stoppers.push(watch(refA, (a: Ta[]) => {
      const AasB = a.map(el => AtoB(el))
      if (JSON.stringify(AasB) !== JSON.stringify(refB.value)) {
        refB.value = AasB
      }
    }, { immediate: true }))
    stoppers.push(watch(refB, (b: Tb[]) => {
      const BasA = b.map(el => BtoA(el))
      if (JSON.stringify(BasA) !== JSON.stringify(refA.value)) {
        refA.value = BasA
      }
    }, { immediate: true }))
  }

  onUnmounted(() => {
    stoppers.forEach(stopper => stopper())
  })

  return { bindPrimitiveRefs, bindObjectRefs, bindPrimitiveRefsOfDifferentTypes, bindArraysRefsOfDifferentTypes }
}
