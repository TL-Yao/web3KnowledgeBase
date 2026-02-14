"use client"

import { useState, useEffect, useCallback, useRef } from "react"
import { useEditor, EditorContent } from "@tiptap/react"
import StarterKit from "@tiptap/starter-kit"
import { Link } from "@tiptap/extension-link"
import { Table } from "@tiptap/extension-table"
import { TableRow } from "@tiptap/extension-table-row"
import { TableCell } from "@tiptap/extension-table-cell"
import { TableHeader } from "@tiptap/extension-table-header"
import { CodeBlockLowlight } from "@tiptap/extension-code-block-lowlight"
import { Markdown } from "tiptap-markdown"
import { common, createLowlight } from "lowlight"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { EditorToolbar } from "./editor-toolbar"
import "./editor-styles.css"

const lowlight = createLowlight(common)

function getMarkdown(editor: ReturnType<typeof useEditor>): string {
  if (!editor) return ""
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return ((editor.storage as any).markdown as { getMarkdown: () => string }).getMarkdown()
}

interface ArticleEditorProps {
  article: { id: string; title: string; content: string; slug: string }
  onSave: (title: string, content: string) => Promise<void>
  onCancel: () => void
  isSaving: boolean
}

export function ArticleEditor({
  article,
  onSave,
  onCancel,
  isSaving,
}: ArticleEditorProps) {
  const [title, setTitle] = useState(article.title)
  const initialContentRef = useRef(article.content)
  const initialTitleRef = useRef(article.title)

  const editor = useEditor({
    immediatelyRender: false,
    extensions: [
      StarterKit.configure({
        codeBlock: false,
        heading: { levels: [1, 2, 3] },
      }),
      Link.configure({
        openOnClick: false,
        protocols: ["http", "https", "mailto"],
        HTMLAttributes: { class: "text-primary underline" },
      }),
      Table.configure({ resizable: false }),
      TableRow,
      TableCell,
      TableHeader,
      CodeBlockLowlight.configure({ lowlight }),
      Markdown.configure({
        html: false,
        transformCopiedText: true,
        transformPastedText: true,
      }),
    ],
    content: article.content,
    editorProps: {
      attributes: {
        class: "outline-none",
      },
    },
  })

  const isDirty = useCallback(() => {
    if (!editor) return false
    const currentMarkdown = getMarkdown(editor)
    return title !== initialTitleRef.current || currentMarkdown !== initialContentRef.current
  }, [editor, title])

  // Unsaved changes warning for browser navigation
  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (isDirty()) {
        e.preventDefault()
      }
    }
    window.addEventListener("beforeunload", handleBeforeUnload)
    return () => window.removeEventListener("beforeunload", handleBeforeUnload)
  }, [isDirty])

  const handleSave = useCallback(async () => {
    if (!editor) return
    const markdown = getMarkdown(editor)
    await onSave(title, markdown)
  }, [editor, title, onSave])

  const handleCancel = useCallback(() => {
    if (isDirty()) {
      if (!window.confirm("有未保存的更改，确定要放弃吗？")) {
        return
      }
    }
    onCancel()
  }, [isDirty, onCancel])

  return (
    <div className="flex flex-col h-full">
      {/* Fixed top: title + toolbar + actions */}
      <div className="shrink-0 border-b bg-background px-6 pt-6 pb-3">
        <div className="max-w-4xl mx-auto">
          <div className="flex items-center justify-between gap-4 mb-3">
            <Input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="text-3xl font-bold flex-1"
              placeholder="文章标题"
            />
            <div className="flex gap-3 shrink-0">
              <Button variant="outline" onClick={handleCancel} disabled={isSaving}>
                取消
              </Button>
              <Button onClick={handleSave} disabled={isSaving}>
                {isSaving ? "保存中..." : "保存"}
              </Button>
            </div>
          </div>
        </div>
        <div className="max-w-4xl mx-auto">
          <div className="tiptap-editor border rounded-t-md overflow-hidden">
            <EditorToolbar editor={editor} />
          </div>
        </div>
      </div>

      {/* Scrollable middle: editor content */}
      <div className="flex-1 overflow-y-auto min-h-0">
        <div className="max-w-4xl mx-auto px-6">
          <div className="tiptap-editor border border-t-0 rounded-b-md">
            <div className="prose prose-neutral max-w-none dark:prose-invert">
              <EditorContent editor={editor} />
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
