import { type ModelRef, type WatchStopHandle } from 'vue'

type MetaRef<T> = ModelRef<T|undefined> | Ref<T|undefined>

export function useRefPipe () {
  const stoppers: WatchStopHandle[] = []

  function pipePrimitiveRefs<T> (refA: MetaRef<T>, refB: MetaRef<T>) {
    stoppers.push(watch(refA, (a) => {
      if (a !== refB.value) {
        refB.value = a
      }
    }, { immediate: true }))
    stoppers.push(watch(refB, (b) => {
      if (b !== refA.value) {
        refA.value = b
      }
    }, { immediate: true }))
  }

  function pipePrimitiveRefsOfDifferentTypes<Ta, Tb> (refA: MetaRef<Ta>, refB: MetaRef<Tb>) {
    stoppers.push(watch(refA, (a) => {
      if (JSON.stringify(a) !== JSON.stringify(refB.value)) {
        refB.value = <Tb>(a as unknown)
      }
    }, { immediate: true }))
    stoppers.push(watch(refB, (b) => {
      if (JSON.stringify(b) !== JSON.stringify(refA.value)) {
        refA.value = <Ta>(b as unknown)
      }
    }, { immediate: true }))
  }

  function pipeObjectRefs<T> (refA: MetaRef<T>, refB: MetaRef<T>) {
    stoppers.push(watch(refA, (a) => {
      if (JSON.stringify(a) !== JSON.stringify(refB.value)) {
        refB.value = a
      }
    }, { immediate: true }))
    stoppers.push(watch(refB, (b) => {
      if (JSON.stringify(b) !== JSON.stringify(refA.value)) {
        refA.value = b
      }
    }, { immediate: true }))
  }

  function pipeArraysRefsOfDifferentTypes<Ta, Tb> (refA: MetaRef<Ta[]>, refB: MetaRef<Tb[]>) {
    stoppers.push(watch(refA, (a) => {
      const AasB = a ? a.map(el => <Tb>(el as unknown)) : undefined
      if (JSON.stringify(AasB) !== JSON.stringify(refB.value)) {
        refB.value = AasB
      }
    }, { immediate: true }))
    stoppers.push(watch(refB, (b) => {
      const BasA = b ? b.map(el => <Ta>(el as unknown)) : undefined
      if (JSON.stringify(BasA) !== JSON.stringify(refA.value)) {
        refA.value = BasA
      }
    }, { immediate: true }))
  }

  onUnmounted(() => {
    stoppers.forEach(stopper => stopper())
  })

  return { pipePrimitiveRefs, pipeObjectRefs, pipePrimitiveRefsOfDifferentTypes, pipeArraysRefsOfDifferentTypes }
}
