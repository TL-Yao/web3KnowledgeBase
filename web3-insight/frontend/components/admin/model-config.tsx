'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select'
import { Check, RefreshCw } from 'lucide-react'

export function ModelConfig() {
  const queryClient = useQueryClient()

  const { data: config, isLoading } = useQuery({
    queryKey: ['model-config'],
    queryFn: async () => {
      // TODO: Fetch from API
      return {
        defaultLocal: 'llama3:70b',
        claudeEnabled: true,
        openaiEnabled: false
      }
    }
  })

  const saveMutation = useMutation({
    mutationFn: async (newConfig: any) => {
      await fetch('/api/config', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ key: 'models', value: newConfig })
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['model-config'] })
    }
  })

  if (isLoading) return <div>加载中...</div>

  return (
    <div className="space-y-6">
      {/* Local Models */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">本地模型 (Ollama)</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center gap-4">
            <span className="text-sm w-24">默认模型</span>
            <Select defaultValue="llama3:70b">
              <SelectTrigger className="w-[200px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="llama3:70b">llama3:70b</SelectItem>
                <SelectItem value="qwen2.5:32b">qwen2.5:32b</SelectItem>
                <SelectItem value="mistral:7b">mistral:7b</SelectItem>
              </SelectContent>
            </Select>
            <Button variant="outline" size="sm">
              <RefreshCw className="w-4 h-4 mr-2" />
              测试连接
            </Button>
          </div>

          <div className="border rounded-lg divide-y">
            {['llama3:70b', 'qwen2.5:32b', 'mistral:7b'].map((model) => (
              <div key={model} className="flex items-center justify-between p-3">
                <div className="flex items-center gap-3">
                  <Switch defaultChecked={model !== 'mistral:7b'} />
                  <span className="font-mono text-sm">{model}</span>
                </div>
                <div className="flex items-center gap-4 text-sm text-muted-foreground">
                  <span>{model === 'llama3:70b' ? '40GB' : model === 'qwen2.5:32b' ? '20GB' : '4GB'}</span>
                  <Badge variant={model === 'llama3:70b' ? 'default' : 'secondary'}>
                    {model === 'llama3:70b' ? '已加载' : '未加载'}
                  </Badge>
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
        <CardContent className="space-y-6">
          {/* Claude */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="font-medium">Claude API</span>
              <Switch defaultChecked />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="text-sm text-muted-foreground">API Key</label>
                <Input type="password" defaultValue="sk-ant-••••••••" />
              </div>
              <div>
                <label className="text-sm text-muted-foreground">默认模型</label>
                <Select defaultValue="claude-sonnet-4-20250514">
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="claude-sonnet-4-20250514">claude-sonnet-4-20250514</SelectItem>
                    <SelectItem value="claude-haiku">claude-haiku</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="flex items-center gap-4">
              <span className="text-sm text-muted-foreground">月预算上限</span>
              <Input type="number" defaultValue={50} className="w-24" />
              <span className="text-sm text-muted-foreground">当前用量: $38.50 / $50.00</span>
            </div>
          </div>

          {/* OpenAI */}
          <div className="space-y-3 pt-4 border-t">
            <div className="flex items-center justify-between">
              <span className="font-medium">OpenAI API</span>
              <div className="flex items-center gap-2">
                <Badge variant="outline">未配置</Badge>
                <Switch />
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Routing */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">模型路由策略</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="border rounded-lg">
            <div className="grid grid-cols-3 gap-4 p-3 bg-muted text-sm font-medium">
              <span>任务类型</span>
              <span>首选模型</span>
              <span>Fallback</span>
            </div>
            {[
              { task: '内容生成 (简单)', primary: 'llama3:70b', fallback: 'claude-haiku' },
              { task: '内容生成 (复杂)', primary: 'claude-sonnet', fallback: '-' },
              { task: '摘要/分类', primary: 'qwen2.5:32b', fallback: 'claude-haiku' },
              { task: '问答对话', primary: 'llama3:70b', fallback: 'claude-sonnet' },
              { task: '翻译', primary: 'qwen2.5:32b', fallback: 'claude-haiku' },
            ].map((route) => (
              <div key={route.task} className="grid grid-cols-3 gap-4 p-3 border-t text-sm">
                <span>{route.task}</span>
                <Select defaultValue={route.primary}>
                  <SelectTrigger className="h-8">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="llama3:70b">llama3:70b</SelectItem>
                    <SelectItem value="qwen2.5:32b">qwen2.5:32b</SelectItem>
                    <SelectItem value="claude-sonnet">claude-sonnet</SelectItem>
                    <SelectItem value="claude-haiku">claude-haiku</SelectItem>
                  </SelectContent>
                </Select>
                <Select defaultValue={route.fallback}>
                  <SelectTrigger className="h-8">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="-">无</SelectItem>
                    <SelectItem value="claude-sonnet">claude-sonnet</SelectItem>
                    <SelectItem value="claude-haiku">claude-haiku</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      <div className="flex justify-end gap-2">
        <Button variant="outline">恢复默认</Button>
        <Button onClick={() => saveMutation.mutate(config)}>
          <Check className="w-4 h-4 mr-2" />
          保存配置
        </Button>
      </div>
    </div>
  )
}
