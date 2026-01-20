import { useEffect, useRef, useCallback } from 'react'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'

interface TerminalViewerProps {
  /** 要显示的内容 */
  content?: string
  /** 是否正在加载 */
  loading?: boolean
  /** 空状态占位符 */
  placeholder?: string
  /** 主题 */
  theme?: 'dark' | 'light'
  /** 字体大小 */
  fontSize?: number
  /** 最小高度 */
  minHeight?: string
}

// ANSI 颜色代码
const ANSI = {
  reset: '\x1b[0m',
  bold: '\x1b[1m',
  dim: '\x1b[2m',
  // 前景色
  black: '\x1b[30m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
  white: '\x1b[37m',
  gray: '\x1b[90m',
  // 明亮前景色
  brightRed: '\x1b[91m',
  brightGreen: '\x1b[92m',
  brightYellow: '\x1b[93m',
  brightBlue: '\x1b[94m',
  brightMagenta: '\x1b[95m',
  brightCyan: '\x1b[96m',
  brightWhite: '\x1b[97m',
}

export { ANSI }

/**
 * 基于 xterm.js 的终端日志查看器
 */
export default function TerminalViewer({
  content = '',
  loading = false,
  placeholder = 'No output yet...',
  theme = 'dark',
  fontSize = 13,
  minHeight = '400px',
}: TerminalViewerProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const terminalRef = useRef<Terminal | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)

  // 初始化终端
  useEffect(() => {
    if (!containerRef.current) return

    const terminal = new Terminal({
      cursorBlink: false,
      cursorStyle: 'block',
      disableStdin: true,
      fontSize,
      fontFamily: 'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace',
      lineHeight: 1.4,
      scrollback: 10000,
      convertEol: true,
      theme:
        theme === 'dark'
          ? {
              background: '#0d1117',
              foreground: '#c9d1d9',
              cursor: '#58a6ff',
              selectionBackground: '#3b5070',
              black: '#484f58',
              red: '#ff7b72',
              green: '#3fb950',
              yellow: '#d29922',
              blue: '#58a6ff',
              magenta: '#bc8cff',
              cyan: '#39c5cf',
              white: '#b1bac4',
              brightBlack: '#6e7681',
              brightRed: '#ffa198',
              brightGreen: '#56d364',
              brightYellow: '#e3b341',
              brightBlue: '#79c0ff',
              brightMagenta: '#d2a8ff',
              brightCyan: '#56d4dd',
              brightWhite: '#f0f6fc',
            }
          : {
              background: '#ffffff',
              foreground: '#24292f',
              cursor: '#0969da',
              selectionBackground: '#b6e3ff',
              black: '#24292f',
              red: '#cf222e',
              green: '#116329',
              yellow: '#9a6700',
              blue: '#0969da',
              magenta: '#8250df',
              cyan: '#1b7c83',
              white: '#6e7781',
              brightBlack: '#57606a',
              brightRed: '#a40e26',
              brightGreen: '#1a7f37',
              brightYellow: '#bf8700',
              brightBlue: '#218bff',
              brightMagenta: '#a475f9',
              brightCyan: '#3192aa',
              brightWhite: '#8c959f',
            },
    })

    const fitAddon = new FitAddon()
    const webLinksAddon = new WebLinksAddon()

    terminal.loadAddon(fitAddon)
    terminal.loadAddon(webLinksAddon)

    terminal.open(containerRef.current)
    fitAddon.fit()

    terminalRef.current = terminal
    fitAddonRef.current = fitAddon

    // 监听窗口大小变化
    const handleResize = () => {
      fitAddon.fit()
    }
    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
      terminal.dispose()
    }
  }, [fontSize, theme])

  // 更新内容
  useEffect(() => {
    const terminal = terminalRef.current
    if (!terminal) return

    terminal.clear()

    if (loading) {
      terminal.writeln(`${ANSI.yellow}Loading...${ANSI.reset}`)
      return
    }

    if (!content) {
      terminal.writeln(`${ANSI.gray}${placeholder}${ANSI.reset}`)
      return
    }

    // 写入内容
    const lines = content.split('\n')
    lines.forEach(line => {
      terminal.writeln(line)
    })

    // 滚动到顶部，让用户从头开始看
    terminal.scrollToTop()
  }, [content, loading, placeholder])

  // 调整大小
  const handleContainerResize = useCallback(() => {
    if (fitAddonRef.current) {
      setTimeout(() => {
        fitAddonRef.current?.fit()
      }, 0)
    }
  }, [])

  // 监听容器大小变化
  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    const resizeObserver = new ResizeObserver(handleContainerResize)
    resizeObserver.observe(container)

    return () => {
      resizeObserver.disconnect()
    }
  }, [handleContainerResize])

  return (
    <div
      ref={containerRef}
      className="terminal-viewer"
      style={{
        minHeight,
        width: '100%',
        height: '100%',
        borderRadius: '8px',
        overflow: 'hidden',
      }}
    />
  )
}

/**
 * 格式化执行记录为带颜色的日志
 */
export function formatExecutionLog(execution: {
  id: string
  prompt: string
  output?: string
  error?: string
  exitCode: number
  status: string
  startedAt: Date
  endedAt?: Date
}): string {
  const lines: string[] = []

  // 状态颜色
  const statusColor =
    execution.status === 'success'
      ? ANSI.green
      : execution.status === 'failed'
        ? ANSI.red
        : execution.status === 'running'
          ? ANSI.yellow
          : ANSI.gray

  // 头部
  lines.push(
    `${ANSI.bold}${ANSI.cyan}=== Execution ${execution.id} ${ANSI.reset}${statusColor}[${execution.status}]${ANSI.reset} ${ANSI.bold}${ANSI.cyan}===${ANSI.reset}`
  )
  lines.push(
    `${ANSI.gray}Started: ${execution.startedAt.toLocaleString()}${ANSI.reset}`
  )
  if (execution.endedAt) {
    lines.push(
      `${ANSI.gray}Ended: ${execution.endedAt.toLocaleString()}${ANSI.reset}`
    )
  }

  // Prompt
  lines.push('')
  lines.push(`${ANSI.green}❯${ANSI.reset} ${ANSI.brightWhite}${execution.prompt}${ANSI.reset}`)

  // Output
  if (execution.output) {
    lines.push('')
    lines.push(`${ANSI.dim}--- Output ---${ANSI.reset}`)
    lines.push(execution.output)
  }

  // Error
  if (execution.error) {
    lines.push('')
    lines.push(`${ANSI.red}--- Error ---${ANSI.reset}`)
    lines.push(`${ANSI.red}${execution.error}${ANSI.reset}`)
  }

  // Exit code
  if (execution.exitCode !== 0) {
    lines.push('')
    lines.push(
      `${ANSI.red}Exit Code: ${execution.exitCode}${ANSI.reset}`
    )
  }

  return lines.join('\n')
}

/**
 * 格式化多个执行记录
 */
export function formatExecutionLogs(
  executions: Array<{
    id: string
    prompt: string
    output?: string
    error?: string
    exitCode: number
    status: string
    startedAt: Date
    endedAt?: Date
  }>
): string {
  return executions.map(formatExecutionLog).join('\n\n')
}
