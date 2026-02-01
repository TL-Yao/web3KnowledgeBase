import { MainLayout } from '@/components/layout/main-layout'
import { ArticleList } from '@/components/knowledge/article-list'

export default function KnowledgePage() {
  return (
    <MainLayout>
      <div className="p-6">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-2xl font-semibold">知识库</h1>
        </div>
        <ArticleList />
      </div>
    </MainLayout>
  )
}
