import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  ArrowLeft,
  Box,
  Download,
  Trash2,
  RefreshCw,
  Loader2,
  AlertCircle,
  HardDrive,
  Clock,
  Tag,
  CheckCircle,
  XCircle,
  Terminal,
  Filter,
} from 'lucide-react'
import type { Image } from '@/types'
import { useImages, usePullImage, useRemoveImage } from '@/hooks'

export default function ImageList() {
  const navigate = useNavigate()
  const [filter, setFilter] = useState<'all' | 'agent'>('all')
  const [pullImageName, setPullImageName] = useState('')
  const [deletingId, setDeletingId] = useState<string | null>(null)

  // React Query hooks
  const { data: images = [], isLoading, isFetching, error, refetch } = useImages({
    agentOnly: filter === 'agent',
  })
  const pullImage = usePullImage()
  const removeImage = useRemoveImage()

  const handlePull = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!pullImageName.trim() || pullImage.isPending) return

    pullImage.mutate(pullImageName.trim(), {
      onSuccess: () => {
        setPullImageName('')
      },
    })
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this image?')) return

    setDeletingId(id)
    removeImage.mutate(id, {
      onSettled: () => {
        setDeletingId(null)
      },
    })
  }

  const formatSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`
    return `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB`
  }

  const formatDate = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleString()
  }

  const getImageName = (image: Image) => {
    if (image.tags && image.tags.length > 0) {
      return image.tags[0]
    }
    return image.id.slice(7, 19) // Remove sha256: prefix, show first 12 chars
  }

  const getTotalSize = () => {
    return images.reduce((acc, img) => acc + img.size, 0)
  }

  if (isLoading && images.length === 0) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
      </div>
    )
  }

  return (
    <div className="min-h-screen">
      {/* Header */}
      <header className="app-header">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate({ to: '/' })} className="btn btn-ghost btn-icon">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-3">
            <Box className="w-6 h-6 text-blue-400" />
            <div>
              <h1 className="text-xl font-semibold text-foreground">Images</h1>
              <p className="text-sm text-muted-foreground">
                {images.length} images • {formatSize(getTotalSize())} total
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <button onClick={() => refetch()} className="btn btn-secondary" disabled={isFetching}>
            <RefreshCw className={`w-4 h-4 ${isFetching ? 'animate-spin' : ''}`} />
            Refresh
          </button>
        </div>
      </header>

      <div className="p-6 space-y-6">
        {error && (
          <div className="p-4 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 flex items-center gap-3">
            <AlertCircle className="w-5 h-5 flex-shrink-0" />
            <span>{error instanceof Error ? error.message : 'Failed to fetch images'}</span>
          </div>
        )}

        {/* Pull Image Form */}
        <div className="card p-6">
          <h3 className="text-lg font-medium text-foreground mb-4 flex items-center gap-2">
            <Download className="w-5 h-5 text-emerald-400" />
            Pull Image
          </h3>
          <form onSubmit={handlePull} className="flex gap-3">
            <input
              type="text"
              value={pullImageName}
              onChange={e => setPullImageName(e.target.value)}
              placeholder="e.g. anthropic/claude-code:latest"
              className="input flex-1"
              disabled={pullImage.isPending}
            />
            <button
              type="submit"
              disabled={pullImage.isPending || !pullImageName.trim()}
              className="btn btn-primary"
            >
              {pullImage.isPending ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Pulling...
                </>
              ) : (
                <>
                  <Download className="w-4 h-4" />
                  Pull
                </>
              )}
            </button>
          </form>
          <p className="text-xs text-muted-foreground mt-2">Enter full image name with registry and tag</p>
        </div>

        {/* Filter */}
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2 text-muted-foreground">
            <Filter className="w-4 h-4" />
            <span className="text-sm">Filter:</span>
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => setFilter('all')}
              className={`px-3 py-1.5 rounded-lg text-sm transition-colors ${
                filter === 'all'
                  ? 'bg-emerald-500/20 text-emerald-400'
                  : 'bg-card text-muted-foreground hover:text-muted-foreground'
              }`}
            >
              All Images
            </button>
            <button
              onClick={() => setFilter('agent')}
              className={`px-3 py-1.5 rounded-lg text-sm transition-colors ${
                filter === 'agent'
                  ? 'bg-emerald-500/20 text-emerald-400'
                  : 'bg-card text-muted-foreground hover:text-muted-foreground'
              }`}
            >
              <Terminal className="w-3 h-3 inline mr-1" />
              Agent Images Only
            </button>
          </div>
        </div>

        {/* Image List */}
        <div className="space-y-3">
          {images.length === 0 ? (
            <div className="card p-12 flex flex-col items-center justify-center text-center">
              <Box className="w-12 h-12 text-muted-foreground mb-4 opacity-50" />
              <p className="text-muted-foreground">No images found</p>
              <p className="text-sm text-muted-foreground mt-1">Pull an image to get started</p>
            </div>
          ) : (
            images.map(image => (
              <div
                key={image.id}
                className={`card p-4 hover:border-emerald-500/30 transition-colors ${
                  image.is_agent_image ? 'border-l-4 border-l-purple-500' : ''
                }`}
              >
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-3 mb-2">
                      <Box
                        className={`w-5 h-5 ${image.is_agent_image ? 'text-purple-400' : 'text-blue-400'}`}
                      />
                      <span className="font-mono text-foreground truncate">{getImageName(image)}</span>
                      {image.is_agent_image && (
                        <span className="px-2 py-0.5 text-xs rounded-full bg-purple-500/20 text-purple-400 border border-purple-500/30">
                          Agent
                        </span>
                      )}
                      {image.in_use ? (
                        <span className="flex items-center gap-1 px-2 py-0.5 text-xs rounded-full bg-emerald-500/20 text-emerald-400">
                          <CheckCircle className="w-3 h-3" />
                          In Use
                        </span>
                      ) : (
                        <span className="flex items-center gap-1 px-2 py-0.5 text-xs rounded-full bg-gray-500/20 text-gray-400">
                          <XCircle className="w-3 h-3" />
                          Unused
                        </span>
                      )}
                    </div>

                    <div className="flex flex-wrap items-center gap-4 text-sm text-muted-foreground">
                      <span className="flex items-center gap-1">
                        <HardDrive className="w-3.5 h-3.5" />
                        {formatSize(image.size)}
                      </span>
                      <span className="flex items-center gap-1">
                        <Clock className="w-3.5 h-3.5" />
                        {formatDate(image.created)}
                      </span>
                      <span className="flex items-center gap-1 font-mono text-xs">
                        ID: {image.id.slice(7, 19)}
                      </span>
                    </div>

                    {/* Tags */}
                    {image.tags && image.tags.length > 1 && (
                      <div className="flex flex-wrap gap-2 mt-3">
                        {image.tags.slice(1).map(tag => (
                          <span
                            key={tag}
                            className="flex items-center gap-1 px-2 py-0.5 text-xs rounded bg-card text-muted-foreground"
                          >
                            <Tag className="w-3 h-3" />
                            {tag}
                          </span>
                        ))}
                      </div>
                    )}
                  </div>

                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => handleDelete(image.id)}
                      disabled={image.in_use || deletingId === image.id}
                      className="btn btn-ghost btn-icon text-red-400 hover:text-red-300 hover:bg-red-500/10 disabled:opacity-50 disabled:cursor-not-allowed"
                      title={image.in_use ? 'Cannot delete: image is in use' : 'Delete image'}
                    >
                      {deletingId === image.id ? (
                        <Loader2 className="w-4 h-4 animate-spin" />
                      ) : (
                        <Trash2 className="w-4 h-4" />
                      )}
                    </button>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  )
}
