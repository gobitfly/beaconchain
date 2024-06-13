export function isParent (parent:HTMLElement | null, child:HTMLElement | null): boolean {
  if (!parent || !child) {
    return false
  }
  let node = child.parentNode

  // keep iterating unless null
  while (node !== null) {
    if (node === parent) {
      return true
    }
    node = node.parentNode
  }
  return false
}

export function isOrIsInIteractiveContainer (child:HTMLElement | null, stopSearchAtElement?: HTMLElement): boolean {
  if (!child || child === stopSearchAtElement) {
    return false
  }

  if (child.nodeName === 'INPUT') {
    return true
  }
  if (child.offsetWidth < child.scrollWidth || child.offsetHeight < child.scrollHeight) {
    return true
  }
  return isOrIsInIteractiveContainer(child.parentElement, stopSearchAtElement)
}
