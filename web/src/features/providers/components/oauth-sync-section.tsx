import { useState } from 'react'
import { RefreshCw, Download, Upload, CheckCircle2, XCircle, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  useOAuthSyncStatus,
  useSyncFromClaudeCli,
  useSyncFromCodexCli,
  useSyncToClaudeCli,
} from '@/hooks/useOAuthSync'
import type { Provider } from '@/types'

type OAuthSyncSectionProps = {
  provider: Provider
}

export function OAuthSyncSection({ provider }: OAuthSyncSectionProps) {
  const { data: status, isLoading } = useOAuthSyncStatus()
  const syncFromClaudeCli = useSyncFromClaudeCli()
  const syncFromCodexCli = useSyncFromCodexCli()
  const syncToClaudeCli = useSyncToClaudeCli()

  const [isExpanded, setIsExpanded] = useState(false)

  // 判断当前 provider 是否支持 OAuth 同步
  const isAnthropic = provider.base_url?.includes('anthropic.com')
  const isOpenAI = provider.name === 'OpenAI' || provider.base_url?.includes('openai.com')

  // 不支持 OAuth 同步的 provider 不显示此功能
  if (!isAnthropic && !isOpenAI) {
    return null
  }

  const handleSyncFromCli = () => {
    if (isAnthropic) {
      syncFromClaudeCli.mutate(provider.id)
    } else if (isOpenAI) {
      syncFromCodexCli.mutate(provider.id)
    }
  }

  const handleSyncToCli = () => {
    if (isAnthropic) {
      syncToClaudeCli.mutate(provider.id)
    }
  }

  return (
    <Card className='mt-4'>
      <CardHeader>
        <div className='flex items-center justify-between'>
          <div>
            <CardTitle className='text-base flex items-center gap-2'>
              <RefreshCw className='h-4 w-4' />
              OAuth 令牌同步
            </CardTitle>
            <CardDescription className='text-sm mt-1'>
              {isAnthropic && '与 Claude Code CLI 同步 OAuth 令牌'}
              {isOpenAI && '与 Codex CLI 同步 OAuth 令牌'}
            </CardDescription>
          </div>
          <Button
            variant='ghost'
            size='sm'
            onClick={() => setIsExpanded(!isExpanded)}
          >
            {isExpanded ? '收起' : '展开'}
          </Button>
        </div>
      </CardHeader>

      {isExpanded && (
        <CardContent className='space-y-4'>
          {/* CLI 可用性状态 */}
          <div className='grid grid-cols-2 gap-4 p-4 bg-muted/50 rounded-lg'>
            <div className='flex items-center justify-between'>
              <span className='text-sm font-medium'>
                {isAnthropic ? 'Claude Code CLI' : 'Codex CLI'}
              </span>
              {isLoading ? (
                <Loader2 className='h-4 w-4 animate-spin' />
              ) : (
                <Badge
                  variant={
                    (isAnthropic && status?.claude_cli_available) ||
                    (isOpenAI && status?.codex_cli_available)
                      ? 'default'
                      : 'secondary'
                  }
                  className='gap-1'
                >
                  {(isAnthropic && status?.claude_cli_available) ||
                  (isOpenAI && status?.codex_cli_available) ? (
                    <>
                      <CheckCircle2 className='h-3 w-3' />
                      可用
                    </>
                  ) : (
                    <>
                      <XCircle className='h-3 w-3' />
                      不可用
                    </>
                  )}
                </Badge>
              )}
            </div>
            <div className='flex items-center justify-between'>
              <span className='text-sm font-medium'>平台</span>
              <Badge variant='outline'>{status?.platform || 'unknown'}</Badge>
            </div>
          </div>

          {/* 同步操作按钮 */}
          <div className='space-y-3'>
            {/* 从 CLI 导入 */}
            <div className='flex items-start gap-3'>
              <Button
                variant='outline'
                className='flex-1'
                onClick={handleSyncFromCli}
                disabled={
                  isLoading ||
                  syncFromClaudeCli.isPending ||
                  syncFromCodexCli.isPending ||
                  !(
                    (isAnthropic && status?.claude_cli_available) ||
                    (isOpenAI && status?.codex_cli_available)
                  )
                }
              >
                {(syncFromClaudeCli.isPending || syncFromCodexCli.isPending) && (
                  <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                )}
                <Download className='mr-2 h-4 w-4' />
                从 {isAnthropic ? 'Claude Code' : 'Codex'} CLI 导入
              </Button>
              <div className='flex-1 text-sm text-muted-foreground'>
                读取 {isAnthropic ? 'Claude Code' : 'Codex'} CLI 中保存的 OAuth 令牌，并添加到此
                Provider
              </div>
            </div>

            {/* 导出到 CLI (仅 Anthropic) */}
            {isAnthropic && (
              <div className='flex items-start gap-3'>
                <Button
                  variant='outline'
                  className='flex-1'
                  onClick={handleSyncToCli}
                  disabled={
                    isLoading ||
                    syncToClaudeCli.isPending ||
                    !status?.claude_cli_available
                  }
                >
                  {syncToClaudeCli.isPending && (
                    <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                  )}
                  <Upload className='mr-2 h-4 w-4' />
                  导出到 Claude Code CLI
                </Button>
                <div className='flex-1 text-sm text-muted-foreground'>
                  将 AgentBox 中刷新后的令牌写回 Claude Code CLI（防止 CLI 登出）
                </div>
              </div>
            )}
          </div>

          {/* 说明文本 */}
          <div className='text-xs text-muted-foreground p-3 bg-muted/30 rounded-md'>
            <p className='font-medium mb-1'>💡 使用提示</p>
            <ul className='list-disc list-inside space-y-1 ml-2'>
              <li>
                <strong>导入令牌</strong>：从 CLI 读取现有的 OAuth 令牌到 AgentBox
              </li>
              {isAnthropic && (
                <li>
                  <strong>导出令牌</strong>：将 AgentBox 刷新后的令牌写回 CLI，防止 CLI
                  登出
                </li>
              )}
              <li>令牌存储位置：macOS Keychain 或 ~/.{isAnthropic ? 'claude' : 'codex'} 目录</li>
            </ul>
          </div>
        </CardContent>
      )}
    </Card>
  )
}
