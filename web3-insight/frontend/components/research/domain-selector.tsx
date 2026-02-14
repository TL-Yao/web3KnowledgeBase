'use client'

import {
  Sparkles,
  Cpu,
  TrendingUp,
  Building2,
  Newspaper,
  Package,
  FlaskConical,
  BookOpen,
} from 'lucide-react'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { LucideIcon } from 'lucide-react'

export interface ResearchDomain {
  id: string
  name: string
  icon: LucideIcon
}

const DOMAINS: ResearchDomain[] = [
  { id: 'auto', name: '自动识别', icon: Sparkles },
  { id: 'tech-engineering', name: '科技/工程', icon: Cpu },
  { id: 'finance-markets', name: '金融/市场', icon: TrendingUp },
  { id: 'industry-business', name: '产业/商业', icon: Building2 },
  { id: 'news-current', name: '新闻/时事', icon: Newspaper },
  { id: 'product-analysis', name: '产品分析', icon: Package },
  { id: 'science-academic', name: '科学/学术', icon: FlaskConical },
  { id: 'general', name: '通用知识', icon: BookOpen },
]

interface DomainSelectorProps {
  value: string
  onChange: (value: string) => void
  className?: string
}

export function DomainSelector({ value, onChange, className }: DomainSelectorProps) {
  const selected = DOMAINS.find((d) => d.id === value) ?? DOMAINS[0]

  return (
    <Select value={value} onValueChange={onChange}>
      <SelectTrigger size="sm" className={className}>
        <SelectValue>
          <selected.icon className="size-4" />
          {selected.name}
        </SelectValue>
      </SelectTrigger>
      <SelectContent>
        {DOMAINS.map((domain) => (
          <SelectItem key={domain.id} value={domain.id}>
            <domain.icon className="size-4" />
            {domain.name}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}

export { DOMAINS }
