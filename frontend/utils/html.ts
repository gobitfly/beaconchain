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

function isScrollable (element: HTMLElement): boolean {
  if (element.offsetWidth < element.scrollWidth || element.offsetHeight < element.scrollHeight) {
    return true
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
  if (isScrollable(child)) {
    return true
  }
  return isOrIsInIteractiveContainer(child.parentElement, stopSearchAtElement)
}

export function findAllScrollParents (child:HTMLElement | null, list?: HTMLElement[]): HTMLElement[] {
  if (!list) {
    list = []
  }
  if (!child) {
    return list
  }
  if (isScrollable(child)) {
    list.push(child)
  }
  return findAllScrollParents(child.parentElement, list)
}
