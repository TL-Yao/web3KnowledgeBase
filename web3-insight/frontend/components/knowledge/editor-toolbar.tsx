"use client"

import { useState, useCallback } from "react"
import type { Editor } from "@tiptap/react"
import { Toggle } from "@/components/ui/toggle"
import { Separator } from "@/components/ui/separator"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip"
import {
  Heading1,
  Heading2,
  Heading3,
  Bold,
  Italic,
  List,
  ListOrdered,
  Quote,
  Code2,
  Link,
  Table,
  Undo2,
  Redo2,
  Plus,
  Trash2,
} from "lucide-react"

interface EditorToolbarProps {
  editor: Editor | null
}

// Tiptap extension commands (toggleHeading, setLink, insertTable, etc.) aren't
// typed on ChainedCommands when extensions are configured in a different file.
// Use this helper for type-safe runtime calls.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function cmd(editor: Editor): any {
  return editor.chain().focus()
}
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function canCmd(editor: Editor): any {
  return editor.can().chain().focus()
}

function ToolbarToggle({
  pressed,
  onPressedChange,
  tooltip,
  children,
  disabled,
}: {
  pressed: boolean
  onPressedChange: () => void
  tooltip: string
  children: React.ReactNode
  disabled?: boolean
}) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Toggle
          size="sm"
          pressed={pressed}
          onPressedChange={onPressedChange}
          disabled={disabled}
        >
          {children}
        </Toggle>
      </TooltipTrigger>
      <TooltipContent>{tooltip}</TooltipContent>
    </Tooltip>
  )
}

function LinkPopover({ editor }: { editor: Editor }) {
  const [open, setOpen] = useState(false)
  const [url, setUrl] = useState("")

  const handleOpen = useCallback((isOpen: boolean) => {
    if (isOpen) {
      const existingUrl = editor.getAttributes("link").href || ""
      setUrl(existingUrl)
    }
    setOpen(isOpen)
  }, [editor])

  const handleSubmit = useCallback(() => {
    if (!url) {
      cmd(editor).extendMarkRange("link").unsetLink().run()
    } else {
      cmd(editor).extendMarkRange("link").setLink({ href: url }).run()
    }
    setOpen(false)
    setUrl("")
  }, [editor, url])

  return (
    <Popover open={open} onOpenChange={handleOpen}>
      <Tooltip>
        <TooltipTrigger asChild>
          <PopoverTrigger asChild>
            <Toggle
              size="sm"
              pressed={editor.isActive("link")}
              onPressedChange={() => handleOpen(!open)}
            >
              <Link className="size-4" />
            </Toggle>
          </PopoverTrigger>
        </TooltipTrigger>
        <TooltipContent>链接</TooltipContent>
      </Tooltip>
      <PopoverContent className="w-80" align="start">
        <form
          onSubmit={(e) => {
            e.preventDefault()
            handleSubmit()
          }}
          className="flex gap-2"
        >
          <Input
            placeholder="https://example.com"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            className="flex-1"
            autoFocus
          />
          <Button type="submit" size="sm">
            确定
          </Button>
        </form>
      </PopoverContent>
    </Popover>
  )
}

function TableControls({ editor }: { editor: Editor }) {
  const isInTable = editor.isActive("table")

  if (!isInTable) {
    return (
      <Tooltip>
        <TooltipTrigger asChild>
          <Toggle
            size="sm"
            pressed={false}
            onPressedChange={() =>
              cmd(editor).insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run()
            }
          >
            <Table className="size-4" />
          </Toggle>
        </TooltipTrigger>
        <TooltipContent>插入表格</TooltipContent>
      </Tooltip>
    )
  }

  return (
    <div className="flex items-center gap-0.5">
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={() => cmd(editor).addColumnAfter().run()}
          >
            <Plus className="size-3" />
            <span className="text-[10px]">列</span>
          </Button>
        </TooltipTrigger>
        <TooltipContent>添加列</TooltipContent>
      </Tooltip>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={() => cmd(editor).addRowAfter().run()}
          >
            <Plus className="size-3" />
            <span className="text-[10px]">行</span>
          </Button>
        </TooltipTrigger>
        <TooltipContent>添加行</TooltipContent>
      </Tooltip>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={() => cmd(editor).deleteTable().run()}
          >
            <Trash2 className="size-3" />
          </Button>
        </TooltipTrigger>
        <TooltipContent>删除表格</TooltipContent>
      </Tooltip>
    </div>
  )
}

export function EditorToolbar({ editor }: EditorToolbarProps) {
  if (!editor) return null

  return (
    <div className="flex flex-wrap items-center gap-0.5 border-b bg-background sticky top-0 z-10 px-2 py-1">
      {/* Headings */}
      <ToolbarToggle
        pressed={editor.isActive("heading", { level: 1 })}
        onPressedChange={() => cmd(editor).toggleHeading({ level: 1 }).run()}
        tooltip="标题 1"
      >
        <Heading1 className="size-4" />
      </ToolbarToggle>
      <ToolbarToggle
        pressed={editor.isActive("heading", { level: 2 })}
        onPressedChange={() => cmd(editor).toggleHeading({ level: 2 }).run()}
        tooltip="标题 2"
      >
        <Heading2 className="size-4" />
      </ToolbarToggle>
      <ToolbarToggle
        pressed={editor.isActive("heading", { level: 3 })}
        onPressedChange={() => cmd(editor).toggleHeading({ level: 3 }).run()}
        tooltip="标题 3"
      >
        <Heading3 className="size-4" />
      </ToolbarToggle>

      <Separator orientation="vertical" className="mx-1 h-6" />

      {/* Text formatting */}
      <ToolbarToggle
        pressed={editor.isActive("bold")}
        onPressedChange={() => cmd(editor).toggleBold().run()}
        tooltip="粗体"
      >
        <Bold className="size-4" />
      </ToolbarToggle>
      <ToolbarToggle
        pressed={editor.isActive("italic")}
        onPressedChange={() => cmd(editor).toggleItalic().run()}
        tooltip="斜体"
      >
        <Italic className="size-4" />
      </ToolbarToggle>

      <Separator orientation="vertical" className="mx-1 h-6" />

      {/* Lists */}
      <ToolbarToggle
        pressed={editor.isActive("bulletList")}
        onPressedChange={() => cmd(editor).toggleBulletList().run()}
        tooltip="无序列表"
      >
        <List className="size-4" />
      </ToolbarToggle>
      <ToolbarToggle
        pressed={editor.isActive("orderedList")}
        onPressedChange={() => cmd(editor).toggleOrderedList().run()}
        tooltip="有序列表"
      >
        <ListOrdered className="size-4" />
      </ToolbarToggle>

      <Separator orientation="vertical" className="mx-1 h-6" />

      {/* Blocks */}
      <ToolbarToggle
        pressed={editor.isActive("blockquote")}
        onPressedChange={() => cmd(editor).toggleBlockquote().run()}
        tooltip="引用"
      >
        <Quote className="size-4" />
      </ToolbarToggle>
      <ToolbarToggle
        pressed={editor.isActive("codeBlock")}
        onPressedChange={() => cmd(editor).toggleCodeBlock().run()}
        tooltip="代码块"
      >
        <Code2 className="size-4" />
      </ToolbarToggle>

      <Separator orientation="vertical" className="mx-1 h-6" />

      {/* Insert */}
      <LinkPopover editor={editor} />
      <TableControls editor={editor} />

      <Separator orientation="vertical" className="mx-1 h-6" />

      {/* History */}
      <ToolbarToggle
        pressed={false}
        onPressedChange={() => cmd(editor).undo().run()}
        tooltip="撤销"
        disabled={!canCmd(editor).undo().run()}
      >
        <Undo2 className="size-4" />
      </ToolbarToggle>
      <ToolbarToggle
        pressed={false}
        onPressedChange={() => cmd(editor).redo().run()}
        tooltip="重做"
        disabled={!canCmd(editor).redo().run()}
      >
        <Redo2 className="size-4" />
      </ToolbarToggle>
    </div>
  )
}
