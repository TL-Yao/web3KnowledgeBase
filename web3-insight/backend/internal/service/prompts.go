package service

// Prompt templates for various LLM tasks

const PromptNewsSummary = `你是一个 Web3 新闻编辑。请将以下英文新闻翻译并总结为中文。

要求：
1. 保留原文的关键信息（数字、公司名、技术术语）
2. 专业术语使用「英文 (中文)」格式，如 "Layer 2 (二层网络)"
3. 总结长度：100-200 字
4. 返回 JSON 格式

原文标题：%s

原文内容：
%s

请返回以下 JSON 格式（不要包含 markdown 代码块标记）：
{
  "title": "中文标题",
  "summary": "中文摘要（100-200字）",
  "category": "tech/finance/product/company/regulation 中的一个",
  "tags": ["标签1", "标签2", "标签3"]
}`

const PromptClassification = `你是一个 Web3 内容分类专家。根据文章内容，推荐最合适的分类。

可用分类（按路径表示层级）：
%s

文章标题：%s
文章内容摘要：%s

请返回以下 JSON 格式（不要包含 markdown 代码块标记）：
{
  "primaryCategory": "分类路径，如 '基础技术/区块链原理/共识机制'",
  "secondaryCategories": ["可选的次要分类路径"],
  "suggestedTags": ["标签1", "标签2"],
  "confidence": 0.85,
  "reasoning": "分类理由简述"
}`

const PromptKnowledgeArticle = `你是一个 Web3 技术专家，正在为一位刚入职区块链公司的程序员撰写技术文档。

要求：
1. 使用中文撰写，保持专业性和准确性
2. 专业术语格式：英文术语 (中文翻译)，如 "Rollup (卷叠)"
3. 首次出现的缩写需要展开，如 "EVM (Ethereum Virtual Machine, 以太坊虚拟机)"
4. 内容结构：
   - ## 概述（一段话介绍）
   - ## 工作原理（详细解释机制）
   - ## 技术细节（深入技术层面）
   - ## 优势与局限（客观分析）
   - ## 实际应用（举例说明）
   - ## 相关技术（关联其他概念）
5. 文章长度：2000-4000 字
6. 如有代码示例，使用 markdown 代码块

主题：%s

参考资料：
%s

请直接输出 markdown 格式的文章内容。`

const PromptInstantResearch = `你是一个 Web3 技术研究助手。用户想了解一个技术概念，请提供全面且深入的解释。

要求：
1. 使用中文回答
2. 专业术语格式：英文术语 (中文翻译)
3. 结构清晰，使用 markdown 标题和列表
4. 如果是比较新的概念，说明其背景和发展现状
5. 提供实用的理解角度

用户问题：%s

%s

请提供详细的解释。`
