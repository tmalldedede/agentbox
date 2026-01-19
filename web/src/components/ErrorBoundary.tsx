import { Component, ErrorInfo, ReactNode } from 'react'
import { AlertTriangle, RefreshCw } from 'lucide-react'

interface Props {
  children: ReactNode
  fallback?: (error: Error, reset: () => void) => ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
}

/**
 * 错误边界组件
 * 捕获子组件树中的 JavaScript 错误，记录错误并显示备用 UI
 */
export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = {
      hasError: false,
      error: null,
    }
  }

  static getDerivedStateFromError(error: Error): State {
    return {
      hasError: true,
      error,
    }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // 在这里可以将错误日志发送到错误追踪服务（如 Sentry）
    console.error('ErrorBoundary caught an error:', error, errorInfo)
  }

  handleReset = () => {
    this.setState({
      hasError: false,
      error: null,
    })
  }

  render() {
    if (this.state.hasError && this.state.error) {
      // 如果提供了自定义 fallback，使用它
      if (this.props.fallback) {
        return this.props.fallback(this.state.error, this.handleReset)
      }

      // 默认错误 UI
      return (
        <div className="min-h-screen flex items-center justify-center p-4 bg-[var(--bg-primary)]">
          <div className="max-w-md w-full bg-[var(--bg-secondary)] rounded-xl border border-[var(--border-color)] p-8">
            <div className="flex items-center justify-center w-16 h-16 mx-auto mb-4 rounded-full bg-red-500/10">
              <AlertTriangle className="w-8 h-8 text-red-500" />
            </div>

            <h1 className="text-2xl font-bold text-center mb-2 text-[var(--text-primary)]">
              出错了
            </h1>

            <p className="text-center text-[var(--text-secondary)] mb-6">
              应用程序遇到了一个错误，我们很抱歉给您带来不便。
            </p>

            <div className="bg-[var(--bg-primary)] rounded-lg p-4 mb-6 max-h-40 overflow-auto">
              <p className="text-sm font-mono text-red-400">
                {this.state.error.message}
              </p>
            </div>

            <div className="flex gap-3">
              <button
                onClick={this.handleReset}
                className="flex-1 flex items-center justify-center gap-2 px-4 py-2 bg-[var(--accent-green)] text-white rounded-lg hover:opacity-90 transition-opacity"
              >
                <RefreshCw className="w-4 h-4" />
                重试
              </button>

              <button
                onClick={() => window.location.reload()}
                className="flex-1 px-4 py-2 bg-[var(--bg-primary)] text-[var(--text-primary)] rounded-lg border border-[var(--border-color)] hover:bg-[var(--bg-tertiary)] transition-colors"
              >
                刷新页面
              </button>
            </div>

            {process.env.NODE_ENV === 'development' && this.state.error.stack && (
              <details className="mt-4">
                <summary className="cursor-pointer text-sm text-[var(--text-secondary)] hover:text-[var(--text-primary)]">
                  查看堆栈跟踪
                </summary>
                <pre className="mt-2 text-xs bg-[var(--bg-primary)] rounded p-3 overflow-auto max-h-60 text-[var(--text-secondary)]">
                  {this.state.error.stack}
                </pre>
              </details>
            )}
          </div>
        </div>
      )
    }

    return this.props.children
  }
}

/**
 * 路由级别的错误边界（更轻量的错误提示）
 */
export function RouteErrorBoundary({ children }: { children: ReactNode }) {
  return (
    <ErrorBoundary
      fallback={(error, reset) => (
        <div className="flex flex-col items-center justify-center min-h-[400px] p-8">
          <AlertTriangle className="w-12 h-12 text-red-500 mb-4" />
          <h2 className="text-xl font-semibold mb-2 text-[var(--text-primary)]">
            页面加载失败
          </h2>
          <p className="text-[var(--text-secondary)] mb-4 text-center max-w-md">
            {error.message}
          </p>
          <button
            onClick={reset}
            className="flex items-center gap-2 px-4 py-2 bg-[var(--accent-green)] text-white rounded-lg hover:opacity-90 transition-opacity"
          >
            <RefreshCw className="w-4 h-4" />
            重新加载
          </button>
        </div>
      )}
    >
      {children}
    </ErrorBoundary>
  )
}
