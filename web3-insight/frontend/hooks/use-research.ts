'use client'

import { useState, useCallback, useEffect, useRef, useMemo } from 'react'
import { useRouter } from 'next/navigation'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  researchAPI,
  type ResearchSession,
  type ResearchSessionStatusResponse,
} from '@/lib/api'

const POLL_INTERVAL = 2000

const ACTIVE_STATUSES = new Set(['pending', 'planning', 'researching', 'writing'])

export function useResearch(sessionId?: string) {
  const router = useRouter()
  const queryClient = useQueryClient()
  const [isPolling, setIsPolling] = useState(false)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)

  // Fetch full session data
  const {
    data: session,
    isLoading: isLoadingSession,
    refetch: refetchSession,
  } = useQuery({
    queryKey: ['research-session', sessionId],
    queryFn: () => researchAPI.getSession(sessionId!),
    enabled: !!sessionId,
    staleTime: 10000,
  })

  // Lightweight status polling
  const {
    data: status,
  } = useQuery({
    queryKey: ['research-session-status', sessionId],
    queryFn: () => researchAPI.getSessionStatus(sessionId!),
    enabled: !!sessionId && isPolling,
    refetchInterval: isPolling ? POLL_INTERVAL : false,
  })

  // Start/stop polling based on session status
  useEffect(() => {
    if (!session && !status) return
    const currentStatus = status?.status ?? session?.status
    if (currentStatus && ACTIVE_STATUSES.has(currentStatus)) {
      setIsPolling(true)
    } else {
      setIsPolling(false)
      // Refetch full session when status transitions to terminal
      if (currentStatus === 'completed' || currentStatus === 'failed') {
        refetchSession()
      }
    }
  }, [status?.status, session?.status, refetchSession])

  // Cleanup polling on unmount
  useEffect(() => {
    return () => {
      if (pollRef.current) clearInterval(pollRef.current)
    }
  }, [])

  // Start a new research session
  const startMutation = useMutation({
    mutationFn: researchAPI.startSession,
    onSuccess: (resp) => {
      queryClient.invalidateQueries({ queryKey: ['research-sessions'] })
      router.push(`/research/${resp.sessionId}`)
    },
  })

  // Approve plan
  const approveMutation = useMutation({
    mutationFn: (editedPlan?: string) =>
      researchAPI.approvePlan(sessionId!, editedPlan),
    onSuccess: () => {
      setIsPolling(true)
      refetchSession()
    },
  })

  // Cancel session
  const cancelMutation = useMutation({
    mutationFn: () => researchAPI.cancelSession(sessionId!),
    onSuccess: () => {
      setIsPolling(false)
      refetchSession()
      queryClient.invalidateQueries({ queryKey: ['research-sessions'] })
    },
  })

  // Pin finding
  const pinMutation = useMutation({
    mutationFn: (content: string) => researchAPI.pinFinding(sessionId!, content),
    onSuccess: () => refetchSession(),
  })

  // Remove pin
  const unpinMutation = useMutation({
    mutationFn: (index: number) => researchAPI.removePin(sessionId!, index),
    onSuccess: () => refetchSession(),
  })

  // Integrate findings
  const integrateMutation = useMutation({
    mutationFn: () => researchAPI.integrateFindings(sessionId!),
    onSuccess: () => {
      setIsPolling(true)
      refetchSession()
    },
  })

  // Merge status into session for convenience
  const mergedStatus: ResearchSessionStatusResponse | undefined = status ?? (session ? {
    status: session.status,
    stage: session.stage,
    stageDetail: session.stageDetail,
    researchPlan: session.researchPlan,
    articleSlug: session.articleSlug,
    error: session.error,
  } : undefined)

  return {
    session,
    status: mergedStatus,
    isLoadingSession,
    isPolling,

    startResearch: startMutation.mutateAsync,
    isStarting: startMutation.isPending,

    approvePlan: approveMutation.mutate,
    isApproving: approveMutation.isPending,

    cancelResearch: cancelMutation.mutate,
    isCancelling: cancelMutation.isPending,

    pinFinding: pinMutation.mutate,
    unpinFinding: unpinMutation.mutate,

    integrateFindings: integrateMutation.mutate,
    isIntegrating: integrateMutation.isPending,
  }
}

// Hook for listing research sessions with load-more pagination
export function useResearchSessions(page: number = 1, limit: number = 20) {
  const queryClient = useQueryClient()
  const [pages, setPages] = useState<Record<number, ResearchSession[]>>({})

  const { data, isLoading, isError } = useQuery({
    queryKey: ['research-sessions', page, limit],
    queryFn: () => researchAPI.listSessions(page, limit),
    staleTime: 30000,
  })

  // Accumulate pages as they load
  useEffect(() => {
    if (data?.sessions) {
      setPages(prev => ({ ...prev, [page]: data.sessions }))
    }
  }, [data, page])

  // Reset accumulated pages when queries are invalidated (e.g. after delete)
  useEffect(() => {
    const unsubscribe = queryClient.getQueryCache().subscribe((event) => {
      if (event.type === 'removed' && (event.query.queryKey[0] as string) === 'research-sessions') {
        setPages({})
      }
    })
    return unsubscribe
  }, [queryClient])

  const sessions = useMemo(() => {
    const allSessions: ResearchSession[] = []
    const sortedPages = Object.keys(pages).map(Number).sort((a, b) => a - b)
    for (const p of sortedPages) {
      allSessions.push(...pages[p])
    }
    return allSessions
  }, [pages])

  const deleteMutation = useMutation({
    mutationFn: researchAPI.deleteSession,
    onSuccess: () => {
      setPages({})
      queryClient.invalidateQueries({ queryKey: ['research-sessions'] })
    },
  })

  return {
    sessions,
    total: data?.total ?? 0,
    isLoading,
    isError,
    deleteSession: deleteMutation.mutate,
    isDeleting: deleteMutation.isPending,
  }
}
