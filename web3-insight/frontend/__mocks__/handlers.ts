// __mocks__/handlers.ts
import { http, HttpResponse } from 'msw'
import { mockArticles, mockCategories, mockDataSources } from './data'

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

export const handlers = [
  // Categories - return empty arrays by default
  http.get(`${API_BASE}/api/categories`, () => {
    return HttpResponse.json([])
  }),

  http.get(`${API_BASE}/api/categories/tree`, () => {
    return HttpResponse.json([])
  }),

  // Articles - return empty results
  http.get(`${API_BASE}/api/articles`, ({ request }) => {
    const url = new URL(request.url)
    const page = parseInt(url.searchParams.get('page') || '1')
    const pageSize = parseInt(url.searchParams.get('page_size') || '10')

    return HttpResponse.json({
      articles: [],
      total: 0,
      page,
      pageSize,
    })
  }),

  http.get(`${API_BASE}/api/articles/:slug`, () => {
    return new HttpResponse(null, { status: 404 })
  }),

  // Data Sources - return empty array
  http.get(`${API_BASE}/api/sources`, () => {
    return HttpResponse.json([])
  }),

  http.post(`${API_BASE}/api/sources/:id/sync`, () => {
    return HttpResponse.json({ message: 'Sync started', itemsFound: 0, itemsNew: 0 })
  }),

  // Search - return empty results
  http.get(`${API_BASE}/api/search`, () => {
    return HttpResponse.json({
      articles: [],
      total: 0,
    })
  }),
]
