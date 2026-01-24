import { useQuery } from '@tanstack/react-query'

export function useSystemHealth() {
  return useQuery({
    queryKey: ['system', 'health'],
    queryFn: async () => {
      const controller = new AbortController()
      const timeout = setTimeout(() => controller.abort(), 5000)
      try {
        const res = await fetch('/api/v1/admin/system/health', {
          signal: controller.signal,
          headers: { 'Content-Type': 'application/json' },
        })
        const data = await res.json()
        if (data.code !== 0) throw new Error(data.message)
        return data.data
      } finally {
        clearTimeout(timeout)
      }
    },
    refetchInterval: 30000,
    retry: false,
    staleTime: 25000,
  })
}

export function useDockerAvailable() {
  const { data } = useSystemHealth()
  return data?.docker?.status === 'healthy'
}
