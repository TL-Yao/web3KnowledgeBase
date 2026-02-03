// frontend/components/ui/disabled-feature.tsx
'use client'

import { Construction } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

interface DisabledFeatureProps {
  /** 功能名称 */
  featureName: string
  /** 可选描述 */
  description?: string
  /** 显示样式 */
  variant?: 'banner' | 'card' | 'inline'
}

export function DisabledFeature({
  featureName,
  description,
  variant = 'banner'
}: DisabledFeatureProps) {
  if (variant === 'inline') {
    return (
      <span className="inline-flex items-center gap-1 text-muted-foreground text-sm">
        <Construction className="w-3 h-3" />
        <span>开发中</span>
      </span>
    )
  }

  if (variant === 'banner') {
    return (
      <div className="flex items-center gap-2 p-3 text-sm text-amber-600 bg-amber-50 rounded-lg border border-amber-200">
        <Construction className="w-4 h-4 shrink-0" />
        <span>
          <strong>{featureName}</strong> 功能正在开发中
          {description && <span className="text-amber-500">，{description}</span>}
        </span>
        <Badge variant="outline" className="ml-auto text-amber-600 border-amber-300">
          即将推出
        </Badge>
      </div>
    )
  }

  // card variant
  return (
    <Card className="border-dashed border-amber-300 bg-amber-50/50">
      <CardContent className="flex flex-col items-center justify-center py-12 text-center">
        <Construction className="w-12 h-12 text-amber-500 mb-4" />
        <h3 className="text-lg font-medium text-amber-700">{featureName}</h3>
        <p className="text-sm text-amber-600 mt-1">
          {description || '此功能正在开发中，敬请期待'}
        </p>
        <Badge variant="outline" className="mt-4 text-amber-600 border-amber-300">
          即将推出
        </Badge>
      </CardContent>
    </Card>
  )
}

/**
 * 用于包装可能禁用的按钮
 */
interface DisabledButtonWrapperProps {
  isDisabled: boolean
  featureName: string
  children: React.ReactNode
}

export function DisabledButtonWrapper({
  isDisabled,
  featureName,
  children
}: DisabledButtonWrapperProps) {
  if (!isDisabled) return <>{children}</>

  return (
    <div className="relative group">
      <div className="opacity-50 pointer-events-none">
        {children}
      </div>
      <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
        <Badge variant="outline" className="bg-background">
          {featureName} 开发中
        </Badge>
      </div>
    </div>
  )
}
