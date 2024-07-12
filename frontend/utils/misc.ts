export const addUpValues = (obj?: Record<string, number>): number => {
  if (!obj) {
    return 0
  }
  return Object.values(obj).reduce((sum, val) => sum + val, 0)
}

/**
 * @returns Levenshtein distance between the two strings. Lower value means better similarity and vice-versa.
 */
export function levenshteinDistance (str1 : string, str2 : string) : number {
  const dist = []

  for (let i = 0; i <= str1.length; i++) {
    dist[i] = [i]
    for (let j = 1; j <= str2.length; j++) {
      if (i === 0) {
        dist[i][j] = j
      } else {
        const subst = (str1[i - 1] === str2[j - 1]) ? 0 : 1
        dist[i][j] = Math.min(dist[i - 1][j] + 1, dist[i][j - 1] + 1, dist[i - 1][j - 1] + subst)
      }
    }
  }
  return dist[str1.length][str2.length]
}

export function generateUUID () {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'
    .replace(/[xy]/g, function (c) {
      const r = Math.random() * 16 | 0
      const v = c === 'x' ? r : (r & 0x3 | 0x8)
      return v.toString(16)
    })
}

export function isInt (value?: string): boolean {
  if (!value) {
    return false
  }
  const parsed = parseInt(value)
  return !isNaN(parsed) && `${parsed}` === value
}
