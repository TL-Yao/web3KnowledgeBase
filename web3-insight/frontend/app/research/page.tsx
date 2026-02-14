import { MainLayout } from '@/components/layout/main-layout'
import { DisabledFeature } from '@/components/ui/disabled-feature'

export default function ResearchPage() {
  return (
    <MainLayout>
      <div className="max-w-6xl mx-auto p-6">
        <h1 className="text-2xl font-semibold mb-2">研究中心</h1>
        <p className="text-muted-foreground mb-6">
          探索 Web3 技术和区块链浏览器
        </p>

        <DisabledFeature
          featureName="研究中心"
          description="功能正在重构中"
          variant="card"
        />
      </div>
    </MainLayout>
  )
}
