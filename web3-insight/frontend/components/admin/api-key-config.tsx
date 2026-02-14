'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Eye, EyeOff, Check, X, Loader2, FlaskConical, Save } from 'lucide-react'
import { apiKeyAPI, type ApiKeyProvider } from '@/lib/api'
import { toast } from 'sonner'

interface ProviderRowState {
  inputValue: string
  showPassword: boolean
  dirty: boolean
}

export function ApiKeyConfig() {
  const queryClient = useQueryClient()
  const [rowState, setRowState] = useState<Record<string, ProviderRowState>>({})
  const [testingProvider, setTestingProvider] = useState<string | null>(null)
  const [testResults, setTestResults] = useState<Record<string, { success: boolean; message: string }>>({})
  const [savingProvider, setSavingProvider] = useState<string | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['api-keys'],
    queryFn: apiKeyAPI.list,
  })

  const saveMutation = useMutation({
    mutationFn: (params: { provider: string; key: string }) =>
      apiKeyAPI.save({ [params.provider]: params.key }),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['api-keys'] })
      setRowState(prev => ({
        ...prev,
        [variables.provider]: { ...prev[variables.provider], inputValue: '', dirty: false, showPassword: false },
      }))
      setTestResults(prev => {
        const next = { ...prev }
        delete next[variables.provider]
        return next
      })
      toast.success('API 密钥已保存')
      setSavingProvider(null)
    },
    onError: () => {
      toast.error('保存失败，请稍后重试')
      setSavingProvider(null)
    },
  })

  const testMutation = useMutation({
    mutationFn: (params: { provider: string; key?: string }) =>
      apiKeyAPI.test(params.provider, params.key),
    onSuccess: (result) => {
      setTestResults(prev => ({
        ...prev,
        [result.provider]: { success: result.success, message: result.message },
      }))
      setTestingProvider(null)
    },
    onError: (_err, variables) => {
      setTestResults(prev => ({
        ...prev,
        [variables.provider]: { success: false, message: '测试请求失败' },
      }))
      setTestingProvider(null)
    },
  })

  const getRowState = (providerId: string): ProviderRowState => {
    return rowState[providerId] || { inputValue: '', showPassword: false, dirty: false }
  }

  const updateRowState = (providerId: string, updates: Partial<ProviderRowState>) => {
    setRowState(prev => ({
      ...prev,
      [providerId]: { ...getRowState(providerId), ...updates },
    }))
  }

  const handleInputChange = (providerId: string, value: string) => {
    updateRowState(providerId, { inputValue: value, dirty: true })
    // Clear test result when input changes
    setTestResults(prev => {
      const next = { ...prev }
      delete next[providerId]
      return next
    })
  }

  const handleSave = (provider: ApiKeyProvider) => {
    const state = getRowState(provider.id)
    if (!state.dirty || !state.inputValue.trim()) return
    setSavingProvider(provider.id)
    saveMutation.mutate({ provider: provider.id, key: state.inputValue.trim() })
  }

  const handleRemove = (provider: ApiKeyProvider) => {
    if (!provider.configured) return
    setSavingProvider(provider.id)
    saveMutation.mutate({ provider: provider.id, key: '' })
  }

  const handleTest = (provider: ApiKeyProvider) => {
    const state = getRowState(provider.id)
    setTestingProvider(provider.id)
    // If user typed a new key, test that; otherwise test the stored key
    if (state.dirty && state.inputValue.trim()) {
      testMutation.mutate({ provider: provider.id, key: state.inputValue.trim() })
    } else {
      testMutation.mutate({ provider: provider.id })
    }
  }

  if (isLoading) {
    return <div className="flex items-center justify-center py-12">加载中...</div>
  }

  const providers = data?.providers || []

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">API 密钥配置</CardTitle>
          <p className="text-sm text-muted-foreground">
            管理各服务提供商的 API 密钥。密钥将安全存储在数据库中，修改后即时生效。
          </p>
        </CardHeader>
        <CardContent className="space-y-6">
          {providers.map((provider) => {
            const state = getRowState(provider.id)
            const testResult = testResults[provider.id]
            const isTesting = testingProvider === provider.id
            const isSaving = savingProvider === provider.id

            return (
              <div key={provider.id} className="space-y-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-sm">{provider.name}</span>
                    {provider.configured ? (
                      <Badge variant="outline" className="text-green-600 border-green-600">
                        已配置
                      </Badge>
                    ) : (
                      <Badge variant="outline" className="text-muted-foreground">
                        未配置
                      </Badge>
                    )}
                  </div>
                  {provider.configured && (
                    <Button
                      variant="ghost"
                      size="sm"
                      className="text-destructive hover:text-destructive h-7 text-xs"
                      onClick={() => handleRemove(provider)}
                      disabled={isSaving}
                    >
                      移除
                    </Button>
                  )}
                </div>

                <div className="flex items-center gap-2">
                  <div className="relative flex-1">
                    <Input
                      type={state.showPassword ? 'text' : 'password'}
                      placeholder={provider.configured ? provider.masked : '输入 API 密钥...'}
                      value={state.inputValue}
                      onChange={(e) => handleInputChange(provider.id, e.target.value)}
                      className="pr-9 font-mono text-sm"
                    />
                    <button
                      type="button"
                      className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                      onClick={() => updateRowState(provider.id, { showPassword: !state.showPassword })}
                      tabIndex={-1}
                    >
                      {state.showPassword ? (
                        <EyeOff className="w-4 h-4" />
                      ) : (
                        <Eye className="w-4 h-4" />
                      )}
                    </button>
                  </div>

                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handleTest(provider)}
                    disabled={isTesting || (!provider.configured && !state.dirty)}
                    className="shrink-0"
                  >
                    {isTesting ? (
                      <Loader2 className="w-4 h-4 mr-1 animate-spin" />
                    ) : (
                      <FlaskConical className="w-4 h-4 mr-1" />
                    )}
                    测试
                  </Button>

                  <Button
                    size="sm"
                    onClick={() => handleSave(provider)}
                    disabled={!state.dirty || !state.inputValue.trim() || isSaving}
                    className="shrink-0"
                  >
                    {isSaving ? (
                      <Loader2 className="w-4 h-4 mr-1 animate-spin" />
                    ) : (
                      <Save className="w-4 h-4 mr-1" />
                    )}
                    保存
                  </Button>
                </div>

                {testResult && (
                  <div className={`flex items-center gap-1.5 text-xs ${testResult.success ? 'text-green-600' : 'text-destructive'}`}>
                    {testResult.success ? (
                      <Check className="w-3.5 h-3.5" />
                    ) : (
                      <X className="w-3.5 h-3.5" />
                    )}
                    {testResult.message}
                  </div>
                )}
              </div>
            )
          })}
        </CardContent>
      </Card>
    </div>
  )
}
