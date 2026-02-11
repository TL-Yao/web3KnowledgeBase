'use client'

import { useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useFeatureFlag } from '@/hooks/use-feature-flag'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
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
import { TagEditor } from '@/components/knowledge/tag-editor'
import { VersionHistory } from '@/components/knowledge/version-history'
import { Edit, RefreshCw, MoreHorizontal, ExternalLink, Clock, Archive, ArchiveRestore, Trash2 } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { Article, articleAPI } from '@/lib/api'

function stripLeadingTitle(content: string, title: string): string {
  const lines = content.split('\n')
  const firstContentIdx = lines.findIndex(line => line.trim().length > 0)
  if (firstContentIdx === -1) return content
  const firstLine = lines[firstContentIdx].trim()
  const match = firstLine.match(/^#\s+(.+)$/)
  if (match && match[1].trim() === title.trim()) {
    lines.splice(firstContentIdx, 1)
    return lines.join('\n').trimStart()
  }
  return content
}

interface ArticleViewProps {
  article: Article
}

export function ArticleView({ article }: ArticleViewProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const { isDisabled: regenerateDisabled } = useFeatureFlag('articleRegenerate')
  const router = useRouter()
  const queryClient = useQueryClient()

  const archiveMutation = useMutation({
    mutationFn: () => articleAPI.toggleArchive(article.id),
    onSuccess: (updated) => {
      queryClient.invalidateQueries({ queryKey: ['article'] })
      queryClient.invalidateQueries({ queryKey: ['articles'] })
      toast.success(updated.archived ? '文章已归档' : '文章已取消归档')
    },
    onError: () => {
      toast.error('操作失败')
    },
  })

  const deleteMutation = useMutation({
    mutationFn: () => articleAPI.delete(article.id),
    onSuccess: () => {
      toast.success('文章已删除')
      router.push('/knowledge')
    },
    onError: () => {
      toast.error('删除失败')
    },
  })

  if (!article) {
    return (
      <div className="max-w-4xl mx-auto p-6">
        <p className="text-muted-foreground">文章不存在或加载失败</p>
      </div>
    )
  }

  return (
    <div className="max-w-4xl mx-auto p-6">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-start justify-between mb-4">
          <div>
            <div className="flex items-center gap-2 text-sm text-muted-foreground mb-2">
              <span className="flex items-center gap-1">
                <Clock className="w-3 h-3" />
                {article.updatedAt && formatDistanceToNow(new Date(article.updatedAt), {
                  addSuffix: true,
                  locale: zhCN
                })}
              </span>
            </div>
            <div className="flex items-center gap-3">
              <h1 className="text-3xl font-bold">{article.title}</h1>
              {article.archived && (
                <Badge variant="secondary" className="bg-amber-100 text-amber-800">
                  已归档
                </Badge>
              )}
            </div>
          </div>

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon">
                <MoreHorizontal className="w-4 h-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => setIsEditing(true)}>
                <Edit className="w-4 h-4 mr-2" />
                编辑
              </DropdownMenuItem>
              <DropdownMenuItem disabled={regenerateDisabled}>
                <RefreshCw className="w-4 h-4 mr-2" />
                重新生成
                {regenerateDisabled && (
                  <span className="ml-auto text-xs text-muted-foreground">开发中</span>
                )}
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => archiveMutation.mutate()}
                disabled={archiveMutation.isPending}
              >
                {article.archived ? (
                  <>
                    <ArchiveRestore className="w-4 h-4 mr-2" />
                    取消归档
                  </>
                ) : (
                  <>
                    <Archive className="w-4 h-4 mr-2" />
                    归档
                  </>
                )}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                onClick={() => setShowDeleteDialog(true)}
                className="text-destructive focus:text-destructive"
              >
                <Trash2 className="w-4 h-4 mr-2" />
                删除
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        {/* Tags - editable */}
        <TagEditor
          articleId={article.id}
          initialTags={article.tags || []}
          onTagClick={(tag) => router.push(`/knowledge?tag=${encodeURIComponent(tag)}`)}
        />
      </div>

      {/* Content */}
      <article className="prose prose-neutral max-w-none dark:prose-invert">
        {article.contentHtml ? (
          <div dangerouslySetInnerHTML={{ __html: article.contentHtml }} />
        ) : (
          <ReactMarkdown remarkPlugins={[remarkGfm]}>
            {stripLeadingTitle(article.content || article.summary, article.title)}
          </ReactMarkdown>
        )}
      </article>

      {/* Sources */}
      {article.sourceUrls && article.sourceUrls.length > 0 && (
        <div className="mt-12 pt-6 border-t border-border">
          <h3 className="text-sm font-medium text-muted-foreground mb-3">原始来源</h3>
          <ul className="space-y-1">
            {article.sourceUrls.map((url) => (
              <li key={url}>
                <a
                  href={url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-primary hover:underline flex items-center gap-1"
                >
                  {url}
                  <ExternalLink className="w-3 h-3" />
                </a>
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Model info */}
      {article.modelUsed && (
        <div className="mt-6 text-xs text-muted-foreground">
          生成模型: {article.modelUsed}
        </div>
      )}

      {/* Version History */}
      <VersionHistory articleId={article.id} articleSlug={article.slug} />

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>
              确定要删除文章「{article.title}」吗？此操作无法撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => deleteMutation.mutate()}
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
