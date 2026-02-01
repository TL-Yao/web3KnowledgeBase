'use client'

import { useState } from 'react'
import { ChevronRight, Folder, FolderOpen } from 'lucide-react'
import { cn } from '@/lib/utils'

interface Category {
  id: string
  name: string
  children?: Category[]
}

// Mock data - will be replaced with API call
const mockCategories: Category[] = [
  {
    id: '1',
    name: 'Layer 1',
    children: [
      { id: '1-1', name: 'Ethereum' },
      { id: '1-2', name: 'Solana' },
      { id: '1-3', name: 'Cosmos' },
    ],
  },
  {
    id: '2',
    name: 'Layer 2',
    children: [
      { id: '2-1', name: 'ZK Rollup' },
      { id: '2-2', name: 'Optimistic Rollup' },
    ],
  },
  {
    id: '3',
    name: 'DeFi',
    children: [
      { id: '3-1', name: 'DEX' },
      { id: '3-2', name: 'Lending' },
      { id: '3-3', name: 'Staking' },
    ],
  },
  {
    id: '4',
    name: 'NFT',
  },
  {
    id: '5',
    name: '钱包与安全',
  },
]

interface CategoryNodeProps {
  category: Category
  level: number
}

function CategoryNode({ category, level }: CategoryNodeProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  const hasChildren = category.children && category.children.length > 0

  return (
    <div>
      <div
        className={cn(
          "flex items-center gap-1 px-2 py-1.5 rounded-md hover:bg-accent/10 cursor-pointer text-sm",
          "transition-colors"
        )}
        style={{ paddingLeft: `${level * 12 + 8}px` }}
        onClick={() => {
          if (hasChildren) {
            setIsExpanded(!isExpanded)
          }
        }}
      >
        {hasChildren ? (
          <ChevronRight
            className={cn(
              "w-4 h-4 text-muted-foreground transition-transform",
              isExpanded && "rotate-90"
            )}
          />
        ) : (
          <span className="w-4" />
        )}
        {hasChildren ? (
          isExpanded ? (
            <FolderOpen className="w-4 h-4 text-accent" />
          ) : (
            <Folder className="w-4 h-4 text-muted-foreground" />
          )
        ) : (
          <span className="w-4 h-4 flex items-center justify-center text-muted-foreground">
            &bull;
          </span>
        )}
        <span className="truncate">{category.name}</span>
      </div>
      {hasChildren && isExpanded && (
        <div>
          {category.children!.map((child) => (
            <CategoryNode key={child.id} category={child} level={level + 1} />
          ))}
        </div>
      )}
    </div>
  )
}

export function CategoryTree() {
  return (
    <div className="py-1">
      {mockCategories.map((category) => (
        <CategoryNode key={category.id} category={category} level={0} />
      ))}
    </div>
  )
}
