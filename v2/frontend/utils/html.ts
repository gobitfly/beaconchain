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
