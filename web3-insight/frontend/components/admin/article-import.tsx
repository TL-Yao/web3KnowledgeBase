'use client'

import { useState, useRef, useCallback } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import {
  importAPI,
  ImportBatch,
  ImportResult,
  ValidationResult,
  categoryAPI,
  type Category,
} from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  AlertCircle,
  CheckCircle,
  Upload,
  Download,
  FileJson,
  Loader2,
  AlertTriangle,
} from 'lucide-react'

export function ArticleImport() {
  const [jsonContent, setJsonContent] = useState('')
  const [importResult, setImportResult] = useState<ImportResult | null>(null)
  const [validationResult, setValidationResult] = useState<ValidationResult | null>(null)
  const [options, setOptions] = useState({
    skipDuplicates: true,
    updateExisting: false,
  })
  const [exportCategory, setExportCategory] = useState<string>('')
  const [exportStatus, setExportStatus] = useState<string>('')
  const fileInputRef = useRef<HTMLInputElement>(null)

  const { data: categories } = useQuery({
    queryKey: ['categories'],
    queryFn: categoryAPI.list,
  })

  const validateMutation = useMutation({
    mutationFn: (batch: ImportBatch) => importAPI.validate(batch),
    onSuccess: (result) => {
      setValidationResult(result)
    },
  })

  const importMutation = useMutation({
    mutationFn: (batch: ImportBatch) => importAPI.import(batch),
    onSuccess: (result) => {
      setImportResult(result)
      setValidationResult(null)
    },
  })

  const uploadMutation = useMutation({
    mutationFn: (file: File) => importAPI.uploadFile(file, options),
    onSuccess: (result) => {
      setImportResult(result)
      setValidationResult(null)
    },
  })

  const parseAndValidate = useCallback(() => {
    if (!jsonContent.trim()) return

    try {
      const data = JSON.parse(jsonContent)
      const batch: ImportBatch = {
        articles: Array.isArray(data) ? data : data.articles || [data],
        options,
      }
      validateMutation.mutate(batch)
    } catch {
      setValidationResult({
        valid: false,
        errors: [{ index: 0, title: '', message: 'JSON 格式无效' }],
        errorCount: 1,
        totalCount: 0,
      })
    }
  }, [jsonContent, options, validateMutation])

  const handleImport = useCallback(() => {
    if (!jsonContent.trim()) return

    try {
      const data = JSON.parse(jsonContent)
      const batch: ImportBatch = {
        articles: Array.isArray(data) ? data : data.articles || [data],
        options,
      }
      importMutation.mutate(batch)
    } catch {
      // JSON parsing error handled in validation
    }
  }, [jsonContent, options, importMutation])

  const handleFileUpload = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    if (file.type === 'application/json' || file.name.endsWith('.json')) {
      // Read file content for preview
      const reader = new FileReader()
      reader.onload = (event) => {
        const content = event.target?.result as string
        setJsonContent(content)
      }
      reader.readAsText(file)

      // Upload file
      uploadMutation.mutate(file)
    } else {
      setValidationResult({
        valid: false,
        errors: [{ index: 0, title: file.name, message: '仅支持 JSON 文件' }],
        errorCount: 1,
        totalCount: 0,
      })
    }

    // Reset file input
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }, [uploadMutation])

  const handleDownloadTemplate = () => {
    window.open(importAPI.getTemplate(), '_blank')
  }

  const handleExport = () => {
    const url = importAPI.export(
      exportCategory || undefined,
      exportStatus || undefined
    )
    window.open(url, '_blank')
  }

  const isLoading = validateMutation.isPending || importMutation.isPending || uploadMutation.isPending

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileJson className="h-5 w-5" />
            文章导入/导出
          </CardTitle>
          <CardDescription>
            从 JSON 导入文章或导出现有文章
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="import" className="space-y-4">
            <TabsList>
              <TabsTrigger value="import">导入</TabsTrigger>
              <TabsTrigger value="export">导出</TabsTrigger>
            </TabsList>

            <TabsContent value="import" className="space-y-4">
              {/* Import Options */}
              <div className="flex flex-wrap gap-4">
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="skipDuplicates"
                    checked={options.skipDuplicates}
                    onCheckedChange={(checked) =>
                      setOptions({ ...options, skipDuplicates: checked as boolean })
                    }
                  />
                  <Label htmlFor="skipDuplicates">跳过重复</Label>
                </div>
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="updateExisting"
                    checked={options.updateExisting}
                    onCheckedChange={(checked) =>
                      setOptions({ ...options, updateExisting: checked as boolean })
                    }
                  />
                  <Label htmlFor="updateExisting">更新已有文章</Label>
                </div>
              </div>

              {/* File Upload */}
              <div className="border-2 border-dashed border-muted rounded-lg p-6 text-center">
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".json"
                  onChange={handleFileUpload}
                  className="hidden"
                  id="file-upload"
                />
                <label
                  htmlFor="file-upload"
                  className="cursor-pointer flex flex-col items-center gap-2"
                >
                  <Upload className="h-8 w-8 text-muted-foreground" />
                  <span className="text-sm text-muted-foreground">
                    点击上传或拖拽 JSON 文件
                  </span>
                  <span className="text-xs text-muted-foreground">
                    最大文件大小: 10MB
                  </span>
                </label>
              </div>

              {/* JSON Input */}
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="json-content">或粘贴 JSON 内容</Label>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={handleDownloadTemplate}
                  >
                    <Download className="h-4 w-4 mr-2" />
                    下载模板
                  </Button>
                </div>
                <Textarea
                  id="json-content"
                  value={jsonContent}
                  onChange={(e) => setJsonContent(e.target.value)}
                  placeholder='[{"title": "Article Title", "content": "Markdown content..."}]'
                  className="font-mono text-sm h-48"
                />
              </div>

              {/* Action Buttons */}
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  onClick={parseAndValidate}
                  disabled={!jsonContent.trim() || isLoading}
                >
                  {validateMutation.isPending ? (
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  ) : (
                    <AlertCircle className="h-4 w-4 mr-2" />
                  )}
                  验证
                </Button>
                <Button
                  onClick={handleImport}
                  disabled={!jsonContent.trim() || isLoading}
                >
                  {importMutation.isPending ? (
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  ) : (
                    <Upload className="h-4 w-4 mr-2" />
                  )}
                  导入
                </Button>
              </div>

              {/* Validation Result */}
              {validationResult && (
                <Card className={validationResult.valid ? 'border-green-500' : 'border-red-500'}>
                  <CardContent className="pt-4">
                    <div className="flex items-center gap-2 mb-2">
                      {validationResult.valid ? (
                        <>
                          <CheckCircle className="h-5 w-5 text-green-500" />
                          <span className="font-medium text-green-700">
                            验证通过
                          </span>
                        </>
                      ) : (
                        <>
                          <AlertCircle className="h-5 w-5 text-red-500" />
                          <span className="font-medium text-red-700">
                            验证失败
                          </span>
                        </>
                      )}
                      <Badge variant="outline">
                        {validationResult.totalCount} 篇文章
                      </Badge>
                    </div>
                    {validationResult.errors.length > 0 && (
                      <div className="mt-2 space-y-1">
                        {validationResult.errors.map((error, idx) => (
                          <div
                            key={idx}
                            className="text-sm text-red-600 flex items-start gap-2"
                          >
                            <AlertTriangle className="h-4 w-4 mt-0.5 flex-shrink-0" />
                            <span>
                              [{error.index}] {error.title}: {error.message}
                            </span>
                          </div>
                        ))}
                      </div>
                    )}
                  </CardContent>
                </Card>
              )}

              {/* Import Result */}
              {importResult && (
                <Card className="border-blue-500">
                  <CardContent className="pt-4">
                    <div className="flex items-center gap-2 mb-4">
                      <CheckCircle className="h-5 w-5 text-blue-500" />
                      <span className="font-medium">导入完成</span>
                    </div>
                    <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 text-center">
                      <div>
                        <div className="text-2xl font-bold text-green-600">
                          {importResult.importedCount}
                        </div>
                        <div className="text-sm text-muted-foreground">已导入</div>
                      </div>
                      <div>
                        <div className="text-2xl font-bold text-blue-600">
                          {importResult.updatedCount}
                        </div>
                        <div className="text-sm text-muted-foreground">已更新</div>
                      </div>
                      <div>
                        <div className="text-2xl font-bold text-yellow-600">
                          {importResult.skippedCount}
                        </div>
                        <div className="text-sm text-muted-foreground">已跳过</div>
                      </div>
                      <div>
                        <div className="text-2xl font-bold text-red-600">
                          {importResult.errorCount}
                        </div>
                        <div className="text-sm text-muted-foreground">错误</div>
                      </div>
                    </div>

                    {importResult.errors && importResult.errors.length > 0 && (
                      <div className="mt-4">
                        <h4 className="font-medium mb-2">错误详情:</h4>
                        <Table>
                          <TableHeader>
                            <TableRow>
                              <TableHead>序号</TableHead>
                              <TableHead>标题</TableHead>
                              <TableHead>错误信息</TableHead>
                            </TableRow>
                          </TableHeader>
                          <TableBody>
                            {importResult.errors.map((error, idx) => (
                              <TableRow key={idx}>
                                <TableCell>{error.index}</TableCell>
                                <TableCell className="max-w-[200px] truncate">
                                  {error.title}
                                </TableCell>
                                <TableCell className="text-red-600">
                                  {error.message}
                                </TableCell>
                              </TableRow>
                            ))}
                          </TableBody>
                        </Table>
                      </div>
                    )}
                  </CardContent>
                </Card>
              )}
            </TabsContent>

            <TabsContent value="export" className="space-y-4">
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-2">
                  <Label>按分类筛选</Label>
                  <Select value={exportCategory} onValueChange={setExportCategory}>
                    <SelectTrigger>
                      <SelectValue placeholder="全部分类" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="">全部分类</SelectItem>
                      {categories?.map((cat: Category) => (
                        <SelectItem key={cat.id} value={cat.id}>
                          {cat.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label>按状态筛选</Label>
                  <Select value={exportStatus} onValueChange={setExportStatus}>
                    <SelectTrigger>
                      <SelectValue placeholder="全部状态" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="">全部状态</SelectItem>
                      <SelectItem value="published">已发布</SelectItem>
                      <SelectItem value="draft">草稿</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <Button onClick={handleExport}>
                <Download className="h-4 w-4 mr-2" />
                导出文章
              </Button>

              <p className="text-sm text-muted-foreground">
                以 JSON 格式导出文章，可用于备份或在不同环境之间迁移数据。
              </p>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  )
}
