import { useState } from 'react'
import { Loader2, ExternalLink } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useConfigureProviderKey } from '@/hooks/useProviders'
import type { Provider } from '@/types'

type ConfigureKeyDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  provider: Provider
}

export function ConfigureKeyDialog({
  open,
  onOpenChange,
  provider,
}: ConfigureKeyDialogProps) {
  const [apiKey, setApiKey] = useState('')
  const configureKey = useConfigureProviderKey()

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!apiKey.trim()) return
    configureKey.mutate(
      { id: provider.id, apiKey: apiKey.trim() },
      {
        onSuccess: () => {
          setApiKey('')
          onOpenChange(false)
        },
      }
    )
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle className='flex items-center gap-2'>
            <span className='text-lg'>{provider.icon || '☁️'}</span>
            {provider.is_configured ? 'Update API Key' : 'Configure API Key'}
          </DialogTitle>
          <DialogDescription>
            {provider.is_configured
              ? `Replace the existing API key for ${provider.name}.`
              : `Enter your API key for ${provider.name} to enable this provider.`}
            {provider.api_key_masked && (
              <span className='block mt-1 font-mono text-xs'>
                Current: {provider.api_key_masked}
              </span>
            )}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className='space-y-4 py-4'>
            <div className='space-y-2'>
              <Label htmlFor='api-key'>API Key</Label>
              <Input
                id='api-key'
                type='password'
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder={
                  provider.env_config
                    ? `e.g. ${Object.keys(provider.env_config)[0] || 'sk-...'}`
                    : 'sk-...'
                }
                className='font-mono'
                autoFocus
              />
            </div>
            {provider.api_key_url && (
              <a
                href={provider.api_key_url}
                target='_blank'
                rel='noopener noreferrer'
                className='inline-flex items-center gap-1 text-sm text-primary hover:underline'
              >
                Get API key
                <ExternalLink className='h-3 w-3' />
              </a>
            )}
          </div>
          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type='submit' disabled={!apiKey.trim() || configureKey.isPending}>
              {configureKey.isPending && (
                <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              )}
              Save Key
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
