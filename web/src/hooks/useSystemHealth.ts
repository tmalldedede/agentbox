import { useQuery } from '@tanstack/react-query'
import { useAuthStore } from '@/stores/auth-store'

export function useSystemHealth() {
  const { auth } = useAuthStore()

  return useQuery({
    queryKey: ['system', 'health'],
    queryFn: async () => {
      const controller = new AbortController()
      const timeout = setTimeout(() => controller.abort(), 5000)
      try {
        const headers: Record<string, string> = {
          'Content-Type': 'application/json',
        }
        if (auth.accessToken) {
          headers['Authorization'] = `Bearer ${auth.accessToken}`
        }
        const res = await fetch('/api/v1/admin/system/health', {
          signal: controller.signal,
          headers,
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
    enabled: !!auth.accessToken && auth.isAdmin(), // Only fetch for admin users
  })
}

export function useDockerAvailable() {
  const { data } = useSystemHealth()
  return data?.docker?.status === 'healthy'
}
