import { useEffect, useState } from 'react'
import { Loader2, ArrowLeft, Eye, EyeOff, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useCreateProvider, useUpdateProvider, useProviderTemplates, useProbeModels, useFetchProviderModels } from '@/hooks/useProviders'
import type { Provider, ProviderCategory } from '@/types'
import { getProviderIcon } from '../data/icons'

type ProviderActionDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentRow?: Provider | null
}

const categoryLabels: Record<ProviderCategory, string> = {
  official: 'Official',
  cn_official: 'CN Official',
  aggregator: 'Aggregator',
  third_party: 'Compatible',
}

const categoryOrder: ProviderCategory[] = ['official', 'cn_official', 'aggregator', 'third_party']

export function ProviderActionDialog({
  open,
  onOpenChange,
  currentRow,
}: ProviderActionDialogProps) {
  const isEdit = !!currentRow
  const createProvider = useCreateProvider()
  const updateProvider = useUpdateProvider()
  const probeModels = useProbeModels()
  const fetchModels = useFetchProviderModels()
  const { data: templates = [] } = useProviderTemplates()

  const [step, setStep] = useState<'select' | 'configure'>('select')
  const [selectedTemplate, setSelectedTemplate] = useState<Provider | null>(null)
  const [showKey, setShowKey] = useState(false)
  const [fetchedModels, setFetchedModels] = useState<string[]>([])

  const [formData, setFormData] = useState({
    id: '',
    name: '',
    base_url: '',
    api_key: '',
    default_model: '',
  })

  useEffect(() => {
    if (!open) {
      // Reset on close
      setTimeout(() => {
        setStep('select')
        setSelectedTemplate(null)
        setShowKey(false)
        setFetchedModels([])
        setFormData({ id: '', name: '', base_url: '', api_key: '', default_model: '' })
      }, 300)
    }
    if (open && currentRow) {
      setStep('configure')
      setFormData({
        id: currentRow.id,
        name: currentRow.name,
        base_url: currentRow.base_url || '',
        api_key: '',
        default_model: currentRow.default_model || '',
      })
    }
  }, [open, currentRow])

  const handleSelectTemplate = (tmpl: Provider) => {
    setSelectedTemplate(tmpl)
    setStep('configure')
    setFormData({
      id: '',
      name: tmpl.name,
      base_url: tmpl.base_url || '',
      api_key: '',
      default_model: tmpl.default_model || '',
    })
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (isEdit) {
      updateProvider.mutate(
        {
          id: currentRow.id,
          req: {
            name: formData.name,
            base_url: formData.base_url || undefined,
            default_model: formData.default_model || undefined,
          },
        },
        { onSuccess: () => onOpenChange(false) }
      )
    } else {
      createProvider.mutate(
        {
          id: formData.id,
          name: formData.name,
          template_id: selectedTemplate?.id,
          base_url: formData.base_url || undefined,
          api_key: formData.api_key || undefined,
          models: formData.default_model ? [formData.default_model] : undefined,
        },
        { onSuccess: () => onOpenChange(false) }
      )
    }
  }

  const handleFetchModels = () => {
    if (isEdit && currentRow) {
      fetchModels.mutate(currentRow.id, {
        onSuccess: (models) => setFetchedModels(models),
      })
    } else if (formData.api_key && (formData.base_url || selectedTemplate)) {
      const baseURL = formData.base_url || selectedTemplate?.base_url || ''
      const agents = selectedTemplate?.agents || []
      probeModels.mutate(
        { baseURL, apiKey: formData.api_key, agents },
        { onSuccess: (models) => setFetchedModels(models) }
      )
    }
  }

  const isModelFetching = probeModels.isPending || fetchModels.isPending
  const canFetchModels = isEdit
    ? currentRow?.is_configured
    : !!(formData.api_key && (formData.base_url || selectedTemplate?.base_url))

  const isPending = createProvider.isPending || updateProvider.isPending

  // Group templates by category
  const groupedTemplates = categoryOrder
    .map(cat => ({
      category: cat,
      label: categoryLabels[cat],
      items: templates.filter(t => t.category === cat),
    }))
    .filter(g => g.items.length > 0)

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className={cn(
        'transition-all',
        step === 'select' ? 'sm:max-w-2xl' : 'sm:max-w-lg'
      )}>
        {step === 'select' && !isEdit ? (
          <>
            <DialogHeader>
              <DialogTitle>Add Provider</DialogTitle>
              <DialogDescription>
                Select a provider template to get started.
              </DialogDescription>
            </DialogHeader>
            <div className='max-h-[60vh] overflow-y-auto space-y-4 py-2'>
              {groupedTemplates.map(group => (
                <div key={group.category}>
                  <h4 className='text-sm font-medium text-muted-foreground mb-2'>
                    {group.label}
                  </h4>
                  <div className='grid grid-cols-2 sm:grid-cols-3 gap-2'>
                    {group.items.map(tmpl => {
                      const iconInfo = getProviderIcon(tmpl.icon)
                      return (
                        <button
                          key={tmpl.id}
                          type='button'
                          onClick={() => handleSelectTemplate(tmpl)}
                          className='flex items-center gap-2.5 p-3 rounded-lg border hover:bg-accent hover:border-primary/30 transition-colors text-left'
                        >
                          <div className={`w-8 h-8 rounded-md flex items-center justify-center text-xs font-bold shrink-0 ${iconInfo.color}`}>
                            {iconInfo.label}
                          </div>
                          <div className='min-w-0'>
                            <div className='text-sm font-medium truncate'>{tmpl.name}</div>
                            <div className='text-xs text-muted-foreground truncate'>
                              {tmpl.agents?.join(', ')}
                            </div>
                          </div>
                        </button>
                      )
                    })}
                  </div>
                </div>
              ))}
            </div>
          </>
        ) : (
          <>
            <DialogHeader>
              <div className='flex items-center gap-2'>
                {!isEdit && (
                  <Button
                    type='button'
                    variant='ghost'
                    size='icon'
                    className='h-7 w-7'
                    onClick={() => setStep('select')}
                  >
                    <ArrowLeft className='h-4 w-4' />
                  </Button>
                )}
                <div>
                  <DialogTitle>
                    {isEdit ? 'Edit Provider' : `Configure ${selectedTemplate?.name || 'Provider'}`}
                  </DialogTitle>
                  <DialogDescription>
                    {isEdit
                      ? 'Update provider configuration.'
                      : selectedTemplate?.description || 'Fill in the provider details.'}
                  </DialogDescription>
                </div>
              </div>
            </DialogHeader>
            <form onSubmit={handleSubmit}>
              <div className='space-y-4 py-4'>
                {!isEdit && (
                  <div className='grid grid-cols-2 gap-4'>
                    <div className='space-y-2'>
                      <Label htmlFor='provider-id'>ID *</Label>
                      <Input
                        id='provider-id'
                        value={formData.id}
                        onChange={(e) =>
                          setFormData({ ...formData, id: e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, '') })
                        }
                        placeholder='my-provider'
                        required
                        className='font-mono'
                      />
                    </div>
                    <div className='space-y-2'>
                      <Label htmlFor='provider-name'>Name *</Label>
                      <Input
                        id='provider-name'
                        value={formData.name}
                        onChange={(e) =>
                          setFormData({ ...formData, name: e.target.value })
                        }
                        required
                      />
                    </div>
                  </div>
                )}

                {isEdit && (
                  <div className='space-y-2'>
                    <Label htmlFor='provider-name'>Name</Label>
                    <Input
                      id='provider-name'
                      value={formData.name}
                      onChange={(e) =>
                        setFormData({ ...formData, name: e.target.value })
                      }
                    />
                  </div>
                )}

                <div className='space-y-2'>
                  <Label htmlFor='provider-base-url'>API Base URL</Label>
                  <Input
                    id='provider-base-url'
                    value={formData.base_url}
                    onChange={(e) =>
                      setFormData({ ...formData, base_url: e.target.value })
                    }
                    placeholder='https://api.example.com/v1'
                    className='font-mono text-sm'
                  />
                  <p className='text-xs text-muted-foreground'>
                    Do NOT include /chat/completions in the URL
                  </p>
                </div>

                {!isEdit && (
                  <div className='space-y-2'>
                    <Label htmlFor='provider-api-key'>API Key</Label>
                    <div className='relative'>
                      <Input
                        id='provider-api-key'
                        type={showKey ? 'text' : 'password'}
                        value={formData.api_key}
                        onChange={(e) =>
                          setFormData({ ...formData, api_key: e.target.value })
                        }
                        placeholder={selectedTemplate?.requires_ak === false ? 'Not required' : 'sk-...'}
                        className='pr-10 font-mono text-sm'
                      />
                      <Button
                        type='button'
                        variant='ghost'
                        size='icon'
                        className='absolute right-0 top-0 h-full w-10'
                        onClick={() => setShowKey(!showKey)}
                      >
                        {showKey ? <EyeOff className='h-4 w-4' /> : <Eye className='h-4 w-4' />}
                      </Button>
                    </div>
                  </div>
                )}

                <div className='space-y-2'>
                  <div className='flex items-center justify-between'>
                    <Label htmlFor='provider-model'>Default Model</Label>
                    <Button
                      type='button'
                      variant='ghost'
                      size='sm'
                      className='h-6 text-xs gap-1'
                      onClick={handleFetchModels}
                      disabled={!canFetchModels || isModelFetching}
                    >
                      {isModelFetching ? (
                        <Loader2 className='h-3 w-3 animate-spin' />
                      ) : (
                        <RefreshCw className='h-3 w-3' />
                      )}
                      Fetch Models
                    </Button>
                  </div>
                  <Input
                    id='provider-model'
                    value={formData.default_model}
                    onChange={(e) =>
                      setFormData({ ...formData, default_model: e.target.value })
                    }
                    placeholder='model-name'
                    className='font-mono text-sm'
                  />
                  {fetchedModels.length > 0 ? (
                    <div className='max-h-32 overflow-y-auto'>
                      <div className='flex flex-wrap gap-1 pt-1'>
                        {fetchedModels.map(m => (
                          <Badge
                            key={m}
                            variant={formData.default_model === m ? 'default' : 'outline'}
                            className='text-xs cursor-pointer hover:bg-accent'
                            onClick={() => setFormData({ ...formData, default_model: m })}
                          >
                            {m}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  ) : selectedTemplate?.default_models && selectedTemplate.default_models.length > 0 ? (
                    <div className='flex flex-wrap gap-1 pt-1'>
                      {selectedTemplate.default_models.map(m => (
                        <Badge
                          key={m}
                          variant={formData.default_model === m ? 'default' : 'outline'}
                          className='text-xs cursor-pointer hover:bg-accent'
                          onClick={() => setFormData({ ...formData, default_model: m })}
                        >
                          {m}
                        </Badge>
                      ))}
                    </div>
                  ) : null}
                </div>

                {selectedTemplate && (
                  <div className='flex gap-2 text-xs text-muted-foreground'>
                    <span>Adapters:</span>
                    {selectedTemplate.agents?.map(a => (
                      <Badge key={a} variant='secondary' className='text-xs'>
                        {a}
                      </Badge>
                    ))}
                  </div>
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
                <Button type='submit' disabled={isPending}>
                  {isPending && <Loader2 className='mr-2 h-4 w-4 animate-spin' />}
                  {isEdit ? 'Save' : 'Create'}
                </Button>
              </DialogFooter>
            </form>
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}
