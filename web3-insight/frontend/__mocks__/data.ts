// __mocks__/data.ts
import type { Article, Category, DataSource } from '@/lib/api'

export const mockCategories: Category[] = [
  {
    id: 'cat-1',
    name: 'Layer 1',
    nameEn: 'Layer 1',
    slug: 'layer-1',
    icon: 'layers',
    articleCount: 5,
    sortOrder: 1,
    children: [
      {
        id: 'cat-1-1',
        name: 'Ethereum',
        nameEn: 'Ethereum',
        slug: 'ethereum',
        icon: 'coins',
        parentId: 'cat-1',
        articleCount: 3,
        sortOrder: 1,
        children: [],
      },
    ],
  },
  {
    id: 'cat-2',
    name: 'DeFi',
    nameEn: 'DeFi',
    slug: 'defi',
    icon: 'coins',
    articleCount: 8,
    sortOrder: 2,
    children: [],
  },
]

export const mockArticles: Article[] = [
  {
    id: 'art-1',
    title: '以太坊 2.0 升级详解',
    slug: 'ethereum-2-upgrade',
    content: '# 以太坊 2.0\n\n以太坊 2.0 是以太坊网络的重大升级...',
    summary: '本文详细介绍以太坊 2.0 的技术架构和升级路线图',
    status: 'published',
    categoryId: 'cat-1-1',
    category: {
      id: 'cat-1-1',
      name: 'Ethereum',
      slug: 'ethereum',
    },
    tags: ['Ethereum', 'PoS', 'Sharding'],
    viewCount: 150,
    createdAt: '2024-01-15T10:00:00Z',
    updatedAt: '2024-01-15T10:00:00Z',
  },
  {
    id: 'art-2',
    title: 'DeFi 借贷协议对比',
    slug: 'defi-lending-comparison',
    content: '# DeFi 借贷\n\n本文对比主流 DeFi 借贷协议...',
    summary: '深入对比 Aave、Compound 和 MakerDAO 等借贷协议',
    status: 'published',
    categoryId: 'cat-2',
    category: {
      id: 'cat-2',
      name: 'DeFi',
      slug: 'defi',
    },
    tags: ['DeFi', 'Lending', 'Aave'],
    viewCount: 230,
    createdAt: '2024-01-10T08:00:00Z',
    updatedAt: '2024-01-12T14:00:00Z',
  },
]

export const mockDataSources: DataSource[] = [
  {
    id: 'ds-1',
    name: 'CoinDesk RSS',
    type: 'rss',
    url: 'https://www.coindesk.com/arc/outboundfeeds/rss/',
    enabled: true,
    fetchInterval: 3600,
    lastFetchedAt: '2024-01-15T12:00:00Z',
    config: {},
    createdAt: '2024-01-01T00:00:00Z',
  },
]
