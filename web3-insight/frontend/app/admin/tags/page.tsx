'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
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
  Tag as TagIcon,
  Search,
  CheckCircle2,
  Clock,
  XCircle,
  Loader2,
  RefreshCw,
} from 'lucide-react'
import { tagAPI, Tag } from '@/lib/api'
import { toast } from 'sonner'

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

  const updateStatusMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: Tag['status'] }) =>
      tagAPI.updateStatus(id, status),
    onSuccess: () => {
      toast.success('标签状态已更新')
      queryClient.invalidateQueries({ queryKey: ['admin-tags'] })
      queryClient.invalidateQueries({ queryKey: ['admin-tag-stats'] })
    },
    onError: (error: Error) => {
      toast.error('更新失败', { description: error.message })
    },
  })

  const tags = tagsData?.tags ?? []

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">标签管理</h1>
        <Button
          variant="outline"
          size="sm"
          onClick={() => {
            queryClient.invalidateQueries({ queryKey: ['admin-tags'] })
            queryClient.invalidateQueries({ queryKey: ['admin-tag-stats'] })
          }}
        >
          <RefreshCw className="w-4 h-4 mr-2" />
          刷新
        </Button>
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
              <CardTitle className="text-sm font-medium text-muted-foreground">平均标签/文章</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.avgTagsPerArticle?.toFixed(1) ?? 0}</div>
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
                    <TableCell>{getStatusBadge(tag.status)}</TableCell>
                    <TableCell>{tag.suggestCount}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
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
              <p className="text-sm">标签会在文章生成时自动创建</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
