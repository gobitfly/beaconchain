import { provide, ref } from 'vue'
import type { DateInfo } from '~/types/date'

// useDateProvider provides a global reactive timestamp, which should be more
// performant than every component ticking their own time.
// -> a global heartbeat
export function useDateProvider() {
  const date = ref(new Date())
  const timestamp = computed(() => date.value.getTime())
  let interval: NodeJS.Timeout

  const upDate = () => {
    date.value = new Date()
  }

  onMounted(() => {
    interval = setInterval(() => upDate(), 1000)
  })

  onUnmounted(() => {
    interval && clearInterval(interval)
  })

  provide<DateInfo>('date-info', { date, timestamp })
}
