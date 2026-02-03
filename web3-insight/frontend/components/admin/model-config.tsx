'use client'

import { useState, useEffect, useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select'
import { Check, AlertCircle, RefreshCw } from 'lucide-react'
import { modelConfigAPI, type TaskSelection, type Model } from '@/lib/api'
import { toast } from 'sonner'

export function ModelConfig() {
  const queryClient = useQueryClient()
  const [selections, setSelections] = useState<TaskSelection[]>([])
  // Track initial selections for proper reset functionality
  const [initialSelections, setInitialSelections] = useState<TaskSelection[]>([])

  // Fetch model registry
  const { data: modelsRegistry, isLoading: loadingModels } = useQuery({
    queryKey: ['models-registry'],
    queryFn: modelConfigAPI.getModelsRegistry,
  })

  // Fetch task types
  const { data: routingConfig, isLoading: loadingTasks } = useQuery({
    queryKey: ['task-types'],
    queryFn: modelConfigAPI.getTaskTypes,
  })

  // Fetch user selections
  const { data: userSelections, isLoading: loadingSelections } = useQuery({
    queryKey: ['model-selections'],
    queryFn: modelConfigAPI.getUserSelections,
  })

  // Set selections when data loads or updates (after save)
  useEffect(() => {
    if (userSelections) {
      setSelections(userSelections)
      setInitialSelections(userSelections)
    }
  }, [userSelections])

  // Save mutation
  const saveMutation = useMutation({
    mutationFn: modelConfigAPI.updateUserSelections,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['model-selections'] })
      toast.success('模型配置已更新')
    },
    onError: () => {
      toast.error('保存失败，请稍后重试')
    },
  })

  const handleSave = () => {
    saveMutation.mutate(selections)
  }

  const updateSelection = (taskId: string, field: 'primary' | 'fallback', value: string) => {
    setSelections(prev =>
      prev.map(sel =>
        sel.taskId === taskId
          ? { ...sel, [field]: value }
          : sel
      )
    )
  }

  // Memoize allModels array to prevent recreation on every render
  const allModels = useMemo(() => {
    if (!modelsRegistry) return []
    return [...modelsRegistry.localModels, ...modelsRegistry.cloudModels]
  }, [modelsRegistry])

  const isModelAvailable = (modelId: string): boolean => {
    const model = allModels.find(m => m.id === modelId)
    return model ? model.enabled : false
  }

  const getModelDisplayName = (modelId: string): string => {
    if (allModels.length === 0) return modelId
    const model = allModels.find(m => m.id === modelId)
    return model ? model.name : modelId
  }

  const renderModelSelect = (
    taskId: string,
    currentValue: string,
    field: 'primary' | 'fallback',
    capability: string
  ) => {
    if (allModels.length === 0) return null

    const availableModels = allModels.filter(m =>
      m.capabilities.includes(capability)
    )

    const isCurrentAvailable = isModelAvailable(currentValue)

    return (
      <div className="flex items-center gap-2">
        <Select
          value={currentValue}
          onValueChange={(value) => updateSelection(taskId, field, value)}
        >
          <SelectTrigger className="h-8 w-[200px]">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {availableModels.map(model => (
              <SelectItem
                key={model.id}
                value={model.id}
                disabled={!model.enabled}
              >
                {model.name} {!model.enabled && '(已禁用)'}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        {!isCurrentAvailable && (
          <Badge variant="outline" className="text-yellow-600 border-yellow-600">
            <AlertCircle className="w-3 h-3 mr-1" />
            不可用
          </Badge>
        )}
      </div>
    )
  }

  if (loadingModels || loadingTasks || loadingSelections) {
    return <div className="flex items-center justify-center py-12">加载中...</div>
  }

  const hasUnavailableModels = selections.some(sel =>
    !isModelAvailable(sel.primary) || !isModelAvailable(sel.fallback)
  )

  return (
    <div className="space-y-6">
      {/* Warning Banner */}
      {hasUnavailableModels && (
        <Card className="border-yellow-500 bg-yellow-50">
          <CardContent className="pt-6">
            <div className="flex items-start gap-3">
              <AlertCircle className="w-5 h-5 text-yellow-600 mt-0.5" />
              <div>
                <p className="font-medium text-yellow-900">检测到模型配置问题</p>
                <p className="text-sm text-yellow-700 mt-1">
                  部分任务的首选模型不可用。系统将自动使用备用模型。建议更新配置。
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Local Models */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">本地模型 (Ollama)</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="border rounded-lg divide-y">
            {modelsRegistry?.localModels.map((model) => (
              <div key={model.id} className="flex items-center justify-between p-3">
                <div className="flex items-center gap-3">
                  <span className="font-mono text-sm">{model.name}</span>
                  <Badge variant={model.enabled ? 'default' : 'secondary'}>
                    {model.enabled ? '已启用' : '已禁用'}
                  </Badge>
                </div>
                <div className="flex items-center gap-4 text-sm text-muted-foreground">
                  <span className="text-xs">{model.capabilities.join(', ')}</span>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Cloud Models */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">云端模型</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="border rounded-lg divide-y">
            {modelsRegistry?.cloudModels.map((model) => (
              <div key={model.id} className="flex items-center justify-between p-3">
                <div className="flex items-center gap-3">
                  <span className="font-mono text-sm">{model.name}</span>
                  <Badge variant={model.enabled ? 'default' : 'secondary'}>
                    {model.enabled ? '已启用' : '已禁用'}
                  </Badge>
                </div>
                <div className="flex items-center gap-4 text-sm text-muted-foreground">
                  <span className="text-xs">${model.costPer1kTokens}/1K tokens</span>
                  <span className="text-xs">{model.capabilities.join(', ')}</span>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Task Routing */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">任务模型路由</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="border rounded-lg">
            <div className="grid grid-cols-4 gap-4 p-3 bg-muted text-sm font-medium">
              <span>任务类型</span>
              <span>首选模型</span>
              <span>备用模型</span>
              <span>状态</span>
            </div>
            {routingConfig?.taskTypes.map((task) => {
              const selection = selections.find(s => s.taskId === task.id)
              if (!selection) return null

              const primaryAvailable = isModelAvailable(selection.primary)
              const fallbackAvailable = isModelAvailable(selection.fallback)
              const bothUnavailable = !primaryAvailable && !fallbackAvailable

              return (
                <div key={task.id} className="grid grid-cols-4 gap-4 p-3 border-t text-sm items-center">
                  <div>
                    <div className="font-medium">{task.name}</div>
                    <div className="text-xs text-muted-foreground">{task.description}</div>
                  </div>

                  {renderModelSelect(task.id, selection.primary, 'primary', task.requiredCapability)}
                  {renderModelSelect(task.id, selection.fallback, 'fallback', task.requiredCapability)}

                  <div>
                    {bothUnavailable ? (
                      <Badge variant="destructive">
                        <AlertCircle className="w-3 h-3 mr-1" />
                        无可用模型
                      </Badge>
                    ) : !primaryAvailable ? (
                      <Badge variant="outline" className="text-yellow-600 border-yellow-600">
                        使用备用模型
                      </Badge>
                    ) : (
                      <Badge variant="outline" className="text-green-600 border-green-600">
                        ✓ 正常
                      </Badge>
                    )}
                  </div>
                </div>
              )
            })}
          </div>
        </CardContent>
      </Card>

      <div className="flex justify-end gap-2">
        <Button
          variant="outline"
          onClick={() => setSelections(initialSelections)}
        >
          <RefreshCw className="w-4 h-4 mr-2" />
          重置
        </Button>
        <Button
          onClick={handleSave}
          disabled={saveMutation.isPending}
        >
          <Check className="w-4 h-4 mr-2" />
          保存配置
        </Button>
      </div>
    </div>
  )
}
