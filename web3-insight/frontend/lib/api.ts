// API client for backend communication

const API_BASE = process.env.NEXT_PUBLIC_API_URL || ''

export class APIError extends Error {
  constructor(public status: number, message: string) {
    super(message)
    this.name = 'APIError'
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
    throw new APIError(res.status, errorText)
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
  modelUsed?: string
  viewCount?: number
  status?: string
  createdAt: string
  updatedAt: string
}

export interface ArticleListResponse {
  data: Article[]
  total: number
  page: number
  limit: number
}

export interface ArticleListParams {
  category?: string
  page?: number
  limit?: number
  q?: string
}

// Article API
export const articleAPI = {
  list: (params?: ArticleListParams) => {
    const searchParams = new URLSearchParams()
    if (params?.category) searchParams.set('category', params.category)
    if (params?.page) searchParams.set('page', String(params.page))
    if (params?.limit) searchParams.set('limit', String(params.limit))
    if (params?.q) searchParams.set('q', params.q)

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
}

// Categories API
export interface Category {
  id: string
  name: string
  slug: string
  count: number
}

export const categoryAPI = {
  list: () => fetchAPI<Category[]>('/api/categories'),
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
