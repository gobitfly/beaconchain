import { ref, provide } from 'vue'
import type { DateInfo } from '~/types/date'

// useDateProvider provides a global reactive timestamp, which should be more performant then every compontent ticking their own time.
// -> a global heartbeat
export function useDateProvider () {
  const date = ref(new Date())
  const timestamp = computed(() => date.value.getTime())
  const interval = ref()

  const upDate = () => {
    date.value = new Date()
  }

  onMounted(() => {
    interval.value = setInterval(() => upDate(), 1000)
  })

  onUnmounted(() => {
    interval.value && clearInterval(interval.value)
  })

  provide<DateInfo>('date-info', { date, timestamp })
}
