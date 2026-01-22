import { useState, useEffect } from 'react'
import { Plus, Trash2, CheckCircle, XCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog'
import { toast } from 'sonner'
import { api } from '@/services/api'
import type { Webhook } from '@/types'

export function WebhookList() {
  const [webhooks, setWebhooks] = useState<Webhook[]>([])
  const [loading, setLoading] = useState(true)
  const [deleteId, setDeleteId] = useState<string | null>(null)

  useEffect(() => { loadWebhooks() }, [])

  const loadWebhooks = async () => {
    try {
      const list = await api.listWebhooks()
      setWebhooks(list)
    } catch (error) {
      console.error('Failed to load webhooks:', error)
      toast.error('Failed to load webhooks')
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async () => {
    if (!deleteId) return
    try {
      await api.deleteWebhook(deleteId)
      setWebhooks((prev) => prev.filter((w) => w.id !== deleteId))
      toast.success('Webhook deleted')
    } catch (error) {
      toast.error('Failed to delete webhook')
    } finally {
      setDeleteId(null)
    }
  }

  if (loading) {
    return <div className="flex items-center justify-center h-64"><div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" /></div>
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">Webhooks</h2>
          <p className="text-muted-foreground">Manage webhook endpoints for event notifications</p>
        </div>
        <Button onClick={() => toast.info('New webhook form coming soon')}><Plus className="mr-2 h-4 w-4" />New Webhook</Button>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>Configured Webhooks</CardTitle>
          <CardDescription>{webhooks.length} webhook(s)</CardDescription>
        </CardHeader>
        <CardContent>
          {webhooks.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">No webhooks configured.</div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>URL</TableHead>
                  <TableHead>Events</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {webhooks.map((w) => (
                  <TableRow key={w.id}>
                    <TableCell className="font-medium">{w.id}</TableCell>
                    <TableCell><span className="truncate max-w-[200px] block">{w.url}</span></TableCell>
                    <TableCell>
                      {w.events?.slice(0,2).map((e) => <Badge key={e} variant="outline" className="text-xs mr-1">{e}</Badge>)}
                      {(w.events?.length || 0) > 2 && <Badge variant="outline" className="text-xs">+{w.events!.length - 2}</Badge>}
                    </TableCell>
                    <TableCell>
                      {w.is_active ? <Badge className="bg-green-500"><CheckCircle className="mr-1 h-3 w-3"/>Active</Badge> : <Badge variant="secondary"><XCircle className="mr-1 h-3 w-3"/>Inactive</Badge>}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button variant="ghost" size="icon" onClick={() => setDeleteId(w.id)}><Trash2 className="h-4 w-4 text-destructive"/></Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
      <AlertDialog open={!!deleteId} onOpenChange={() => setDeleteId(null)}>
        <AlertDialogContent>
          <AlertDialogHeader><AlertDialogTitle>Delete Webhook?</AlertDialogTitle><AlertDialogDescription>This cannot be undone.</AlertDialogDescription></AlertDialogHeader>
          <AlertDialogFooter><AlertDialogCancel>Cancel</AlertDialogCancel><AlertDialogAction onClick={handleDelete} className="bg-destructive">Delete</AlertDialogAction></AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

export default WebhookList
