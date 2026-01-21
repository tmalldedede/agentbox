import { useNavigate } from '@tanstack/react-router'
import {
  ArrowLeft,
  Store,
  Construction,
} from 'lucide-react'

export default function SkillStoreList() {
  const navigate = useNavigate()

  return (
    <div className="min-h-screen">
      {/* Header */}
      <header className="app-header">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate({ to: '/' })} className="btn btn-ghost btn-icon">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-3">
            <Store className="w-6 h-6 text-purple-400" />
            <span className="text-lg font-bold">Skill Store</span>
          </div>
        </div>
      </header>

      <div className="p-6">
        {/* Description */}
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-foreground mb-2">Skill Store</h1>
          <p className="text-muted-foreground">
            Browse and install community skills from GitHub repositories.
          </p>
        </div>

        {/* Coming Soon Placeholder */}
        <div className="flex flex-col items-center justify-center h-64 text-center">
          <Construction className="w-16 h-16 text-muted-foreground mb-4" />
          <p className="text-muted-foreground text-lg">Coming Soon</p>
          <p className="text-muted-foreground mt-2">
            This page will allow browsing and installing skills from GitHub
          </p>
        </div>
      </div>
    </div>
  )
}
