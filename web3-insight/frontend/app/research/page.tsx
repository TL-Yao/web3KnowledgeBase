'use client'

import { useState } from 'react'
import { MainLayout } from '@/components/layout/main-layout'
import { ResearchPanel } from '@/components/research/research-panel'
import { ExplorerResearchPanel } from '@/components/research/explorer-research'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Search, Globe2 } from 'lucide-react'

export default function ResearchPage() {
  const [activeTab, setActiveTab] = useState('instant')

  return (
    <MainLayout>
      <div className="max-w-6xl mx-auto p-6">
        <h1 className="text-2xl font-semibold mb-2">研究中心</h1>
        <p className="text-muted-foreground mb-6">
          探索 Web3 技术和区块链浏览器
        </p>

        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="mb-6">
            <TabsTrigger value="instant" className="flex items-center gap-2">
              <Search className="h-4 w-4" />
              即时研究
            </TabsTrigger>
            <TabsTrigger value="explorers" className="flex items-center gap-2">
              <Globe2 className="h-4 w-4" />
              浏览器调研
            </TabsTrigger>
          </TabsList>

          <TabsContent value="instant">
            <div className="max-w-4xl">
              <p className="text-muted-foreground mb-8">
                输入任何 Web3 技术名词或问题，AI 将为你搜索、分析并生成详细解释。
              </p>
              <ResearchPanel />
            </div>
          </TabsContent>

          <TabsContent value="explorers">
            <ExplorerResearchPanel />
          </TabsContent>
        </Tabs>
      </div>
    </MainLayout>
  )
}
