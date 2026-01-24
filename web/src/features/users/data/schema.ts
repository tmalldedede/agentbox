import { z } from 'zod'

const userStatusSchema = z.union([
  z.literal('active'),
  z.literal('inactive'),
])
export type UserStatus = z.infer<typeof userStatusSchema>

const userRoleSchema = z.union([
  z.literal('admin'),
  z.literal('user'),
])
export type UserRole = z.infer<typeof userRoleSchema>

const userSchema = z.object({
  id: z.string(),
  username: z.string(),
  role: userRoleSchema,
  is_active: z.boolean(),
  created_at: z.string(),
})
export type User = z.infer<typeof userSchema>

export const userListSchema = z.array(userSchema)
