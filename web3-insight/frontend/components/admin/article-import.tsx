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
        errors: [{ index: 0, title: '', message: 'Invalid JSON format' }],
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
        errors: [{ index: 0, title: file.name, message: 'Only JSON files are supported' }],
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
            Article Import/Export
          </CardTitle>
          <CardDescription>
            Import articles from JSON or export existing articles
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="import" className="space-y-4">
            <TabsList>
              <TabsTrigger value="import">Import</TabsTrigger>
              <TabsTrigger value="export">Export</TabsTrigger>
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
                  <Label htmlFor="skipDuplicates">Skip duplicates</Label>
                </div>
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="updateExisting"
                    checked={options.updateExisting}
                    onCheckedChange={(checked) =>
                      setOptions({ ...options, updateExisting: checked as boolean })
                    }
                  />
                  <Label htmlFor="updateExisting">Update existing</Label>
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
                    Click to upload or drag and drop a JSON file
                  </span>
                  <span className="text-xs text-muted-foreground">
                    Max file size: 10MB
                  </span>
                </label>
              </div>

              {/* JSON Input */}
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="json-content">Or paste JSON content</Label>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={handleDownloadTemplate}
                  >
                    <Download className="h-4 w-4 mr-2" />
                    Download Template
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
                  Validate
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
                  Import
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
                            Validation passed
                          </span>
                        </>
                      ) : (
                        <>
                          <AlertCircle className="h-5 w-5 text-red-500" />
                          <span className="font-medium text-red-700">
                            Validation failed
                          </span>
                        </>
                      )}
                      <Badge variant="outline">
                        {validationResult.totalCount} articles
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
                      <span className="font-medium">Import completed</span>
                    </div>
                    <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 text-center">
                      <div>
                        <div className="text-2xl font-bold text-green-600">
                          {importResult.importedCount}
                        </div>
                        <div className="text-sm text-muted-foreground">Imported</div>
                      </div>
                      <div>
                        <div className="text-2xl font-bold text-blue-600">
                          {importResult.updatedCount}
                        </div>
                        <div className="text-sm text-muted-foreground">Updated</div>
                      </div>
                      <div>
                        <div className="text-2xl font-bold text-yellow-600">
                          {importResult.skippedCount}
                        </div>
                        <div className="text-sm text-muted-foreground">Skipped</div>
                      </div>
                      <div>
                        <div className="text-2xl font-bold text-red-600">
                          {importResult.errorCount}
                        </div>
                        <div className="text-sm text-muted-foreground">Errors</div>
                      </div>
                    </div>

                    {importResult.errors && importResult.errors.length > 0 && (
                      <div className="mt-4">
                        <h4 className="font-medium mb-2">Errors:</h4>
                        <Table>
                          <TableHeader>
                            <TableRow>
                              <TableHead>Index</TableHead>
                              <TableHead>Title</TableHead>
                              <TableHead>Error</TableHead>
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
                  <Label>Filter by Category</Label>
                  <Select value={exportCategory} onValueChange={setExportCategory}>
                    <SelectTrigger>
                      <SelectValue placeholder="All categories" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="">All categories</SelectItem>
                      {categories?.map((cat: Category) => (
                        <SelectItem key={cat.id} value={cat.id}>
                          {cat.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label>Filter by Status</Label>
                  <Select value={exportStatus} onValueChange={setExportStatus}>
                    <SelectTrigger>
                      <SelectValue placeholder="All statuses" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="">All statuses</SelectItem>
                      <SelectItem value="published">Published</SelectItem>
                      <SelectItem value="draft">Draft</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <Button onClick={handleExport}>
                <Download className="h-4 w-4 mr-2" />
                Export Articles
              </Button>

              <p className="text-sm text-muted-foreground">
                Export articles in JSON format compatible with import. Use this to
                backup articles or transfer between environments.
              </p>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  )
}
