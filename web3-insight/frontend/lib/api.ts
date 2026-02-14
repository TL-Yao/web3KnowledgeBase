// API client for backend communication

const API_BASE = process.env.NEXT_PUBLIC_API_URL || ''

export class APIError extends Error {
  public type?: string

  constructor(public status: number, message: string, type?: string) {
    super(message)
    this.name = 'APIError'
    this.type = type
  }
}

export async function fetchAPI<T>(
  endpoint: string,
  options?: RequestInit
): Promise<T> {
  const res = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  })

  if (!res.ok) {
    const errorText = await res.text().catch(() => 'Unknown error')
    let errorType: string | undefined
    try {
      const parsed = JSON.parse(errorText)
      errorType = parsed.type
    } catch {}
    throw new APIError(res.status, errorText, errorType)
  }

  if (res.status === 204 || res.headers.get('content-length') === '0') {
    return undefined as T
  }

  return res.json()
}

// Types
export interface Article {
  id: string
  title: string
  slug: string
  content: string
  contentHtml?: string
  summary: string
  categoryId?: string
  category?: { id: string; name: string; slug: string }
  tags: string[]
  sourceUrls?: string[]
  sourceLanguage?: string
  sourceKeyword?: string
  modelUsed?: string
  viewCount?: number
  status?: string
  archived?: boolean
  createdAt: string
  updatedAt: string
}

export interface ArticleListResponse {
  articles: Article[]
  total: number
  page: number
  pageSize: number
}

export interface ArticleListParams {
  category?: string
  tag?: string
  page?: number
  limit?: number
  q?: string
  archived?: 'true' | 'false' | 'all'
}

// Article Update Types
export interface GenerateUpdateRequest {
  conversationHistory: Array<{ role: 'user' | 'assistant'; content: string }>
}

export interface GenerateUpdateResponse {
  updatedContent: string
  changeSummary: string
  model: string
  noChange?: boolean
  noChangeReason?: string
}

export interface GenerateUpdateAsyncResponse {
  jobId: string
}

export interface UpdateJobStatus {
  status: 'pending' | 'running' | 'completed' | 'failed'
  result?: GenerateUpdateResponse
  error?: string
  errorType?: string
}

export interface ApplyUpdateRequest {
  updatedContent: string
  changeSummary: string
}

export interface ArticleVersion {
  id: string
  articleId: string
  editedBy: string
  changeSummary: string
  createdAt: string
}

// Article API
export const articleAPI = {
  list: (params?: ArticleListParams) => {
    const searchParams = new URLSearchParams()
    if (params?.category) searchParams.set('category', params.category)
    if (params?.tag) searchParams.set('tag', params.tag)
    if (params?.page) searchParams.set('page', String(params.page))
    if (params?.limit) searchParams.set('page_size', String(params.limit))
    if (params?.q) searchParams.set('q', params.q)
    if (params?.archived) searchParams.set('archived', params.archived)

    const query = searchParams.toString()
    return fetchAPI<ArticleListResponse>(`/api/articles${query ? `?${query}` : ''}`)
  },

  get: (slug: string) => fetchAPI<Article>(`/api/articles/${slug}`),

  create: (data: Partial<Article>) =>
    fetchAPI<Article>('/api/articles', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (id: string, data: Partial<Article>) =>
    fetchAPI<Article>(`/api/articles/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: (id: string) =>
    fetchAPI<void>(`/api/articles/${id}`, {
      method: 'DELETE',
    }),

  toggleArchive: (id: string) =>
    fetchAPI<Article>(`/api/articles/${id}/archive`, {
      method: 'PATCH',
    }),

  updateTags: (id: string, tags: string[]) =>
    fetchAPI<Article>(`/api/articles/${id}/tags`, {
      method: 'PUT',
      body: JSON.stringify({ tags }),
    }),

  generateUpdate: (id: string, data: GenerateUpdateRequest, options?: { method?: 'cli' | 'api'; model?: string }) => {
    const params = new URLSearchParams()
    if (options?.method) params.set('method', options.method)
    if (options?.model) params.set('model', options.model)
    const query = params.toString()
    return fetchAPI<GenerateUpdateAsyncResponse>(`/api/articles/${id}/generate-update${query ? `?${query}` : ''}`, {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  getUpdateStatus: (id: string, jobId: string) =>
    fetchAPI<UpdateJobStatus>(`/api/articles/${id}/update-status?jobId=${jobId}`),

  cancelUpdate: (id: string, jobId: string) =>
    fetchAPI<void>(`/api/articles/${id}/cancel-update?jobId=${jobId}`, {
      method: 'POST',
    }),

  applyUpdate: (id: string, data: ApplyUpdateRequest) =>
    fetchAPI<{ article: Article; version: ArticleVersion }>(`/api/articles/${id}/apply-update`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  listVersions: (id: string) =>
    fetchAPI<{ versions: ArticleVersion[] }>(`/api/articles/${id}/versions`),

  rollback: (id: string, versionId: string) =>
    fetchAPI<{ article: Article; message: string }>(`/api/articles/${id}/versions/${versionId}/rollback`, {
      method: 'POST',
    }),
}

// Categories API
export interface Category {
  id: string
  name: string
  nameEn: string
  slug: string
  parentId?: string
  description?: string
  icon?: string
  sortOrder: number
  children?: Category[]
  articleCount: number
  createdAt: string
  updatedAt: string
}

export const categoryAPI = {
  list: () => fetchAPI<Category[]>('/api/categories'),
  getTree: () => fetchAPI<Category[]>('/api/categories/tree'),
}

// Data Sources API
export interface DataSource {
  id: string
  name: string
  type: 'rss' | 'api' | 'crawl'
  url: string
  config?: Record<string, unknown>
  enabled: boolean
  fetchInterval: number
  lastFetchedAt?: string
  lastError?: string
  createdAt: string
}

export interface CreateDataSourceRequest {
  name: string
  type: 'rss' | 'api' | 'crawl'
  url: string
  config?: Record<string, unknown>
  enabled?: boolean
  fetchInterval?: number
}

export interface ValidateURLResponse {
  valid: boolean
  error?: string
  title?: string
  description?: string
  itemCount?: number
}

export interface SyncResult {
  message: string
  itemsFound: number
  itemsNew: number
}

export const dataSourceAPI = {
  list: () => fetchAPI<DataSource[]>('/api/sources'),

  get: (id: string) => fetchAPI<DataSource>(`/api/sources/${id}`),

  create: (data: CreateDataSourceRequest) =>
    fetchAPI<DataSource>('/api/sources', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (id: string, data: CreateDataSourceRequest) =>
    fetchAPI<DataSource>(`/api/sources/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: (id: string) =>
    fetchAPI<void>(`/api/sources/${id}`, {
      method: 'DELETE',
    }),

  sync: (id: string) =>
    fetchAPI<SyncResult>(`/api/sources/${id}/sync`, {
      method: 'POST',
    }),

  validate: (url: string, type: string) =>
    fetchAPI<ValidateURLResponse>('/api/sources/validate', {
      method: 'POST',
      body: JSON.stringify({ url, type }),
    }),
}

// Import/Export API
export interface ImportArticle {
  title: string
  content: string
  contentHtml?: string
  summary?: string
  categoryPath?: string
  categoryId?: string
  tags?: string[]
  status?: 'draft' | 'published'
  sourceUrls?: string[]
  slug?: string
}

export interface ImportBatch {
  articles: ImportArticle[]
  options?: {
    skipDuplicates?: boolean
    updateExisting?: boolean
    generateSummary?: boolean
    defaultStatus?: string
  }
}

export interface ImportError {
  index: number
  title: string
  message: string
}

export interface ImportResult {
  totalCount: number
  importedCount: number
  skippedCount: number
  updatedCount: number
  errorCount: number
  errors?: ImportError[]
  importedIds?: string[]
}

export interface ValidationResult {
  valid: boolean
  errors: ImportError[]
  errorCount: number
  totalCount: number
}

export const importAPI = {
  import: (batch: ImportBatch) =>
    fetchAPI<ImportResult>('/api/import', {
      method: 'POST',
      body: JSON.stringify(batch),
    }),

  validate: (batch: ImportBatch) =>
    fetchAPI<ValidationResult>('/api/import/validate', {
      method: 'POST',
      body: JSON.stringify(batch),
    }),

  getTemplate: () => `/api/import/template`,

  export: (categoryId?: string, status?: string) => {
    const params = new URLSearchParams()
    if (categoryId) params.set('categoryId', categoryId)
    if (status) params.set('status', status)
    const query = params.toString()
    return `/api/import/export${query ? `?${query}` : ''}`
  },

  uploadFile: async (file: File, options?: { skipDuplicates?: boolean; updateExisting?: boolean }) => {
    const formData = new FormData()
    formData.append('file', file)
    if (options?.skipDuplicates) formData.append('skipDuplicates', 'true')
    if (options?.updateExisting) formData.append('updateExisting', 'true')

    const res = await fetch('/api/import/upload', {
      method: 'POST',
      body: formData,
    })

    if (!res.ok) {
      const errorText = await res.text().catch(() => 'Unknown error')
      throw new APIError(res.status, errorText)
    }

    return res.json() as Promise<ImportResult>
  },
}

// Explorer Research API
export interface ExplorerResearch {
  id: string
  chainName: string
  chainType?: string
  explorerName: string
  explorerUrl: string
  explorerType?: string
  features?: Record<string, unknown>
  uiFeatures?: Record<string, unknown>
  apiFeatures?: Record<string, unknown>
  screenshots?: string[]
  analysis?: string
  strengths?: string[]
  weaknesses?: string[]
  popularityScore?: number
  researchStatus: 'pending' | 'in_progress' | 'completed'
  researchNotes?: string
  lastUpdated: string
  createdAt: string
}

export interface ExplorerFeature {
  id: string
  category: string
  name: string
  description: string
  importance: 'high' | 'medium' | 'low'
  sortOrder: number
}

export interface ExplorerStats {
  total: number
  byStatus: Record<string, number>
  byChain: Array<{ chain: string; count: number }>
}

export interface CreateExplorerRequest {
  chainName: string
  chainType?: string
  explorerName: string
  explorerUrl: string
  explorerType?: string
  features?: Record<string, unknown>
  uiFeatures?: Record<string, unknown>
  apiFeatures?: Record<string, unknown>
  screenshots?: string[]
  analysis?: string
  strengths?: string[]
  weaknesses?: string[]
  popularityScore?: number
  researchStatus?: string
  researchNotes?: string
}

export const explorerAPI = {
  list: (chain?: string, status?: string) => {
    const params = new URLSearchParams()
    if (chain) params.set('chain', chain)
    if (status) params.set('status', status)
    const query = params.toString()
    return fetchAPI<{ data: ExplorerResearch[]; count: number }>(`/api/explorers${query ? `?${query}` : ''}`)
  },

  get: (id: string) => fetchAPI<ExplorerResearch>(`/api/explorers/${id}`),

  create: (data: CreateExplorerRequest) =>
    fetchAPI<ExplorerResearch>('/api/explorers', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (id: string, data: CreateExplorerRequest) =>
    fetchAPI<ExplorerResearch>(`/api/explorers/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: (id: string) =>
    fetchAPI<void>(`/api/explorers/${id}`, { method: 'DELETE' }),

  getChains: () => fetchAPI<{ chains: string[]; count: number }>('/api/explorers/chains'),

  getStats: () => fetchAPI<ExplorerStats>('/api/explorers/stats'),

  getFeatures: (category?: string) => {
    const query = category ? `?category=${category}` : ''
    return fetchAPI<{
      features: ExplorerFeature[]
      byCategory: Record<string, ExplorerFeature[]>
      categories: string[]
    }>(`/api/explorers/features${query}`)
  },

  compare: (ids: string[]) =>
    fetchAPI<{
      explorers: ExplorerResearch[]
      features: ExplorerFeature[]
      count: number
    }>(`/api/explorers/compare?ids=${ids.join(',')}`),

  updateStatus: (id: string, status: string) =>
    fetchAPI<{ message: string; status: string }>(`/api/explorers/${id}/status`, {
      method: 'POST',
      body: JSON.stringify({ status }),
    }),

  seedFeatures: () =>
    fetchAPI<{ message: string }>('/api/explorers/features/seed', { method: 'POST' }),
}

// Model Configuration Types
export interface Model {
  id: string
  name: string
  provider: string
  enabled: boolean
  capabilities: string[]
  contextWindow: number
  costPer1kTokens: number
}

export interface ModelsConfig {
  localModels: Model[]
  cloudModels: Model[]
}

export interface TaskType {
  id: string
  name: string
  description: string
  defaultPrimary: string
  defaultFallback: string
  requiredCapability: string
}

export interface RoutingConfig {
  taskTypes: TaskType[]
}

export interface TaskSelection {
  taskId: string
  primary: string
  fallback: string
}

// Model Configuration API
export const modelConfigAPI = {
  // Get available models from registry
  getModelsRegistry: () =>
    fetchAPI<ModelsConfig>('/api/models/registry'),

  // Get task types
  getTaskTypes: () =>
    fetchAPI<RoutingConfig>('/api/models/tasks'),

  // Get user's model selections
  getUserSelections: () =>
    fetchAPI<TaskSelection[]>('/api/models/selections'),

  // Update user's model selections
  updateUserSelections: (selections: TaskSelection[]) =>
    fetchAPI<TaskSelection[]>('/api/models/selections', {
      method: 'PUT',
      body: JSON.stringify(selections),
    }),
}

// Theme Types
export interface Theme {
  id: string
  name: string
  category: string
  description: string
  status: 'active' | 'paused'
  sortOrder: number
  keywordsPending?: number
  keywordsUsed?: number
  keywordsTotal?: number
}

export interface ThemeListResponse {
  themes: Theme[]
  activeThemeId: string | null
}

export interface KBConfig {
  batchSize: number
  maxBatchSize: number
}

// Knowledge Base Update Types
export interface KBUpdateJob {
  id: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  triggerType: 'manual' | 'scheduled'
  keywordsGenerated: number
  articlesGenerated: number
  articlesPublished: number
  startedAt?: string
  completedAt?: string
  errorMessage?: string
  createdAt: string
  updatedAt: string
}

export interface KBUpdateJobListResponse {
  total: number
  page: number
  pageSize: number
  items: KBUpdateJob[]
}

export interface KBKeywordStats {
  pending: number
  used: number
  total: number
}

export interface TriggerUpdateRequest {
  triggerType?: 'manual' | 'scheduled'
}

export interface TriggerUpdateResponse {
  status: string
  message: string
  jobId?: string
}

export interface InitKeywordsRequest {
  count?: number
}

export interface InitKeywordsResponse {
  status: string
  message: string
  count: number
}

// Knowledge Base Update API
export const kbUpdateAPI = {
  // Trigger update
  trigger: (data?: TriggerUpdateRequest) =>
    fetchAPI<TriggerUpdateResponse>('/api/kb/update/trigger', {
      method: 'POST',
      body: JSON.stringify(data || { triggerType: 'manual' }),
    }),

  // Get job status
  getJob: (jobId: string) =>
    fetchAPI<KBUpdateJob>(`/api/kb/update/jobs/${jobId}`),

  // Get job history
  getJobs: (page: number = 1, pageSize: number = 20) =>
    fetchAPI<KBUpdateJobListResponse>(`/api/kb/update/jobs?page=${page}&page_size=${pageSize}`),

  // Initialize keywords
  initKeywords: (data?: InitKeywordsRequest) =>
    fetchAPI<InitKeywordsResponse>('/api/kb/keywords/init', {
      method: 'POST',
      body: JSON.stringify(data || {}),
    }),

  // Get keyword stats
  getKeywordStats: () =>
    fetchAPI<KBKeywordStats>('/api/kb/keywords/stats'),

  // Theme management
  getThemes: () =>
    fetchAPI<ThemeListResponse>('/api/kb/themes'),

  getActiveTheme: () =>
    fetchAPI<Theme>('/api/kb/themes/active'),

  setActiveTheme: (themeId: string) =>
    fetchAPI<{ message: string }>(`/api/kb/themes/${themeId}/activate`, {
      method: 'PUT',
    }),

  // KB config
  getConfig: () =>
    fetchAPI<KBConfig>('/api/kb/config'),

  updateBatchSize: (size: number) =>
    fetchAPI<{ message: string }>('/api/kb/config/batch-size', {
      method: 'PUT',
      body: JSON.stringify({ batchSize: size }),
    }),
}

// Tag Types
export interface Tag {
  id: string
  name: string
  nameEn: string
  themeId: string | null
  status: 'active' | 'pending' | 'deprecated'
  suggestCount: number
  createdAt: string
  updatedAt: string
}

export interface TagListResponse {
  tags: Tag[]
  total: number
}

export interface TagStats {
  totalTags: number
  activeTags: number
  pendingTags: number
  universalTags: number
}

export interface TagSearchResult {
  name: string
  nameEn: string
  themeId: string | null
}

export interface TagSearchResponse {
  tags: TagSearchResult[]
}

export interface TagInUse {
  name: string
  articleCount: number
}

export interface TagInUseResponse {
  tags: TagInUse[]
}

export interface CreateTagRequest {
  name: string
  nameEn?: string
  themeId?: string | null
}

export interface UpdateTagRequest {
  name?: string
  nameEn?: string
  themeId?: string | null
}

export interface DeleteTagResponse {
  message: string
  name: string
  articlesAffected: number
}

// Tag API
export const tagAPI = {
  getInUse: () => fetchAPI<TagInUseResponse>('/api/tags/in-use'),

  search: (q: string, limit?: number) =>
    fetchAPI<TagSearchResponse>(`/api/tags/search?q=${encodeURIComponent(q)}&limit=${limit || 10}`),

  list: (params?: { status?: string; page?: number; limit?: number; q?: string }) => {
    const searchParams = new URLSearchParams()
    if (params?.status) searchParams.set('status', params.status)
    if (params?.page) searchParams.set('page', String(params.page))
    if (params?.limit) searchParams.set('limit', String(params.limit))
    if (params?.q) searchParams.set('q', params.q)
    const query = searchParams.toString()
    return fetchAPI<TagListResponse>(`/api/tags${query ? `?${query}` : ''}`)
  },

  getStats: () => fetchAPI<TagStats>('/api/tags/stats'),

  updateStatus: (id: string, status: Tag['status']) =>
    fetchAPI<Tag>(`/api/tags/${id}/status`, {
      method: 'PUT',
      body: JSON.stringify({ status }),
    }),

  create: (data: CreateTagRequest) =>
    fetchAPI<Tag>('/api/tags', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (id: string, data: UpdateTagRequest) =>
    fetchAPI<Tag>(`/api/tags/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: (id: string) =>
    fetchAPI<DeleteTagResponse>(`/api/tags/${id}`, {
      method: 'DELETE',
    }),
}

// Config API
export interface ConfigValue {
  key: string
  value: string
  description?: string
}

export const configAPI = {
  get: (key: string) =>
    fetchAPI<ConfigValue>(`/api/config/${key}`),

  set: (key: string, value: string, description?: string) =>
    fetchAPI<ConfigValue>(`/api/config/${key}`, {
      method: 'PUT',
      body: JSON.stringify({ value, description }),
    }),
}

// API Key Types
export interface ApiKeyProvider {
  id: string
  name: string
  configured: boolean
  masked: string
}

export interface ApiKeyListResponse {
  providers: ApiKeyProvider[]
}

export interface ApiKeyTestResponse {
  provider: string
  success: boolean
  message: string
}

// API Key API
export const apiKeyAPI = {
  list: () =>
    fetchAPI<ApiKeyListResponse>('/api/models/keys'),

  save: (keys: Record<string, string>) =>
    fetchAPI<ApiKeyListResponse>('/api/models/keys', {
      method: 'PUT',
      body: JSON.stringify({ keys }),
    }),

  test: (provider: string, key?: string) =>
    fetchAPI<ApiKeyTestResponse>('/api/models/keys/test', {
      method: 'POST',
      body: JSON.stringify({ provider, ...(key ? { key } : {}) }),
    }),
}
