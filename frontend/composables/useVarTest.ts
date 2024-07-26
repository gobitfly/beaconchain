import { warn } from 'vue'

const mySuperVar:{count:number, keys: string[]} = {
  count: 0,
  keys: []
}

export function useVarTest (key: string) {
  mySuperVar.count++
  mySuperVar.keys.push(`${process.server ? 'server' : 'client'}.${key}`)

  warn('useVarTest', `isServer: ${process.server}`, `count: ${mySuperVar.count}`, `keys: ${mySuperVar.keys}`)

  return mySuperVar
}
