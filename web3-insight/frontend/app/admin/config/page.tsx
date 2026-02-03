import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ModelConfig } from '@/components/admin/model-config'
import { SourceConfig } from '@/components/admin/source-config'
import { isFeatureEnabled } from '@/config/features'
import { Badge } from '@/components/ui/badge'
import { DisabledFeature } from '@/components/ui/disabled-feature'

export default function ConfigPage() {
  return (
    <div className="p-6">
      <h1 className="text-2xl font-semibold mb-6">配置管理</h1>

      <Tabs defaultValue="models">
        <TabsList>
          <TabsTrigger value="models">模型设置</TabsTrigger>
          <TabsTrigger value="sources">数据源</TabsTrigger>
          <TabsTrigger value="crawler" disabled={!isFeatureEnabled('crawlerConfig')}>
            爬虫
            {!isFeatureEnabled('crawlerConfig') && (
              <Badge variant="outline" className="ml-2 text-xs">开发中</Badge>
            )}
          </TabsTrigger>
          <TabsTrigger value="prompts" disabled={!isFeatureEnabled('promptTemplates')}>
            Prompt 模板
            {!isFeatureEnabled('promptTemplates') && (
              <Badge variant="outline" className="ml-2 text-xs">开发中</Badge>
            )}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="models" className="mt-6">
          <ModelConfig />
        </TabsContent>

        <TabsContent value="sources" className="mt-6">
          <SourceConfig />
        </TabsContent>

        <TabsContent value="crawler" className="mt-6">
          <DisabledFeature
            featureName="爬虫配置"
            description="功能正在开发中"
            variant="card"
          />
        </TabsContent>

        <TabsContent value="prompts" className="mt-6">
          <DisabledFeature
            featureName="Prompt 模板配置"
            description="功能正在开发中"
            variant="card"
          />
        </TabsContent>
      </Tabs>
    </div>
  )
}
