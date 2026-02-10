'use client'

import { useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { useFeatureFlag } from '@/hooks/use-feature-flag'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
import { Edit, RefreshCw, MoreHorizontal, ExternalLink, Clock } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { useRouter } from 'next/navigation'
import { Article } from '@/lib/api'

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
  const { isDisabled: regenerateDisabled } = useFeatureFlag('articleRegenerate')
  const router = useRouter()

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
              <span>{article.category?.name}</span>
              <span>·</span>
              <span className="flex items-center gap-1">
                <Clock className="w-3 h-3" />
                {article.updatedAt && formatDistanceToNow(new Date(article.updatedAt), {
                  addSuffix: true,
                  locale: zhCN
                })}
              </span>
            </div>
            <h1 className="text-3xl font-bold">{article.title}</h1>
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
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        {/* Tags */}
        <div className="flex items-center gap-2">
          {article.tags?.map((tag) => (
            <Badge
              key={tag}
              variant="secondary"
              className="cursor-pointer hover:bg-primary/20 transition-colors"
              onClick={() => router.push(`/knowledge?tag=${encodeURIComponent(tag)}`)}
            >
              {tag}
            </Badge>
          ))}
        </div>
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
    </div>
  )
}
