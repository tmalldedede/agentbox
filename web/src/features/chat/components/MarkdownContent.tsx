import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeHighlight from 'rehype-highlight'
import rehypeRaw from 'rehype-raw'
import { Copy, Check } from 'lucide-react'
import { useState } from 'react'
import { cn } from '@/lib/utils'
import 'highlight.js/styles/github-dark.css'

interface MarkdownContentProps {
  content: string
  className?: string
}

// 代码块组件（独立组件以支持状态）
function CodeBlock({
  language,
  codeString,
  children,
  codeClassName,
  ...props
}: {
  language: string
  codeString: string
  children: React.ReactNode
  codeClassName?: string
  [key: string]: any
}) {
  const [copied, setCopied] = useState(false)

  const copyCode = async () => {
    await navigator.clipboard.writeText(codeString)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className='relative group my-4'>
      <div className='flex items-center justify-between bg-muted/50 px-4 py-2 rounded-t-lg border-b'>
        <span className='text-xs text-muted-foreground font-mono'>{language}</span>
        <button
          onClick={copyCode}
          className='flex items-center gap-1.5 px-2 py-1 text-xs text-muted-foreground hover:text-foreground transition-colors rounded hover:bg-muted'
        >
          {copied ? (
            <>
              <Check className='h-3 w-3' />
              <span>已复制</span>
            </>
          ) : (
            <>
              <Copy className='h-3 w-3' />
              <span>复制</span>
            </>
          )}
        </button>
      </div>
      <pre className={cn('overflow-x-auto rounded-b-lg p-4 bg-muted/30', codeClassName)}>
        <code className={codeClassName} {...props}>
          {children}
        </code>
      </pre>
    </div>
  )
}

export function MarkdownContent({ content, className }: MarkdownContentProps) {
  return (
    <div className={cn('markdown-content prose prose-sm dark:prose-invert max-w-none', className)}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[rehypeRaw, rehypeHighlight]}
        components={{
          // 代码块组件
          code({ node, inline, className: codeClassName, children, ...props }: any) {
            const match = /language-(\w+)/.exec(codeClassName || '')
            const language = match ? match[1] : ''
            const codeString = String(children).replace(/\n$/, '')

            if (!inline && match) {
              return (
                <CodeBlock
                  language={language}
                  codeString={codeString}
                  codeClassName={codeClassName}
                  {...props}
                >
                  {children}
                </CodeBlock>
              )
            }

            return (
              <code
                className={cn(
                  'relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm',
                  codeClassName
                )}
                {...props}
              >
                {children}
              </code>
            )
          },
          // 段落组件
          p({ children }: any) {
            return <p className='mb-4 last:mb-0 leading-relaxed'>{children}</p>
          },
          // 标题组件
          h1({ children }: any) {
            return <h1 className='text-2xl font-bold mt-6 mb-4 first:mt-0'>{children}</h1>
          },
          h2({ children }: any) {
            return <h2 className='text-xl font-semibold mt-5 mb-3 first:mt-0'>{children}</h2>
          },
          h3({ children }: any) {
            return <h3 className='text-lg font-semibold mt-4 mb-2 first:mt-0'>{children}</h3>
          },
          // 列表组件
          ul({ children }: any) {
            return <ul className='list-disc list-inside mb-4 space-y-1'>{children}</ul>
          },
          ol({ children }: any) {
            return <ol className='list-decimal list-inside mb-4 space-y-1'>{children}</ol>
          },
          li({ children }: any) {
            return <li className='ml-4'>{children}</li>
          },
          // 引用组件
          blockquote({ children }: any) {
            return (
              <blockquote className='border-l-4 border-primary/50 pl-4 my-4 italic text-muted-foreground'>
                {children}
              </blockquote>
            )
          },
          // 链接组件
          a({ href, children }: any) {
            return (
              <a
                href={href}
                target='_blank'
                rel='noopener noreferrer'
                className='text-primary hover:underline'
              >
                {children}
              </a>
            )
          },
          // 表格组件
          table({ children }: any) {
            return (
              <div className='overflow-x-auto my-4'>
                <table className='min-w-full border-collapse border border-border'>
                  {children}
                </table>
              </div>
            )
          },
          thead({ children }: any) {
            return <thead className='bg-muted'>{children}</thead>
          },
          th({ children }: any) {
            return (
              <th className='border border-border px-4 py-2 text-left font-semibold'>
                {children}
              </th>
            )
          },
          td({ children }: any) {
            return <td className='border border-border px-4 py-2'>{children}</td>
          },
          // 水平线组件
          hr() {
            return <hr className='my-6 border-border' />
          },
          // 强调组件
          strong({ children }: any) {
            return <strong className='font-semibold'>{children}</strong>
          },
          em({ children }: any) {
            return <em className='italic'>{children}</em>
          },
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  )
}
