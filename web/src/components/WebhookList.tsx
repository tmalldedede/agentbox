import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  Plus,
  RefreshCw,
  AlertCircle,
  Loader2,
  Webhook,
  Trash2,
  Power,
  PowerOff,
  Edit2,
  Copy,
  CheckCircle2,
  Bell,
  Link,
  Shield,
  Calendar,
} from 'lucide-react'
import type { Webhook as WebhookType } from '../types'
import { useWebhooks, useCreateWebhook, useUpdateWebhook, useDeleteWebhook } from '../hooks'
import { useLanguage } from '../contexts/LanguageContext'

// 格式化时间
function formatTime(dateStr?: string): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// Webhook 卡片组件
function WebhookCard({
  webhook,
  onEdit,
  onDelete,
  onToggle,
}: {
  webhook: WebhookType
  onEdit: () => void
  onDelete: () => void
  onToggle: () => void
}) {
  const { t } = useLanguage()
  const [copied, setCopied] = useState(false)

  const copyUrl = () => {
    navigator.clipboard.writeText(webhook.url)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div
      className={`card p-4 group transition-colors ${
        webhook.is_active
          ? 'hover:border-emerald-500/50'
          : 'opacity-60 hover:border-gray-500/50'
      }`}
    >
      <div className="flex items-start gap-4">
        {/* Icon */}
        <div
          className={`w-12 h-12 rounded-xl flex items-center justify-center ${
            webhook.is_active ? 'bg-emerald-500/20 text-emerald-400' : 'bg-gray-500/20 text-gray-400'
          }`}
        >
          <Webhook className="w-5 h-5" />
        </div>

        {/* Info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-primary truncate">{webhook.id}</span>
            {!webhook.is_active && (
              <span className="text-xs px-2 py-0.5 rounded bg-gray-500/20 text-gray-400">
                {t('disabled')}
              </span>
            )}
            {webhook.secret && (
              <span className="text-xs px-2 py-0.5 rounded bg-amber-500/20 text-amber-400">
                <Shield className="w-3 h-3 inline mr-1" />
                {t('signed')}
              </span>
            )}
          </div>

          {/* URL */}
          <div className="flex items-center gap-2 mt-1">
            <Link className="w-3 h-3 text-muted flex-shrink-0" />
            <code className="text-sm text-emerald-400 font-mono truncate flex-1">{webhook.url}</code>
            <button onClick={copyUrl} className="btn btn-ghost btn-icon flex-shrink-0" title={t('copyURL')}>
              {copied ? (
                <CheckCircle2 className="w-3 h-3 text-green-400" />
              ) : (
                <Copy className="w-3 h-3" />
              )}
            </button>
          </div>

          {/* Events */}
          <div className="flex items-center gap-2 mt-3 flex-wrap">
            {webhook.events.map(event => (
              <span
                key={event}
                className="text-xs px-2 py-0.5 rounded bg-blue-500/20 text-blue-400 border border-blue-500/30"
              >
                {event}
              </span>
            ))}
          </div>

          {/* Meta */}
          <div className="flex items-center gap-4 mt-3 text-xs text-muted">
            <div className="flex items-center gap-1">
              <Calendar className="w-3 h-3" />
              <span>{t('created')}: {formatTime(webhook.created_at)}</span>
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
          {/* Toggle Button */}
          <button
            onClick={onToggle}
            className="btn btn-ghost btn-icon"
            title={webhook.is_active ? t('disable') : t('enable')}
          >
            {webhook.is_active ? (
              <Power className="w-4 h-4 text-emerald-400" />
            ) : (
              <PowerOff className="w-4 h-4 text-gray-400" />
            )}
          </button>

          {/* Edit Button */}
          <button onClick={onEdit} className="btn btn-ghost btn-icon" title={t('edit')}>
            <Edit2 className="w-4 h-4" />
          </button>

          {/* Delete Button */}
          <button onClick={onDelete} className="btn btn-ghost btn-icon text-red-400" title={t('delete')}>
            <Trash2 className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  )
}

// 创建/编辑 Webhook 模态框
function WebhookModal({
  isOpen,
  onClose,
  onSave,
  webhook,
  isSaving,
}: {
  isOpen: boolean
  onClose: () => void
  onSave: (data: { url: string; secret?: string; events: string[] }) => void
  webhook?: WebhookType | null
  isSaving: boolean
}) {
  const { t } = useLanguage()
  const [url, setUrl] = useState('')
  const [secret, setSecret] = useState('')
  const [selectedEvents, setSelectedEvents] = useState<string[]>([])

  useEffect(() => {
    if (webhook) {
      setUrl(webhook.url)
      setSecret('') // Secret is not returned from API
      setSelectedEvents(webhook.events)
    } else {
      setUrl('')
      setSecret('')
      setSelectedEvents([])
    }
  }, [webhook, isOpen])

  if (!isOpen) return null

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (url.trim() && selectedEvents.length > 0) {
      onSave({
        url: url.trim(),
        secret: secret.trim() || undefined,
        events: selectedEvents,
      })
    }
  }

  const toggleEvent = (eventId: string) => {
    setSelectedEvents(prev =>
      prev.includes(eventId) ? prev.filter(e => e !== eventId) : [...prev, eventId]
    )
  }

  const eventTypes = [
    { id: 'task.created', label: t('taskCreated'), description: t('taskCreatedDesc') },
    { id: 'task.completed', label: t('taskCompleted'), description: t('taskCompletedDesc') },
    { id: 'task.failed', label: t('taskFailed'), description: t('taskFailedDesc') },
    { id: 'session.started', label: t('sessionStarted'), description: t('sessionStartedDesc') },
    { id: 'session.stopped', label: t('sessionStopped'), description: t('sessionStoppedDesc') },
  ]

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="card w-full max-w-lg p-6">
        <h2 className="text-xl font-bold text-primary mb-4">
          {webhook ? t('editWebhook') : t('createWebhook')}
        </h2>
        <form onSubmit={handleSubmit}>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-secondary mb-2">{t('webhookURL')}</label>
              <input
                type="url"
                value={url}
                onChange={e => setUrl(e.target.value)}
                className="input w-full"
                placeholder="https://example.com/webhook"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-secondary mb-2">
                {t('secret')} ({t('optional')})
                <span className="text-muted ml-2 font-normal">{t('secretDesc')}</span>
              </label>
              <input
                type="password"
                value={secret}
                onChange={e => setSecret(e.target.value)}
                className="input w-full"
                placeholder={webhook ? t('unchanged') : t('enterSecretPlaceholder')}
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-secondary mb-2">{t('events')}</label>
              <div className="space-y-2">
                {eventTypes.map(event => (
                  <label
                    key={event.id}
                    className={`flex items-center gap-3 p-3 rounded-lg border cursor-pointer transition-colors ${
                      selectedEvents.includes(event.id)
                        ? 'border-emerald-500/50 bg-emerald-500/10'
                        : 'border-border hover:border-gray-500/50'
                    }`}
                  >
                    <input
                      type="checkbox"
                      checked={selectedEvents.includes(event.id)}
                      onChange={() => toggleEvent(event.id)}
                      className="w-4 h-4 rounded border-gray-600 text-emerald-500 focus:ring-emerald-500"
                    />
                    <div>
                      <div className="font-medium text-primary">{event.label}</div>
                      <div className="text-xs text-muted">{event.description}</div>
                    </div>
                  </label>
                ))}
              </div>
              {selectedEvents.length === 0 && (
                <p className="text-xs text-red-400 mt-2">{t('selectAtLeastOneEvent')}</p>
              )}
            </div>
          </div>

          <div className="flex justify-end gap-2 mt-6">
            <button type="button" onClick={onClose} className="btn btn-ghost" disabled={isSaving}>
              {t('cancel')}
            </button>
            <button
              type="submit"
              className="btn btn-primary"
              disabled={selectedEvents.length === 0 || isSaving}
            >
              {isSaving ? <Loader2 className="w-4 h-4 animate-spin" /> : null}
              {webhook ? t('saveChanges') : t('createWebhook')}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default function WebhookList() {
  const navigate = useNavigate()
  const { t } = useLanguage()
  const [showModal, setShowModal] = useState(false)
  const [editingWebhook, setEditingWebhook] = useState<WebhookType | null>(null)

  // React Query hooks
  const { data: webhooks = [], isLoading, isFetching, error, refetch } = useWebhooks()
  const createWebhook = useCreateWebhook()
  const updateWebhook = useUpdateWebhook()
  const deleteWebhook = useDeleteWebhook()

  const handleCreate = (data: { url: string; secret?: string; events: string[] }) => {
    createWebhook.mutate(data, {
      onSuccess: () => {
        setShowModal(false)
      },
    })
  }

  const handleUpdate = (data: { url: string; secret?: string; events: string[] }) => {
    if (!editingWebhook) return
    updateWebhook.mutate(
      { id: editingWebhook.id, data },
      {
        onSuccess: () => {
          setShowModal(false)
          setEditingWebhook(null)
        },
      }
    )
  }

  const handleDelete = (webhook: WebhookType) => {
    if (!confirm(`${t('confirmDeleteWebhook')} "${webhook.id}"?`)) return
    deleteWebhook.mutate(webhook.id)
  }

  const handleToggle = (webhook: WebhookType) => {
    updateWebhook.mutate({ id: webhook.id, data: { is_active: !webhook.is_active } })
  }

  // 统计
  const stats = {
    total: webhooks.length,
    active: webhooks.filter(w => w.is_active).length,
    inactive: webhooks.filter(w => !w.is_active).length,
  }

  return (
    <div className="min-h-screen">
      {/* Header */}
      <header className="app-header">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate('/')} className="btn btn-ghost btn-icon">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-3">
            <Webhook className="w-6 h-6 text-emerald-400" />
            <span className="text-lg font-bold">{t('webhooksTitle')}</span>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button onClick={() => refetch()} className="btn btn-ghost btn-icon" disabled={isFetching}>
            <RefreshCw className={`w-4 h-4 ${isFetching ? 'animate-spin' : ''}`} />
          </button>
          <button
            className="btn btn-primary"
            onClick={() => {
              setEditingWebhook(null)
              setShowModal(true)
            }}
          >
            <Plus className="w-4 h-4" />
            {t('newWebhook')}
          </button>
        </div>
      </header>

      <div className="p-6">
        {/* Error */}
        {error && (
          <div className="mb-6 p-4 rounded-xl bg-red-500/10 border border-red-500/20 flex items-center gap-3">
            <AlertCircle className="w-5 h-5 text-red-400" />
            <span className="text-red-400">
              {error instanceof Error ? error.message : t('failedToLoadWebhooks')}
            </span>
          </div>
        )}

        {/* Stats */}
        <div className="grid grid-cols-3 gap-4 mb-8">
          <div className="card p-4">
            <div className="text-2xl font-bold text-primary">{stats.total}</div>
            <div className="text-sm text-muted">{t('totalWebhooks')}</div>
          </div>
          <div className="card p-4 border-emerald-500/30">
            <div className="text-2xl font-bold text-emerald-400">{stats.active}</div>
            <div className="text-sm text-muted">{t('active')}</div>
          </div>
          <div className="card p-4 border-gray-500/30">
            <div className="text-2xl font-bold text-gray-400">{stats.inactive}</div>
            <div className="text-sm text-muted">{t('inactive')}</div>
          </div>
        </div>

        {/* Description */}
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-primary mb-2">{t('webhooksTitle')}</h1>
          <p className="text-secondary">
            {t('webhooksDesc')}
          </p>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
          </div>
        ) : webhooks.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64 text-center">
            <Bell className="w-16 h-16 text-muted mb-4" />
            <p className="text-secondary text-lg">{t('noWebhooksFound')}</p>
            <p className="text-muted mt-2">{t('createWebhookToGetStarted')}</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {webhooks.map(webhook => (
              <WebhookCard
                key={webhook.id}
                webhook={webhook}
                onEdit={() => {
                  setEditingWebhook(webhook)
                  setShowModal(true)
                }}
                onDelete={() => handleDelete(webhook)}
                onToggle={() => handleToggle(webhook)}
              />
            ))}
          </div>
        )}
      </div>

      {/* Modal */}
      <WebhookModal
        isOpen={showModal}
        onClose={() => {
          setShowModal(false)
          setEditingWebhook(null)
        }}
        onSave={editingWebhook ? handleUpdate : handleCreate}
        webhook={editingWebhook}
        isSaving={createWebhook.isPending || updateWebhook.isPending}
      />
    </div>
  )
}
