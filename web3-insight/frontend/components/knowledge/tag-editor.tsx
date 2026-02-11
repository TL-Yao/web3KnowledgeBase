'use client'

import { useState, useEffect, useRef } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { X, Plus } from 'lucide-react'
import { toast } from 'sonner'
import { articleAPI, tagAPI, TagSearchResult } from '@/lib/api'
import { cn } from '@/lib/utils'

interface TagEditorProps {
  articleId: string
  initialTags: string[]
  onTagClick?: (tag: string) => void
}

export function TagEditor({ articleId, initialTags, onTagClick }: TagEditorProps) {
  const [tags, setTags] = useState<string[]>(initialTags)
  const [isAdding, setIsAdding] = useState(false)
  const [inputValue, setInputValue] = useState('')
  const [suggestions, setSuggestions] = useState<TagSearchResult[]>([])
  const [highlightedIndex, setHighlightedIndex] = useState(-1)
  const inputRef = useRef<HTMLInputElement>(null)
  const dropdownRef = useRef<HTMLDivElement>(null)
  const queryClient = useQueryClient()

  useEffect(() => {
    setTags(initialTags)
  }, [initialTags])

  const updateMutation = useMutation({
    mutationFn: (newTags: string[]) => articleAPI.updateTags(articleId, newTags),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['article'] })
      queryClient.invalidateQueries({ queryKey: ['articles'] })
      queryClient.invalidateQueries({ queryKey: ['tags', 'in-use'] })
    },
    onError: () => {
      setTags(initialTags)
      toast.error('更新标签失败')
    },
  })

  // Debounced search
  useEffect(() => {
    if (inputValue.length === 0) {
      setSuggestions([])
      return
    }
    const timer = setTimeout(async () => {
      try {
        const res = await tagAPI.search(inputValue)
        setSuggestions(res.tags.filter(t => !tags.includes(t.name)))
      } catch {
        setSuggestions([])
      }
    }, 300)
    return () => clearTimeout(timer)
  }, [inputValue, tags])

  // Close dropdown on outside click
  useEffect(() => {
    if (!isAdding) return
    const handleClick = (e: MouseEvent) => {
      if (
        dropdownRef.current && !dropdownRef.current.contains(e.target as Node) &&
        inputRef.current && !inputRef.current.contains(e.target as Node)
      ) {
        setIsAdding(false)
        setInputValue('')
        setSuggestions([])
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [isAdding])

  const addTag = (name: string) => {
    const trimmed = name.trim()
    if (!trimmed || tags.includes(trimmed)) return
    const newTags = [...tags, trimmed]
    setTags(newTags)
    updateMutation.mutate(newTags)
    setInputValue('')
    setSuggestions([])
    setHighlightedIndex(-1)
    setIsAdding(false)
  }

  const createAndAddTag = async (name: string) => {
    const trimmed = name.trim()
    if (!trimmed || tags.includes(trimmed)) return
    try {
      await tagAPI.create({ name: trimmed })
      addTag(trimmed)
      toast.success(`标签「${trimmed}」已创建`)
    } catch (err: unknown) {
      if (err instanceof Error && err.message?.includes('already exists')) {
        addTag(trimmed)
      } else {
        toast.error('创建标签失败')
      }
    }
  }

  const removeTag = (name: string) => {
    const newTags = tags.filter(t => t !== name)
    setTags(newTags)
    updateMutation.mutate(newTags)
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      if (highlightedIndex >= 0 && highlightedIndex < suggestions.length) {
        addTag(suggestions[highlightedIndex].name)
      } else if (inputValue.trim()) {
        createAndAddTag(inputValue)
      }
    } else if (e.key === 'Escape') {
      setIsAdding(false)
      setInputValue('')
      setSuggestions([])
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      setHighlightedIndex(prev =>
        prev < suggestions.length - 1 ? prev + 1 : prev
      )
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setHighlightedIndex(prev => (prev > 0 ? prev - 1 : -1))
    }
  }

  const startAdding = () => {
    setIsAdding(true)
    setTimeout(() => inputRef.current?.focus(), 0)
  }

  return (
    <div className="flex items-center gap-2 flex-wrap">
      {tags.map((tag) => (
        <Badge
          key={tag}
          variant="secondary"
          className="group gap-1 pr-1 cursor-pointer hover:bg-primary/20 transition-colors"
        >
          <span onClick={() => onTagClick?.(tag)}>{tag}</span>
          <button
            onClick={(e) => {
              e.stopPropagation()
              removeTag(tag)
            }}
            className="ml-0.5 rounded-full p-0.5 hover:bg-muted-foreground/20 opacity-0 group-hover:opacity-100 transition-opacity"
          >
            <X className="w-3 h-3" />
          </button>
        </Badge>
      ))}

      {isAdding ? (
        <div className="relative">
          <Input
            ref={inputRef}
            value={inputValue}
            onChange={(e) => {
              setInputValue(e.target.value)
              setHighlightedIndex(-1)
            }}
            onKeyDown={handleKeyDown}
            placeholder="输入标签名..."
            className="h-7 w-48 text-xs"
          />
          {(suggestions.length > 0 || (inputValue.length > 0 && suggestions.length === 0)) && (
            <div
              ref={dropdownRef}
              className="absolute top-full left-0 mt-1 w-64 bg-popover border border-border rounded-md shadow-md z-50 max-h-48 overflow-y-auto"
            >
              {suggestions.length > 0 ? (
                suggestions.map((s, i) => (
                  <button
                    key={s.name}
                    onClick={() => addTag(s.name)}
                    className={cn(
                      "w-full text-left px-3 py-1.5 text-sm hover:bg-accent transition-colors flex items-center justify-between",
                      i === highlightedIndex && "bg-accent"
                    )}
                  >
                    <span>{s.name}</span>
                    {s.nameEn && (
                      <span className="text-xs text-muted-foreground ml-2">{s.nameEn}</span>
                    )}
                  </button>
                ))
              ) : (
                <button
                  onClick={() => createAndAddTag(inputValue)}
                  className="w-full text-left px-3 py-1.5 text-sm hover:bg-accent transition-colors"
                >
                  创建标签「{inputValue.trim()}」
                </button>
              )}
            </div>
          )}
        </div>
      ) : (
        <Button
          variant="ghost"
          size="xs"
          onClick={startAdding}
          className="text-muted-foreground hover:text-foreground"
        >
          <Plus className="w-3 h-3" />
          添加
        </Button>
      )}
    </div>
  )
}
