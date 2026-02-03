import { ArticleImport } from '@/components/admin/article-import'

export default function ImportPage() {
  return (
    <div className="p-6 space-y-6">
      <h1 className="text-2xl font-semibold">文章导入/导出</h1>
      <ArticleImport />
    </div>
  )
}
