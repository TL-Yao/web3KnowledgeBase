'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import {
  Tag as TagIcon,
  Search,
  CheckCircle2,
  Clock,
  XCircle,
  Loader2,
  RefreshCw,
  Plus,
  Pencil,
  Trash2,
} from 'lucide-react'
import { tagAPI, configAPI, Tag, CreateTagRequest, UpdateTagRequest } from '@/lib/api'
import { Switch } from '@/components/ui/switch'
import { toast } from 'sonner'

const THEME_OPTIONS = [
  { value: '', label: '通用 (无主题)' },
  { value: 'web3_basics', label: 'Web3基础知识' },
  { value: 'defi_basics', label: 'DeFi基础知识' },
  { value: 'tradfi_basics', label: '传统金融基础知识' },
  { value: 'advanced_tech', label: '进阶技术深度解析' },
  { value: 'advanced_defi', label: 'DeFi进阶机制' },
  { value: 'crypto_history', label: '加密行业重大事件' },
  { value: 'notable_products', label: '知名产品与协议' },
  { value: 'notable_companies', label: '知名公司与组织' },
]

function getThemeLabel(themeId: string | null): string {
  if (!themeId) return '通用'
  const found = THEME_OPTIONS.find(t => t.value === themeId)
  return found ? found.label : themeId
}

function getStatusBadge(status: Tag['status']) {
  switch (status) {
    case 'active':
      return <Badge className="bg-green-500"><CheckCircle2 className="w-3 h-3 mr-1" />已激活</Badge>
    case 'pending':
      return <Badge variant="secondary"><Clock className="w-3 h-3 mr-1" />待审核</Badge>
    case 'deprecated':
      return <Badge variant="outline" className="text-muted-foreground"><XCircle className="w-3 h-3 mr-1" />已弃用</Badge>
    default:
      return <Badge variant="outline">{status}</Badge>
  }
}

export default function AdminTagsPage() {
  const queryClient = useQueryClient()
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [searchQuery, setSearchQuery] = useState('')

  // Dialog state
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingTag, setEditingTag] = useState<Tag | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<Tag | null>(null)

  // Form state
  const [formName, setFormName] = useState('')
  const [formNameEn, setFormNameEn] = useState('')
  const [formThemeId, setFormThemeId] = useState('')

  const { data: tagsData, isLoading } = useQuery({
    queryKey: ['admin-tags', statusFilter, searchQuery],
    queryFn: () => tagAPI.list({
      status: statusFilter === 'all' ? undefined : statusFilter,
      q: searchQuery || undefined,
      limit: 100,
    }),
    staleTime: 30000,
  })

  const { data: stats } = useQuery({
    queryKey: ['admin-tag-stats'],
    queryFn: tagAPI.getStats,
    staleTime: 60000,
  })

  const { data: autoTagConfig, isError: autoTagConfigError } = useQuery({
    queryKey: ['config', 'auto_tagging_enabled'],
    queryFn: () => configAPI.get('auto_tagging_enabled'),
    staleTime: 30000,
    retry: false,
  })

  const autoTagEnabled = autoTagConfigError ? true : autoTagConfig?.value !== 'false'

  function invalidateTagQueries() {
    queryClient.invalidateQueries({ queryKey: ['admin-tags'] })
    queryClient.invalidateQueries({ queryKey: ['admin-tag-stats'] })
    queryClient.invalidateQueries({ queryKey: ['tags', 'in-use'] })
    queryClient.invalidateQueries({ queryKey: ['articles'] })
  }

  const toggleAutoTagMutation = useMutation({
    mutationFn: (enabled: boolean) =>
      configAPI.set('auto_tagging_enabled', String(enabled), '是否在文章生成时自动打标签'),
    onSuccess: (_, enabled) => {
      toast.success(enabled ? '自动标签已开启' : '自动标签已关闭')
      queryClient.invalidateQueries({ queryKey: ['config', 'auto_tagging_enabled'] })
    },
    onError: (error: Error) => {
      toast.error('切换失败', { description: error.message })
    },
  })

  const updateStatusMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: Tag['status'] }) =>
      tagAPI.updateStatus(id, status),
    onSuccess: () => {
      toast.success('标签状态已更新')
      invalidateTagQueries()
    },
    onError: (error: Error) => {
      toast.error('更新失败', { description: error.message })
    },
  })

  const createMutation = useMutation({
    mutationFn: (data: CreateTagRequest) => tagAPI.create(data),
    onSuccess: () => {
      toast.success('标签已创建')
      setDialogOpen(false)
      invalidateTagQueries()
    },
    onError: (error: Error) => {
      if (error.message?.includes('already exists')) {
        toast.error('标签名已存在')
      } else {
        toast.error('创建失败')
      }
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateTagRequest }) =>
      tagAPI.update(id, data),
    onSuccess: () => {
      toast.success('标签已更新')
      setDialogOpen(false)
      invalidateTagQueries()
    },
    onError: (error: Error) => {
      if (error.message?.includes('already exists')) {
        toast.error('标签名已存在')
      } else {
        toast.error('更新失败')
      }
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => tagAPI.delete(id),
    onSuccess: (result) => {
      toast.success(`标签已删除，${result.articlesAffected} 篇文章已更新`)
      setDeleteTarget(null)
      invalidateTagQueries()
    },
    onError: (error: Error) => {
      toast.error('删除失败', { description: error.message })
    },
  })

  const openCreateDialog = () => {
    setEditingTag(null)
    setFormName('')
    setFormNameEn('')
    setFormThemeId('')
    setDialogOpen(true)
  }

  const openEditDialog = (tag: Tag) => {
    setEditingTag(tag)
    setFormName(tag.name)
    setFormNameEn(tag.nameEn)
    setFormThemeId(tag.themeId || '')
    setDialogOpen(true)
  }

  const handleSave = () => {
    const name = formName.trim()
    if (!name) return

    if (editingTag) {
      const data: UpdateTagRequest = {}
      if (name !== editingTag.name) data.name = name
      if (formNameEn.trim() !== editingTag.nameEn) data.nameEn = formNameEn.trim()
      const newThemeId = formThemeId || null
      if (newThemeId !== editingTag.themeId) data.themeId = newThemeId
      updateMutation.mutate({ id: editingTag.id, data })
    } else {
      createMutation.mutate({
        name,
        nameEn: formNameEn.trim() || undefined,
        themeId: formThemeId || null,
      })
    }
  }

  const tags = tagsData?.tags ?? []
  const isSaving = createMutation.isPending || updateMutation.isPending

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">标签管理</h1>
        <div className="flex items-center gap-4">
          <Button size="sm" onClick={openCreateDialog}>
            <Plus className="w-4 h-4 mr-2" />
            新建标签
          </Button>
          <div className="flex items-center gap-2">
            <Switch
              id="auto-tag"
              checked={autoTagEnabled}
              onCheckedChange={(checked) => toggleAutoTagMutation.mutate(checked)}
              disabled={toggleAutoTagMutation.isPending}
            />
            <Label htmlFor="auto-tag" className="text-sm cursor-pointer">
              自动标签
            </Label>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={() => invalidateTagQueries()}
          >
            <RefreshCw className="w-4 h-4 mr-2" />
            刷新
          </Button>
        </div>
      </div>

      {/* Stats */}
      {stats && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">总标签数</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.totalTags}</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">已激活</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-green-600">{stats.activeTags}</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">待审核</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-amber-600">{stats.pendingTags}</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">通用标签</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.universalTags}</div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Filters */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">标签列表</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-4 mb-4">
            <div className="relative flex-1 max-w-sm">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input
                placeholder="搜索标签..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-9"
              />
            </div>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-32">
                <SelectValue placeholder="状态" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部</SelectItem>
                <SelectItem value="active">已激活</SelectItem>
                <SelectItem value="pending">待审核</SelectItem>
                <SelectItem value="deprecated">已弃用</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {isLoading ? (
            <div className="flex items-center justify-center py-8 text-muted-foreground">
              <Loader2 className="w-6 h-6 animate-spin mr-2" />
              加载中...
            </div>
          ) : tags.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>标签</TableHead>
                  <TableHead>英文名</TableHead>
                  <TableHead>主题</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>推荐次数</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {tags.map((tag) => (
                  <TableRow key={tag.id}>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <TagIcon className="w-4 h-4 text-muted-foreground" />
                        <span className="font-medium">{tag.name}</span>
                      </div>
                    </TableCell>
                    <TableCell className="text-muted-foreground">{tag.nameEn}</TableCell>
                    <TableCell className="text-muted-foreground text-sm">
                      {getThemeLabel(tag.themeId)}
                    </TableCell>
                    <TableCell>{getStatusBadge(tag.status)}</TableCell>
                    <TableCell>{tag.suggestCount}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => openEditDialog(tag)}
                        >
                          <Pencil className="w-3.5 h-3.5" />
                        </Button>
                        {tag.status === 'pending' && (
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => updateStatusMutation.mutate({ id: tag.id, status: 'active' })}
                            disabled={updateStatusMutation.isPending}
                          >
                            激活
                          </Button>
                        )}
                        {tag.status === 'active' && (
                          <Button
                            size="sm"
                            variant="outline"
                            className="text-muted-foreground"
                            onClick={() => updateStatusMutation.mutate({ id: tag.id, status: 'deprecated' })}
                            disabled={updateStatusMutation.isPending}
                          >
                            弃用
                          </Button>
                        )}
                        {tag.status === 'deprecated' && (
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => updateStatusMutation.mutate({ id: tag.id, status: 'active' })}
                            disabled={updateStatusMutation.isPending}
                          >
                            重新激活
                          </Button>
                        )}
                        <Button
                          size="sm"
                          variant="ghost"
                          className="text-destructive hover:text-destructive"
                          onClick={() => setDeleteTarget(tag)}
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : (
            <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
              <TagIcon className="w-12 h-12 mb-4 opacity-50" />
              <p className="text-lg font-medium">暂无标签</p>
              <p className="text-sm">点击"新建标签"添加标签</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Create/Edit Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingTag ? '编辑标签' : '新建标签'}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label htmlFor="tag-name">标签名称 *</Label>
              <Input
                id="tag-name"
                value={formName}
                onChange={(e) => setFormName(e.target.value)}
                placeholder="例：DeFi协议"
                maxLength={100}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="tag-name-en">英文名称</Label>
              <Input
                id="tag-name-en"
                value={formNameEn}
                onChange={(e) => setFormNameEn(e.target.value)}
                placeholder="例：DeFi Protocol"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="tag-theme">所属主题</Label>
              <Select
                value={formThemeId || '_universal'}
                onValueChange={(v) => setFormThemeId(v === '_universal' ? '' : v)}
              >
                <SelectTrigger id="tag-theme">
                  <SelectValue placeholder="通用 (无主题)" />
                </SelectTrigger>
                <SelectContent>
                  {THEME_OPTIONS.map((opt) => (
                    <SelectItem key={opt.value || '_universal'} value={opt.value || '_universal'}>
                      {opt.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              取消
            </Button>
            <Button
              onClick={handleSave}
              disabled={!formName.trim() || isSaving}
            >
              {isSaving ? '保存中...' : editingTag ? '保存' : '创建'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation */}
      <AlertDialog open={!!deleteTarget} onOpenChange={(open) => !open && setDeleteTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>
              确定要删除标签「{deleteTarget?.name}」吗？该标签将从所有文章中移除。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => deleteTarget && deleteMutation.mutate(deleteTarget.id)}
              className="bg-destructive text-white hover:bg-destructive/90"
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? '删除中...' : '删除'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
