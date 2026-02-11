'use client'

import { useState, useEffect } from 'react'
import { MessageCircle } from 'lucide-react'

interface SidebarToggleProps {
  onClick: () => void
}

export function SidebarToggle({ onClick }: SidebarToggleProps) {
  const [showHint, setShowHint] = useState(true)

  useEffect(() => {
    const timer = setTimeout(() => setShowHint(false), 4000)
    return () => clearTimeout(timer)
  }, [])

  return (
    <div className="absolute right-0 top-1/2 -translate-y-1/2 flex flex-col items-end gap-1">
      <button
        onClick={onClick}
        className="w-9 h-12 rounded-l-md bg-primary text-primary-foreground shadow-md flex items-center justify-center hover:translate-x-0 translate-x-1 transition-transform"
        title="打开问答 (⌘/)"
      >
        <MessageCircle className="w-4 h-4" />
      </button>
      <span
        className={`text-[10px] text-muted-foreground bg-background/80 px-1.5 py-0.5 rounded mr-1 transition-opacity duration-500 ${showHint ? 'opacity-100' : 'opacity-0'}`}
      >
        ⌘/
      </span>
    </div>
  )
}
