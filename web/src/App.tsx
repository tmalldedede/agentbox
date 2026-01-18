import { useState, useEffect } from 'react'
import { Box, Play, Square, Trash2, Terminal, RefreshCw, Plus } from 'lucide-react'
import type { Session, Agent } from './types'
import { api } from './services/api'

function App() {
  const [sessions, setSessions] = useState<Session[]>([])
  const [agents, setAgents] = useState<Agent[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showCreate, setShowCreate] = useState(false)

  const fetchData = async () => {
    try {
      setLoading(true)
      const [sessionsData, agentsData] = await Promise.all([
        api.listSessions(),
        api.listAgents(),
      ])
      setSessions(sessionsData || [])
      setAgents(agentsData || [])
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch data')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
    const interval = setInterval(fetchData, 5000)
    return () => clearInterval(interval)
  }, [])

  const handleDelete = async (id: string) => {
    if (!confirm('确定删除此会话?')) return
    try {
      await api.deleteSession(id)
      fetchData()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete')
    }
  }

  const handleStop = async (id: string) => {
    try {
      await api.stopSession(id)
      fetchData()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to stop')
    }
  }

  const handleStart = async (id: string) => {
    try {
      await api.startSession(id)
      fetchData()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to start')
    }
  }

  const statusColor = (status: Session['status']) => {
    switch (status) {
      case 'running': return 'text-green-400'
      case 'stopped': return 'text-gray-400'
      case 'creating': return 'text-yellow-400'
      case 'error': return 'text-red-400'
      default: return 'text-gray-400'
    }
  }

  return (
    <div className="min-h-screen bg-gray-900">
      {/* Header */}
      <header className="bg-gray-800 border-b border-gray-700">
        <div className="max-w-7xl mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Box className="w-8 h-8 text-blue-400" />
            <h1 className="text-xl font-bold text-white">AgentBox</h1>
            <span className="text-xs text-gray-500">v0.1.0</span>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={fetchData}
              className="p-2 text-gray-400 hover:text-white rounded-lg hover:bg-gray-700"
            >
              <RefreshCw className="w-5 h-5" />
            </button>
            <button
              onClick={() => setShowCreate(true)}
              className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg"
            >
              <Plus className="w-4 h-4" />
              新建会话
            </button>
          </div>
        </div>
      </header>

      {/* Main */}
      <main className="max-w-7xl mx-auto px-4 py-6">
        {error && (
          <div className="mb-4 p-4 bg-red-900/50 border border-red-700 rounded-lg text-red-300">
            {error}
          </div>
        )}

        {/* Agents */}
        <section className="mb-8">
          <h2 className="text-lg font-semibold text-white mb-4">支持的 Agent</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {agents.map((agent) => (
              <div
                key={agent.name}
                className="p-4 bg-gray-800 rounded-lg border border-gray-700"
              >
                <h3 className="font-medium text-white">{agent.display_name}</h3>
                <p className="text-sm text-gray-400 mt-1">{agent.description}</p>
                <div className="mt-2 text-xs text-gray-500">
                  需要: {agent.required_env.join(', ')}
                </div>
              </div>
            ))}
          </div>
        </section>

        {/* Sessions */}
        <section>
          <h2 className="text-lg font-semibold text-white mb-4">会话列表</h2>
          {loading && sessions.length === 0 ? (
            <div className="text-center py-8 text-gray-500">加载中...</div>
          ) : sessions.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              暂无会话，点击"新建会话"开始
            </div>
          ) : (
            <div className="space-y-3">
              {sessions.map((session) => (
                <div
                  key={session.id}
                  className="p-4 bg-gray-800 rounded-lg border border-gray-700 flex items-center justify-between"
                >
                  <div className="flex items-center gap-4">
                    <Terminal className="w-5 h-5 text-gray-500" />
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-mono text-white">{session.id}</span>
                        <span className="px-2 py-0.5 text-xs rounded bg-gray-700 text-gray-300">
                          {session.agent}
                        </span>
                        <span className={`text-sm ${statusColor(session.status)}`}>
                          ● {session.status}
                        </span>
                      </div>
                      <div className="text-sm text-gray-500 mt-1">
                        {session.workspace}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {session.status === 'running' ? (
                      <button
                        onClick={() => handleStop(session.id)}
                        className="p-2 text-yellow-400 hover:text-yellow-300 rounded hover:bg-gray-700"
                        title="停止"
                      >
                        <Square className="w-4 h-4" />
                      </button>
                    ) : session.status === 'stopped' ? (
                      <button
                        onClick={() => handleStart(session.id)}
                        className="p-2 text-green-400 hover:text-green-300 rounded hover:bg-gray-700"
                        title="启动"
                      >
                        <Play className="w-4 h-4" />
                      </button>
                    ) : null}
                    <button
                      onClick={() => handleDelete(session.id)}
                      className="p-2 text-red-400 hover:text-red-300 rounded hover:bg-gray-700"
                      title="删除"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>
      </main>

      {/* Create Modal */}
      {showCreate && (
        <CreateSessionModal
          agents={agents}
          onClose={() => setShowCreate(false)}
          onCreated={() => {
            setShowCreate(false)
            fetchData()
          }}
        />
      )}
    </div>
  )
}

function CreateSessionModal({
  agents,
  onClose,
  onCreated,
}: {
  agents: Agent[]
  onClose: () => void
  onCreated: () => void
}) {
  const [agent, setAgent] = useState(agents[0]?.name || '')
  const [workspace, setWorkspace] = useState('/tmp/myproject')
  const [apiKey, setApiKey] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const selectedAgent = agents.find((a) => a.name === agent)
  const envKey = selectedAgent?.required_env[0] || ''

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      setLoading(true)
      setError(null)
      await api.createSession({
        agent,
        workspace,
        env: apiKey ? { [envKey]: apiKey } : undefined,
      })
      onCreated()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4">
      <div className="bg-gray-800 rounded-lg w-full max-w-md p-6">
        <h2 className="text-lg font-semibold text-white mb-4">新建会话</h2>

        {error && (
          <div className="mb-4 p-3 bg-red-900/50 border border-red-700 rounded text-red-300 text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm text-gray-400 mb-1">Agent</label>
            <select
              value={agent}
              onChange={(e) => setAgent(e.target.value)}
              className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white"
            >
              {agents.map((a) => (
                <option key={a.name} value={a.name}>
                  {a.display_name}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-sm text-gray-400 mb-1">工作目录</label>
            <input
              type="text"
              value={workspace}
              onChange={(e) => setWorkspace(e.target.value)}
              className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white"
              placeholder="/path/to/project"
            />
          </div>

          {envKey && (
            <div>
              <label className="block text-sm text-gray-400 mb-1">{envKey}</label>
              <input
                type="password"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white"
                placeholder="sk-..."
              />
            </div>
          )}

          <div className="flex gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded"
            >
              取消
            </button>
            <button
              type="submit"
              disabled={loading}
              className="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded disabled:opacity-50"
            >
              {loading ? '创建中...' : '创建'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default App
