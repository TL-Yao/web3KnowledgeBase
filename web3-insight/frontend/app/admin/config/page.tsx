import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ModelConfig } from '@/components/admin/model-config'
import { SourceConfig } from '@/components/admin/source-config'

export default function ConfigPage() {
  return (
    <div className="p-6">
      <h1 className="text-2xl font-semibold mb-6">配置管理</h1>

      <Tabs defaultValue="models">
        <TabsList>
          <TabsTrigger value="models">模型设置</TabsTrigger>
          <TabsTrigger value="sources">数据源</TabsTrigger>
          <TabsTrigger value="crawler">爬虫</TabsTrigger>
          <TabsTrigger value="prompts">Prompt 模板</TabsTrigger>
        </TabsList>

        <TabsContent value="models" className="mt-6">
          <ModelConfig />
        </TabsContent>

        <TabsContent value="sources" className="mt-6">
          <SourceConfig />
        </TabsContent>

        <TabsContent value="crawler" className="mt-6">
          <div className="text-muted-foreground">爬虫配置 - 开发中</div>
        </TabsContent>

        <TabsContent value="prompts" className="mt-6">
          <div className="text-muted-foreground">Prompt 模板配置 - 开发中</div>
        </TabsContent>
      </Tabs>
    </div>
  )
}
