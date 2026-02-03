// __mocks__/handlers.ts
import { http, HttpResponse } from 'msw'
import { mockArticles, mockCategories, mockDataSources } from './data'

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

export const handlers = [
  // Categories
  http.get(`${API_BASE}/api/categories`, () => {
    return HttpResponse.json(mockCategories)
  }),

  http.get(`${API_BASE}/api/categories/tree`, () => {
    return HttpResponse.json(mockCategories)
  }),

  // Articles
  http.get(`${API_BASE}/api/articles`, ({ request }) => {
    const url = new URL(request.url)
    const page = parseInt(url.searchParams.get('page') || '1')
    const pageSize = parseInt(url.searchParams.get('page_size') || '10')
    const search = url.searchParams.get('search') || ''

    let filtered = mockArticles
    if (search) {
      filtered = mockArticles.filter(a =>
        a.title.includes(search) || a.summary?.includes(search)
      )
    }

    return HttpResponse.json({
      articles: filtered.slice((page - 1) * pageSize, page * pageSize),
      total: filtered.length,
      page,
      pageSize,
    })
  }),

  http.get(`${API_BASE}/api/articles/:slug`, ({ params }) => {
    const article = mockArticles.find(a => a.slug === params.slug)
    if (!article) {
      return new HttpResponse(null, { status: 404 })
    }
    return HttpResponse.json(article)
  }),

  // Data Sources (endpoint is /api/sources per the actual API)
  http.get(`${API_BASE}/api/sources`, () => {
    return HttpResponse.json(mockDataSources)
  }),

  http.post(`${API_BASE}/api/sources/:id/sync`, () => {
    return HttpResponse.json({ message: 'Sync started', itemsFound: 0, itemsNew: 0 })
  }),

  // Search
  http.get(`${API_BASE}/api/search`, ({ request }) => {
    const url = new URL(request.url)
    const q = url.searchParams.get('q') || ''
    const filtered = mockArticles.filter(a =>
      a.title.includes(q) || a.content?.includes(q)
    )
    return HttpResponse.json({
      articles: filtered,
      total: filtered.length,
    })
  }),
]
