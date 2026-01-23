import { useRef, useState } from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { ThemeSwitch } from '@/components/theme-switch'
import { useFiles, useUploadFile, useDeleteFile } from '@/hooks/useFiles'
import { api } from '@/services/api'
import {
  Loader2,
  Upload,
  Download,
  Trash2,
  FileText,
  FileImage,
  FileCode,
  FileArchive,
  File,
} from 'lucide-react'
import type { UploadedFile } from '@/types'

function getFileIcon(mimeType: string) {
  if (mimeType.startsWith('image/')) return <FileImage className='h-4 w-4 text-purple-500' />
  if (mimeType.startsWith('text/') || mimeType.includes('json') || mimeType.includes('xml'))
    return <FileCode className='h-4 w-4 text-blue-500' />
  if (mimeType.includes('zip') || mimeType.includes('gzip') || mimeType.includes('tar'))
    return <FileArchive className='h-4 w-4 text-orange-500' />
  if (mimeType.includes('pdf')) return <FileText className='h-4 w-4 text-red-500' />
  return <File className='h-4 w-4 text-gray-500' />
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMin = Math.floor(diffMs / 60000)
  if (diffMin < 1) return 'just now'
  if (diffMin < 60) return `${diffMin}m ago`
  const diffHour = Math.floor(diffMin / 60)
  if (diffHour < 24) return `${diffHour}h ago`
  return date.toLocaleDateString()
}

export function Files() {
  const { data: files, isLoading } = useFiles()
  const uploadFile = useUploadFile()
  const deleteFile = useDeleteFile()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [dragOver, setDragOver] = useState(false)

  const handleUpload = (fileList: FileList | null) => {
    if (!fileList) return
    Array.from(fileList).forEach(file => {
      uploadFile.mutate(file)
    })
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    setDragOver(false)
    handleUpload(e.dataTransfer.files)
  }

  return (
    <>
      <Header fixed>
        <div className='ms-auto flex items-center space-x-4'>
          <ThemeSwitch />
          <ProfileDropdown />
        </div>
      </Header>

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className='flex items-center justify-between'>
          <div>
            <h2 className='text-2xl font-bold tracking-tight'>Files</h2>
            <p className='text-muted-foreground'>
              Upload files as task attachments
            </p>
          </div>
          <Button
            onClick={() => fileInputRef.current?.click()}
            disabled={uploadFile.isPending}
          >
            {uploadFile.isPending ? (
              <Loader2 className='h-4 w-4 mr-2 animate-spin' />
            ) : (
              <Upload className='h-4 w-4 mr-2' />
            )}
            Upload
          </Button>
          <input
            ref={fileInputRef}
            type='file'
            multiple
            className='hidden'
            onChange={(e) => handleUpload(e.target.files)}
          />
        </div>

        {/* Drop zone */}
        <div
          className={`rounded-lg border-2 border-dashed p-8 text-center transition-colors ${
            dragOver ? 'border-primary bg-primary/5' : 'border-muted-foreground/25'
          }`}
          onDragOver={(e) => { e.preventDefault(); setDragOver(true) }}
          onDragLeave={() => setDragOver(false)}
          onDrop={handleDrop}
        >
          <Upload className='h-8 w-8 mx-auto text-muted-foreground/50 mb-2' />
          <p className='text-sm text-muted-foreground'>
            Drag & drop files here, or click Upload
          </p>
          <p className='text-xs text-muted-foreground/70 mt-1'>
            Max 100MB per file
          </p>
        </div>

        {/* File list */}
        {isLoading ? (
          <div className='flex items-center justify-center py-12'>
            <Loader2 className='h-6 w-6 animate-spin text-muted-foreground' />
          </div>
        ) : !files || files.length === 0 ? (
          <div className='flex flex-col items-center justify-center py-12 text-muted-foreground'>
            <FileText className='h-12 w-12 mb-3 opacity-30' />
            <p>No files uploaded yet</p>
          </div>
        ) : (
          <div className='rounded-md border'>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>File</TableHead>
                  <TableHead>Size</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Uploaded</TableHead>
                  <TableHead>ID</TableHead>
                  <TableHead className='text-right'>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {files.map((file: UploadedFile) => (
                  <TableRow key={file.id}>
                    <TableCell>
                      <div className='flex items-center gap-2'>
                        {getFileIcon(file.mime_type)}
                        <span className='font-medium text-sm'>{file.name}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <span className='text-sm text-muted-foreground'>
                        {formatFileSize(file.size)}
                      </span>
                    </TableCell>
                    <TableCell>
                      <Badge variant='secondary' className='text-xs font-mono'>
                        {file.mime_type}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <span className='text-sm text-muted-foreground'>
                        {formatTime(file.uploaded_at)}
                      </span>
                    </TableCell>
                    <TableCell>
                      <code className='text-xs text-muted-foreground font-mono'>
                        {file.id.slice(0, 8)}...
                      </code>
                    </TableCell>
                    <TableCell className='text-right'>
                      <div className='flex items-center justify-end gap-1'>
                        <Button
                          variant='ghost'
                          size='icon'
                          className='h-7 w-7'
                          title='Download'
                          asChild
                        >
                          <a href={api.getFileDownloadUrl(file.id)} download={file.name}>
                            <Download className='h-3.5 w-3.5' />
                          </a>
                        </Button>
                        <Button
                          variant='ghost'
                          size='icon'
                          className='h-7 w-7 text-destructive hover:text-destructive'
                          title='Delete'
                          onClick={() => deleteFile.mutate(file.id)}
                          disabled={deleteFile.isPending}
                        >
                          <Trash2 className='h-3.5 w-3.5' />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </Main>
    </>
  )
}
