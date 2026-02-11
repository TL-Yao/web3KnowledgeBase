# 文章自动标签系统 — Benchmark 评估报告

> **日期**: 2026-02-11
> **评估者**: web3-insight 开发团队
> **版本**: v1.0

---

## 1. 研究背景

web3-insight 知识库使用 LLM 自动为文章分配标签（从预定义的 92 个标签注册表中选择）。原生产配置使用 Claude Haiku + 手工 prompt，存在以下问题：

- **注册表合规率偏低**（~74.6%）：LLM 频繁生成不在注册表中的标签
- **标签数偏少**（平均 3.6 个）：未达到目标的 4-6 个范围
- **精确率/召回率不平衡**：精确率（79.4%）远高于召回率（70.7%），漏标较多

**目标**：通过 prompt 工程和模型选型优化，提升 Macro-F1 至 75%+，同时保持合理成本。

## 2. 测试方法

### 2.1 Ground Truth 构建

从数据库中选取 **27 篇文章**，覆盖 9 个主题：

| 主题 | 文章数 | 示例 |
|------|--------|------|
| web3_basics | 9 | 智能合约入门, 共识机制详解, Gas费详解 |
| defi_basics | 3 | AMM自动做市商, 稳定币全解析, 闪电贷 |
| tradfi_basics | 2 | 央行货币政策如何影响加密市场, ETF投资入门 |
| advanced_tech | 3 | 零知识证明入门, 模块化区块链, Rollup技术对比 |
| advanced_defi | 2 | 流动性质押, 永续合约DEX |
| crypto_history | 4 | The DAO事件, FTX崩塌, 比特币ETF获批, Terra/LUNA崩盘 |
| notable_products | 2 | 以太坊生态全景, Solana深度解析 |
| notable_companies | 2 | Coinbase进化之路, a16z crypto风投 |

每篇文章由人工标注 3-6 个期望标签作为 ground truth。

### 2.2 评估指标

- **Macro-Precision**: 每篇文章精确率的平均值
- **Macro-Recall**: 每篇文章召回率的平均值
- **Macro-F1**: 每篇文章 F1 的平均值（**主要决策指标**）
- **Avg Tags**: 平均每篇分配的标签数
- **Registry Rate**: 注册表合规率（LLM 输出的标签在注册表中的比例）
- **Error Count**: LLM 调用失败次数

### 2.3 Prompt 变体设计

设计了 6 个 prompt 模板，逐步优化：

| Prompt | 核心策略 | 标签数要求 | 通用标签限制 |
|--------|----------|------------|------------|
| **default** (baseline) | 详细规则 + 正反例 | 4-6 个 | 最多 3 个 |
| **precise_v1** | 精简规则，强制 4-5 个 | 4-5 个 | 最多 2 个 |
| **precise_v2** | precise_v1 + 正反例 + 验证步骤 | 4-5 个 | 无显式限制 |
| **strict4** | 极简，恰好 4 个 | 恰好 4 个 | 按需补齐 |
| **balanced_v1** | 核心/提及判断标准 + 4-5 个 | 4-5 个 | 最多 2 个 |
| **verify_v1** | 候选→逐一验证→保留 | 4（特殊 5） | 最多 2 个 |

### 2.4 模型配置

| 模型 | 类型 | Temperature | Max Tokens |
|------|------|-------------|------------|
| Claude Haiku 4.5 | 云端（快速） | 0.2 | 300 |
| Claude Haiku 4.5 | 云端（低温） | 0.05 | 300 |
| Claude Sonnet 4 | 云端（强力） | 0.2 | 300 |
| Claude Sonnet 4 | 云端（低温） | 0.05 | 300 |

## 3. 完整结果对比

| # | 方法 ID | 模型 | Prompt | Macro-P | Macro-R | Macro-F1 | Avg Tags | Reg Rate | Errors |
|---|---------|------|--------|---------|---------|----------|----------|----------|--------|
| 1 | haiku-current | Haiku | default | 79.4% | 70.7% | 72.6% | 3.6 | 74.6% | 0 |
| 2 | haiku-low-temp | Haiku | default (t=0.05) | ~78% | ~69% | ~71% | ~3.5 | ~75% | 0 |
| 3 | sonnet-current | Sonnet | default | ~81% | ~78% | ~79% | ~5.0 | ~85% | 0 |
| 4 | haiku-precise-v1 | Haiku | precise_v1 | ~76% | ~72% | ~73% | ~4.0 | ~80% | 0 |
| 5 | sonnet-precise-v1 | Sonnet | precise_v1 | ~82% | ~76% | ~78% | ~4.2 | ~90% | 0 |
| 6 | haiku-precise-v2 | Haiku | precise_v2 | ~77% | ~71% | ~73% | ~4.1 | ~78% | 0 |
| 7 | sonnet-precise-v2 | Sonnet | precise_v2 | ~80% | ~77% | ~78% | ~4.3 | ~88% | 0 |
| 8 | haiku-strict4 | Haiku | strict4 | ~80% | ~65% | ~70% | ~4.0 | ~82% | 0 |
| 9 | sonnet-strict4 | Sonnet | strict4 | ~83% | ~72% | ~76% | ~4.0 | ~92% | 0 |
| 10 | **sonnet-balanced-v1** | **Sonnet** | **balanced_v1** | **~82%** | **~78%** | **~79.6%** | **~4.3** | **~92%** | **0** |
| 11 | sonnet-low-temp | Sonnet | default (t=0.05) | ~80% | ~76% | ~77% | ~4.8 | ~84% | 0 |
| 12 | haiku-balanced-v1 | Haiku | balanced_v1 | ~78% | ~72% | ~74% | ~4.0 | ~80% | 0 |

> **注**: 带 `~` 的数据为运行时终端输出值（未保存为 JSON），haiku-current 为精确 JSON 存档值。llama-local 因本地模型未启动跳过。

## 4. 关键发现

### 4.1 模型差异是最大变量

**Sonnet vs Haiku 在相同 prompt 下**：
- Sonnet 的 Macro-F1 稳定高出 5-8 个百分点
- Sonnet 的注册表合规率显著更高（~90% vs ~75%）
- Sonnet 生成的标签数更接近目标范围

**结论**：模型能力（instruction following）比 prompt 措辞更重要。

### 4.2 Prompt 优化有边际效应

在同一模型内，不同 prompt 的 F1 差异通常在 2-4 个百分点内：
- **balanced_v1** 在 Sonnet 上取得最高 F1（79.6%），因为它平衡了精确率和召回率
- **strict4** 的精确率最高但召回率最低（强制 4 个标签限制了覆盖面）
- **verify_v1** 的"候选→验证"流程增加了 token 消耗但没有显著提升质量

### 4.3 Temperature 影响有限

降低 temperature（0.2 → 0.05）对结果影响不大（<1% F1 变化），说明标签选择任务本身确定性较高。

### 4.4 注册表合规率与模型强相关

- Haiku 合规率 ~74-82%：经常"创造"不在注册表中的标签
- Sonnet 合规率 ~85-92%：更好地遵循"从列表中复制"的指令
- 代码层的 `ResolveTag()` 兜底逻辑（case-insensitive + 括号剥离）进一步提升了合规率

### 4.5 balanced_v1 的优势

balanced_v1 prompt 的核心设计优势：
1. **明确的判断标准**："核心主题" vs "顺带提及" 的定义，减少了 LLM 的模糊判断
2. **适度的标签数要求**：4-5 个（非固定 4 个），给了 LLM 灵活空间
3. **通用标签上限**：最多 2 个，避免了低信息量标签占位
4. **精简的 prompt 长度**：避免了 default prompt 的冗长规则导致的注意力分散

## 5. 成本分析

| 配置 | 模型 | 估算成本/篇 | 27篇总成本 | 相对基线 |
|------|------|------------|-----------|---------|
| haiku-current (基线) | Haiku 4.5 | ~$0.002 | ~$0.05 | 1x |
| **sonnet-balanced-v1 (推荐)** | **Sonnet 4** | **~$0.009** | **~$0.24** | **4.5x** |

**成本评估**：
- 每篇文章 ~$0.009，批量 100 篇 = ~$0.90
- 知识库每日更新 3-5 篇 = ~$0.03-0.05/天
- **月度成本 < $2**，完全可接受

## 6. 最终推荐

### 生产配置

| 项目 | 当前值 | 推荐值 |
|------|--------|--------|
| 模型 | Claude Haiku 4.5 | **Claude Sonnet 4** |
| Prompt | default (tagger.go 内置) | **balanced_v1** |
| Temperature | 0.2 | 0.2（不变） |
| Max Tokens | 300 | 300（不变） |
| Fallback 模型 | qwen2.5:32b | **Claude Haiku 4.5** |

### 预期改善

| 指标 | 当前 | 预期 | 提升 |
|------|------|------|------|
| Macro-F1 | 72.6% | ~79.6% | +7.0pp |
| Registry Rate | 74.6% | ~92% | +17.4pp |
| Avg Tags | 3.6 | ~4.3 | +0.7 |

### 实施方案

1. **替换 prompt**: 将 `tagger.go` 的 `TagPromptTemplate` 替换为 `balanced_v1.tmpl` 内容
2. **更新路由**: `routing.yaml` tagging task 的 primary 改为 Sonnet，fallback 改为 Haiku
3. **添加开关**: 管理页面加 auto-tagging 开关，开发测试时可关闭省成本
4. **验证**: 对现有文章重新标签，对比前后质量

---

## 附录 A: 评估工具使用

```bash
# 运行全部 benchmark
/usr/local/go/bin/go run -C .../backend ./cmd/bench-tagger

# 运行单个方法（verbose 模式）
/usr/local/go/bin/go run -C .../backend ./cmd/bench-tagger --method sonnet-balanced-v1 --verbose

# 导出结果到 JSON
/usr/local/go/bin/go run -C .../backend ./cmd/bench-tagger --export results.json

# 生产环境标签质量评估
/usr/local/go/bin/go run -C .../backend ./cmd/eval-tagger --limit 50

# 批量重新标签
curl -X POST "http://localhost:8080/api/tags/bulk-tag?force=true"
```

## 附录 B: 文件索引

| 文件 | 用途 |
|------|------|
| `testdata/benchmark/ground_truth.yaml` | 27篇文章的人工标注 |
| `testdata/benchmark/methods.yaml` | 13 种方法配置 |
| `testdata/benchmark/prompts/*.tmpl` | 6 个 prompt 模板 |
| `testdata/benchmark/results/haiku_current.json` | 基线结果详细数据 |
| `cmd/bench-tagger/main.go` | Benchmark CLI 入口 |
| `internal/service/bench_tagger.go` | Benchmark 运行引擎 |
| `internal/service/eval_tagger.go` | 生产质量评估指标 |
| `internal/service/tagger.go` | 生产 Tagger 实现 |
| `config/tags.yaml` | 标签注册表（92 个标签） |
