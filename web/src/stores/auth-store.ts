import { create } from 'zustand'

const TOKEN_KEY = 'agentbox_token'

export interface AuthUser {
  id: string
  username: string
  role: string
  exp: number
}

interface AuthState {
  auth: {
    user: AuthUser | null
    setUser: (user: AuthUser | null) => void
    accessToken: string
    setAccessToken: (accessToken: string) => void
    resetAccessToken: () => void
    reset: () => void
    isAdmin: () => boolean
    isAuthenticated: () => boolean
  }
}

function loadToken(): string {
  try {
    return localStorage.getItem(TOKEN_KEY) || ''
  } catch {
    return ''
  }
}

function parseJwtPayload(token: string): AuthUser | null {
  try {
    const payload = token.split('.')[1]
    const decoded = JSON.parse(atob(payload))
    return {
      id: decoded.user_id,
      username: decoded.username,
      role: decoded.role,
      exp: decoded.exp * 1000, // convert to ms
    }
  } catch {
    return null
  }
}

export const useAuthStore = create<AuthState>()((set, get) => {
  const initToken = loadToken()
  const initUser = initToken ? parseJwtPayload(initToken) : null

  return {
    auth: {
      user: initUser,
      setUser: (user) =>
        set((state) => ({ ...state, auth: { ...state.auth, user } })),
      accessToken: initToken,
      setAccessToken: (accessToken) =>
        set((state) => {
          localStorage.setItem(TOKEN_KEY, accessToken)
          const user = parseJwtPayload(accessToken)
          return { ...state, auth: { ...state.auth, accessToken, user } }
        }),
      resetAccessToken: () =>
        set((state) => {
          localStorage.removeItem(TOKEN_KEY)
          return { ...state, auth: { ...state.auth, accessToken: '', user: null } }
        }),
      reset: () =>
        set((state) => {
          localStorage.removeItem(TOKEN_KEY)
          return {
            ...state,
            auth: { ...state.auth, user: null, accessToken: '' },
          }
        }),
      isAdmin: () => {
        const state = get()
        return state.auth.user?.role === 'admin'
      },
      isAuthenticated: () => {
        const state = get()
        if (!state.auth.accessToken || !state.auth.user) return false
        // Check if token is expired
        if (state.auth.user.exp < Date.now()) {
          localStorage.removeItem(TOKEN_KEY)
          return false
        }
        return true
      },
    },
  }
})
