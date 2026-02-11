'use client'

import { useState, useCallback, useRef, useEffect } from 'react'

interface UseResizeOptions {
  onResize: (delta: number) => void
  onResizeEnd?: () => void
}

export function useResize({ onResize, onResizeEnd }: UseResizeOptions) {
  const [isDragging, setIsDragging] = useState(false)
  const startXRef = useRef(0)
  const rafRef = useRef<number | null>(null)

  const handleMove = useCallback((clientX: number) => {
    if (rafRef.current !== null) return
    rafRef.current = requestAnimationFrame(() => {
      const delta = clientX - startXRef.current
      startXRef.current = clientX
      onResize(delta)
      rafRef.current = null
    })
  }, [onResize])

  const handleEnd = useCallback(() => {
    setIsDragging(false)
    if (rafRef.current !== null) {
      cancelAnimationFrame(rafRef.current)
      rafRef.current = null
    }
    document.body.classList.remove('cursor-col-resize', 'select-none')
    onResizeEnd?.()
  }, [onResizeEnd])

  useEffect(() => {
    if (!isDragging) return

    const onMouseMove = (e: MouseEvent) => handleMove(e.clientX)
    const onMouseUp = () => handleEnd()
    const onTouchMove = (e: TouchEvent) => {
      if (e.touches.length === 1) handleMove(e.touches[0].clientX)
    }
    const onTouchEnd = () => handleEnd()

    document.addEventListener('mousemove', onMouseMove)
    document.addEventListener('mouseup', onMouseUp)
    document.addEventListener('touchmove', onTouchMove)
    document.addEventListener('touchend', onTouchEnd)

    return () => {
      document.removeEventListener('mousemove', onMouseMove)
      document.removeEventListener('mouseup', onMouseUp)
      document.removeEventListener('touchmove', onTouchMove)
      document.removeEventListener('touchend', onTouchEnd)
    }
  }, [isDragging, handleMove, handleEnd])

  const handleMouseDown = useCallback((e: React.MouseEvent | React.TouchEvent) => {
    e.preventDefault()
    const clientX = 'touches' in e ? e.touches[0].clientX : e.clientX
    startXRef.current = clientX
    setIsDragging(true)
    document.body.classList.add('cursor-col-resize', 'select-none')
  }, [])

  return { isDragging, handleMouseDown }
}
