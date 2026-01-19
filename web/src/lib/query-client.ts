import { QueryClient } from '@tanstack/react-query'

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60, // 数据在 1 分钟内被认为是新鲜的
      gcTime: 1000 * 60 * 5, // 缓存保留 5 分钟 (gcTime replaces cacheTime in v5)
      refetchOnWindowFocus: false, // 窗口聚焦时不自动重新获取
      retry: 1, // 失败后重试 1 次
      retryDelay: attemptIndex => Math.min(1000 * 2 ** attemptIndex, 30000),
    },
    mutations: {
      retry: 0, // mutation 失败不重试
    },
  },
})
