// frontend/config/features.ts

/**
 * 功能开关配置
 * 用于控制未实现或正在开发中的功能
 *
 * true = 功能已实现，正常显示
 * false = 功能未实现，隐藏或禁用相关 UI
 */
export const featureFlags = {
  // ===== 核心功能 =====
  /** 知识库文章列表和详情 */
  knowledgeBase: true,
  /** 文章 CRUD 操作 */
  articleCRUD: true,
  /** 分类管理 */
  categoryManagement: true,
  /** 全局搜索 */
  globalSearch: true,
  /** 文章详情页 AI 聊天 */
  articleChat: true,

  // ===== 管理功能 =====
  /** 系统状态监控（后端 /health 端点） */
  systemStatus: true,
  /** 配置管理 API */
  configAPI: true,
  /** 任务队列监控 */
  taskMonitor: true,

  // ===== 未完全实现功能 =====
  /** 即时研究功能（后端 API 返回占位符） */
  instantResearch: false,
  /** 数据源管理（无后端 API） */
  dataSourceManagement: false,
  /** RSS 同步功能 */
  rssSync: false,
  /** 网页爬虫功能 */
  webCrawler: false,
  /** 文章重新生成（需要 Asynq 集成） */
  articleRegenerate: false,
  /** 模型路由配置 UI */
  modelRoutingConfig: false,
  /** 提示词模板管理 */
  promptTemplates: false,
  /** 爬虫配置 UI */
  crawlerConfig: false,

  // ===== 模型相关 =====
  /** Ollama 本地模型 */
  ollamaModels: true,
  /** Claude API */
  claudeAPI: true,
  /** OpenAI API（未配置） */
  openAIAPI: false,
} as const

export type FeatureFlag = keyof typeof featureFlags

/**
 * 检查功能是否启用
 */
export function isFeatureEnabled(feature: FeatureFlag): boolean {
  return featureFlags[feature]
}

/**
 * 开发模式下显示所有功能（用于开发调试）
 */
export const DEV_SHOW_ALL_FEATURES = false
