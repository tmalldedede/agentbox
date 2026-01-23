import { Code, FileSearch, FileText, Shield, TestTube, Box } from 'lucide-react'
import type { SkillCategory } from '@/types'

export const categoryOptions = [
  { label: 'Coding', value: 'coding', icon: Code },
  { label: 'Review', value: 'review', icon: FileSearch },
  { label: 'Docs', value: 'docs', icon: FileText },
  { label: 'Security', value: 'security', icon: Shield },
  { label: 'Testing', value: 'testing', icon: TestTube },
  { label: 'Other', value: 'other', icon: Box },
] as const

export const statusOptions = [
  { label: 'Enabled', value: 'enabled' },
  { label: 'Disabled', value: 'disabled' },
] as const

export const categoryColorMap: Record<SkillCategory, string> = {
  coding: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
  review: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400',
  docs: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400',
  security: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
  testing: 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-400',
  other: 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400',
}

export const categoryBgColors: Record<SkillCategory, string> = {
  coding: 'bg-blue-500/20 text-blue-500',
  review: 'bg-purple-500/20 text-purple-500',
  docs: 'bg-emerald-500/20 text-emerald-500',
  security: 'bg-red-500/20 text-red-500',
  testing: 'bg-amber-500/20 text-amber-500',
  other: 'bg-gray-500/20 text-gray-500',
}
