import { MainLayout } from '@/components/layout/main-layout'
import { ResearchPanel } from '@/components/research/research-panel'

export default function ResearchPage() {
  return (
    <MainLayout>
      <div className="max-w-4xl mx-auto p-6">
        <h1 className="text-2xl font-semibold mb-6">即时研究</h1>
        <p className="text-muted-foreground mb-8">
          输入任何 Web3 技术名词或问题，AI 将为你搜索、分析并生成详细解释。
        </p>
        <ResearchPanel />
      </div>
    </MainLayout>
  )
}
