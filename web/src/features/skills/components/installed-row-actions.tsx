import { useNavigate } from '@tanstack/react-router'
import { DotsHorizontalIcon } from '@radix-ui/react-icons'
import { type Row } from '@tanstack/react-table'
import { Edit, Copy, Download, Power, PowerOff, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import type { Skill } from '@/types'
import { useUpdateSkill, useDeleteSkill, useCloneSkill, useExportSkill } from '@/hooks'

type Props = {
  row: Row<Skill>
}

export function InstalledRowActions({ row }: Props) {
  const navigate = useNavigate()
  const skill = row.original
  const updateSkill = useUpdateSkill()
  const deleteSkill = useDeleteSkill()
  const cloneSkill = useCloneSkill()
  const exportSkill = useExportSkill()

  const handleToggle = () => {
    updateSkill.mutate({ id: skill.id, data: { is_enabled: !skill.is_enabled } })
  }

  const handleClone = () => {
    cloneSkill.mutate({
      id: skill.id,
      newId: `${skill.id}-copy-${Date.now()}`,
      newName: `${skill.name} (Copy)`,
    })
  }

  const handleExport = () => {
    exportSkill.mutate(skill.id, {
      onSuccess: (content) => {
        const blob = new Blob([content], { type: 'text/markdown' })
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `${skill.id}-SKILL.md`
        a.click()
        URL.revokeObjectURL(url)
      },
    })
  }

  const handleDelete = () => {
    if (!confirm(`Delete skill "${skill.name}"?`)) return
    deleteSkill.mutate(skill.id)
  }

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger asChild>
        <Button
          variant='ghost'
          className='flex h-8 w-8 p-0 data-[state=open]:bg-muted'
        >
          <DotsHorizontalIcon className='h-4 w-4' />
          <span className='sr-only'>Open menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-[160px]'>
        <DropdownMenuItem onClick={() => navigate({ to: `/skills/${skill.id}` })}>
          Edit
          <DropdownMenuShortcut><Edit size={16} /></DropdownMenuShortcut>
        </DropdownMenuItem>
        <DropdownMenuItem onClick={handleToggle}>
          {skill.is_enabled ? 'Disable' : 'Enable'}
          <DropdownMenuShortcut>
            {skill.is_enabled ? <PowerOff size={16} /> : <Power size={16} />}
          </DropdownMenuShortcut>
        </DropdownMenuItem>
        <DropdownMenuItem onClick={handleClone}>
          Clone
          <DropdownMenuShortcut><Copy size={16} /></DropdownMenuShortcut>
        </DropdownMenuItem>
        <DropdownMenuItem onClick={handleExport}>
          Export
          <DropdownMenuShortcut><Download size={16} /></DropdownMenuShortcut>
        </DropdownMenuItem>
        {!skill.is_built_in && (
          <>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={handleDelete} className='text-red-500!'>
              Delete
              <DropdownMenuShortcut><Trash2 size={16} /></DropdownMenuShortcut>
            </DropdownMenuItem>
          </>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
