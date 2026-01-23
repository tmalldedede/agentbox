import { Code, FileSearch, FileText, Shield, TestTube, Box } from 'lucide-react'
import type { Skill, SkillCategory } from '@/types'

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

// --- Mock Stats for Demo ---
function hashStr(s: string): number {
  let h = 0
  for (let i = 0; i < s.length; i++) {
    h = ((h << 5) - h + s.charCodeAt(i)) | 0
  }
  return Math.abs(h)
}

export interface SkillStats {
  usageCount: number
  successRate: number
  lastUsed: string
  avgDuration: number // seconds
}

export function getSkillStats(skill: Skill): SkillStats {
  const h = hashStr(skill.id + skill.name)
  const isBuiltIn = skill.is_built_in
  const base = isBuiltIn ? 80 : 10
  const usageCount = base + (h % 400)
  const successRate = 85 + (h % 15)
  const daysAgo = h % 7
  const lastUsed = daysAgo === 0 ? 'Today' : daysAgo === 1 ? 'Yesterday' : `${daysAgo}d ago`
  const avgDuration = 5 + (h % 55)
  return { usageCount, successRate, lastUsed, avgDuration }
}

export function getAggregateStats(skills: Skill[]) {
  const stats = skills.map(getSkillStats)
  const totalUsage = stats.reduce((a, s) => a + s.usageCount, 0)
  const avgSuccess = stats.length > 0
    ? Math.round(stats.reduce((a, s) => a + s.successRate, 0) / stats.length * 10) / 10
    : 0
  const avgDuration = stats.length > 0
    ? Math.round(stats.reduce((a, s) => a + s.avgDuration, 0) / stats.length * 10) / 10
    : 0
  return {
    totalSkills: skills.length,
    enabledSkills: skills.filter(s => s.is_enabled).length,
    builtInSkills: skills.filter(s => s.is_built_in).length,
    totalUsage,
    avgSuccess,
    avgDuration,
  }
}

export function getRemoteSkillStats(id: string) {
  const h = hashStr(id)
  return {
    downloads: 50 + (h % 2000),
    stars: 5 + (h % 200),
    rating: (3.5 + (h % 15) / 10).toFixed(1),
  }
}
