'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { cn } from '@/lib/utils'
import {
  LayoutDashboard,
  Settings,
  Activity,
  DollarSign,
  FileText,
  ArrowLeft
} from 'lucide-react'

const navItems = [
  { href: '/admin', label: '概览', icon: LayoutDashboard },
  { href: '/admin/config', label: '配置', icon: Settings },
  { href: '/admin/tasks', label: '任务', icon: Activity },
  { href: '/admin/costs', label: '成本', icon: DollarSign },
  { href: '/admin/content', label: '内容', icon: FileText },
]

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()

  return (
    <div className="h-screen flex">
      {/* Sidebar */}
      <aside className="w-56 border-r border-border bg-background flex flex-col">
        <div className="h-14 flex items-center px-4 border-b border-border">
          <Link href="/" className="flex items-center gap-2 text-muted-foreground hover:text-foreground">
            <ArrowLeft className="w-4 h-4" />
            <span className="text-sm">返回主站</span>
          </Link>
        </div>

        <div className="px-4 py-4">
          <h1 className="font-semibold">后台管理</h1>
        </div>

        <nav className="flex-1 px-2">
          {navItems.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center gap-2 px-3 py-2 rounded-md text-sm mb-1",
                pathname === item.href
                  ? "bg-primary/10 text-primary"
                  : "hover:bg-muted text-muted-foreground"
              )}
            >
              <item.icon className="w-4 h-4" />
              {item.label}
            </Link>
          ))}
        </nav>
      </aside>

      {/* Main */}
      <main className="flex-1 overflow-auto bg-muted/30">
        {children}
      </main>
    </div>
  )
}
