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
