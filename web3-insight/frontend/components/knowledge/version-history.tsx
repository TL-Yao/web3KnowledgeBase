'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
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
import { History, ChevronDown, ChevronUp, RotateCcw, Loader2 } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { toast } from 'sonner'
import { articleAPI, ArticleVersion } from '@/lib/api'

interface VersionHistoryProps {
  articleId: string
  articleSlug: string
}

function getEditedByLabel(editedBy: string): string {
  switch (editedBy) {
    case 'chat_refinement': return 'AI对话更新'
    case 'rollback': return '版本回滚'
    case 'manual_edit': return '手动编辑'
    default: return editedBy
  }
}

export function VersionHistory({ articleId, articleSlug }: VersionHistoryProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  const [rollbackTarget, setRollbackTarget] = useState<ArticleVersion | null>(null)
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['article-versions', articleId],
    queryFn: () => articleAPI.listVersions(articleId),
    enabled: isExpanded,
    staleTime: 30000,
  })

  const rollbackMutation = useMutation({
    mutationFn: (versionId: string) => articleAPI.rollback(articleId, versionId),
    onSuccess: (result) => {
      toast.success(result.message || '版本已回滚')
      setRollbackTarget(null)
      queryClient.invalidateQueries({ queryKey: ['article', articleSlug] })
      queryClient.invalidateQueries({ queryKey: ['article-versions', articleId] })
    },
    onError: () => {
      toast.error('回滚失败')
    },
  })

  const versions = data?.versions ?? []

  return (
    <div className="mt-8 border-t border-border pt-6">
      <Button
        variant="ghost"
        size="sm"
        className="text-muted-foreground hover:text-foreground"
        onClick={() => setIsExpanded(prev => !prev)}
      >
        <History className="w-4 h-4 mr-2" />
        版本历史
        {isExpanded ? (
          <ChevronUp className="w-4 h-4 ml-1" />
        ) : (
          <ChevronDown className="w-4 h-4 ml-1" />
        )}
      </Button>

      {isExpanded && (
        <div className="mt-3">
          {isLoading ? (
            <div className="flex items-center gap-2 text-sm text-muted-foreground py-4">
              <Loader2 className="w-4 h-4 animate-spin" />
              加载版本历史...
            </div>
          ) : versions.length === 0 ? (
            <p className="text-sm text-muted-foreground py-4">暂无版本历史</p>
          ) : (
            <div className="space-y-2">
              {versions.map((version) => (
                <div
                  key={version.id}
                  className="flex items-center justify-between py-2 px-3 rounded-md border border-border text-sm"
                >
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-xs text-muted-foreground">
                        {formatDistanceToNow(new Date(version.createdAt), {
                          addSuffix: true,
                          locale: zhCN,
                        })}
                      </span>
                      <span className="text-xs px-1.5 py-0.5 rounded bg-muted text-muted-foreground">
                        {getEditedByLabel(version.editedBy)}
                      </span>
                    </div>
                    <p className="text-sm mt-0.5 truncate">{version.changeSummary}</p>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="ml-3 text-xs shrink-0"
                    onClick={() => setRollbackTarget(version)}
                  >
                    <RotateCcw className="w-3 h-3 mr-1" />
                    回滚
                  </Button>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Rollback confirmation */}
      <AlertDialog open={!!rollbackTarget} onOpenChange={(open) => !open && setRollbackTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认回滚</AlertDialogTitle>
            <AlertDialogDescription>
              确定要回滚到此版本吗？当前内容将被保存为新版本快照，然后恢复到选中版本的内容。
              {rollbackTarget && (
                <span className="block mt-2 font-medium">
                  {rollbackTarget.changeSummary}
                </span>
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => rollbackTarget && rollbackMutation.mutate(rollbackTarget.id)}
              disabled={rollbackMutation.isPending}
            >
              {rollbackMutation.isPending ? '回滚中...' : '确认回滚'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
