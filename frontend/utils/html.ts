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

export function hasClassOrParentWithClass (child:HTMLElement | null, classList: string[]): boolean {
  if (!child) {
    return false
  }

  if (classList.find((c) => {
    if (child.classList?.contains(c)) {
      console.log('we found a match', child)
      return true
    }
    return false
  })) {
    return true
  }
  return hasClassOrParentWithClass(child.parentElement, classList)
}
