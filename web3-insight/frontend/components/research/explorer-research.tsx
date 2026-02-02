'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  explorerAPI,
  type ExplorerResearch,
  type CreateExplorerRequest,
} from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Plus,
  ExternalLink,
  Loader2,
  Search,
  Edit,
  Trash2,
  CheckCircle,
  Clock,
  AlertCircle,
  BarChart3,
} from 'lucide-react'

const STATUS_CONFIG = {
  pending: { label: '待研究', icon: Clock, variant: 'outline' as const },
  in_progress: { label: '进行中', icon: AlertCircle, variant: 'secondary' as const },
  completed: { label: '已完成', icon: CheckCircle, variant: 'default' as const },
}

const CHAIN_TYPES = [
  { value: 'L1', label: 'Layer 1' },
  { value: 'L2', label: 'Layer 2' },
  { value: 'sidechain', label: 'Sidechain' },
  { value: 'appchain', label: 'App Chain' },
]

export function ExplorerResearchPanel() {
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [chainFilter, setChainFilter] = useState<string>('')
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false)
  const [editingExplorer, setEditingExplorer] = useState<ExplorerResearch | null>(null)
  const queryClient = useQueryClient()

  // Fetch explorers
  const { data: explorersData, isLoading } = useQuery({
    queryKey: ['explorers', chainFilter, statusFilter],
    queryFn: () => explorerAPI.list(chainFilter || undefined, statusFilter || undefined),
  })

  // Fetch chains for filter
  const { data: chainsData } = useQuery({
    queryKey: ['explorer-chains'],
    queryFn: explorerAPI.getChains,
  })

  // Fetch stats
  const { data: stats } = useQuery({
    queryKey: ['explorer-stats'],
    queryFn: explorerAPI.getStats,
  })

  // Mutations
  const createMutation = useMutation({
    mutationFn: (data: CreateExplorerRequest) => explorerAPI.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['explorers'] })
      queryClient.invalidateQueries({ queryKey: ['explorer-chains'] })
      queryClient.invalidateQueries({ queryKey: ['explorer-stats'] })
      setIsAddDialogOpen(false)
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: CreateExplorerRequest }) =>
      explorerAPI.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['explorers'] })
      setEditingExplorer(null)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => explorerAPI.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['explorers'] })
      queryClient.invalidateQueries({ queryKey: ['explorer-stats'] })
    },
  })

  const updateStatusMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      explorerAPI.updateStatus(id, status),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['explorers'] })
      queryClient.invalidateQueries({ queryKey: ['explorer-stats'] })
    },
  })

  const explorers = explorersData?.data || []
  const filteredExplorers = searchQuery
    ? explorers.filter(
        (e) =>
          e.chainName.toLowerCase().includes(searchQuery.toLowerCase()) ||
          e.explorerName.toLowerCase().includes(searchQuery.toLowerCase())
      )
    : explorers

  return (
    <div className="space-y-6">
      {/* Stats */}
      {stats && (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Total Explorers
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.total}</div>
            </CardContent>
          </Card>
          {Object.entries(STATUS_CONFIG).map(([status, config]) => (
            <Card key={status}>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                  <config.icon className="h-4 w-4" />
                  {config.label}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {stats.byStatus?.[status] || 0}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Controls */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center justify-between">
            <span className="flex items-center gap-2">
              <BarChart3 className="h-5 w-5" />
              Explorer Research
            </span>
            <Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
              <DialogTrigger asChild>
                <Button>
                  <Plus className="h-4 w-4 mr-2" />
                  Add Explorer
                </Button>
              </DialogTrigger>
              <DialogContent className="max-w-lg">
                <ExplorerForm
                  onSubmit={(data) => createMutation.mutate(data)}
                  isLoading={createMutation.isPending}
                />
              </DialogContent>
            </Dialog>
          </CardTitle>
          <CardDescription>
            Track and research blockchain explorers across different chains
          </CardDescription>
        </CardHeader>
        <CardContent>
          {/* Filters */}
          <div className="flex flex-wrap gap-4 mb-4">
            <div className="flex-1 min-w-[200px]">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search explorers..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-9"
                />
              </div>
            </div>
            <Select value={chainFilter} onValueChange={setChainFilter}>
              <SelectTrigger className="w-[150px]">
                <SelectValue placeholder="All chains" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">All chains</SelectItem>
                {chainsData?.chains.map((chain) => (
                  <SelectItem key={chain} value={chain}>
                    {chain}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-[150px]">
                <SelectValue placeholder="All statuses" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">All statuses</SelectItem>
                {Object.entries(STATUS_CONFIG).map(([value, config]) => (
                  <SelectItem key={value} value={value}>
                    {config.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Table */}
          {isLoading ? (
            <div className="flex justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin" />
            </div>
          ) : filteredExplorers.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No explorers found. Add your first explorer to start researching.
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Chain</TableHead>
                  <TableHead>Explorer</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Score</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredExplorers.map((explorer) => {
                  const statusConfig = STATUS_CONFIG[explorer.researchStatus]
                  return (
                    <TableRow key={explorer.id}>
                      <TableCell>
                        <div>
                          <div className="font-medium">{explorer.chainName}</div>
                          {explorer.chainType && (
                            <div className="text-xs text-muted-foreground">
                              {explorer.chainType}
                            </div>
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <span>{explorer.explorerName}</span>
                          <a
                            href={explorer.explorerUrl}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-muted-foreground hover:text-foreground"
                          >
                            <ExternalLink className="h-3 w-3" />
                          </a>
                        </div>
                      </TableCell>
                      <TableCell>
                        {explorer.explorerType && (
                          <Badge variant="outline">{explorer.explorerType}</Badge>
                        )}
                      </TableCell>
                      <TableCell>
                        <Select
                          value={explorer.researchStatus}
                          onValueChange={(status) =>
                            updateStatusMutation.mutate({ id: explorer.id, status })
                          }
                        >
                          <SelectTrigger className="w-[120px] h-8">
                            <Badge variant={statusConfig.variant}>
                              {statusConfig.label}
                            </Badge>
                          </SelectTrigger>
                          <SelectContent>
                            {Object.entries(STATUS_CONFIG).map(([value, config]) => (
                              <SelectItem key={value} value={value}>
                                {config.label}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </TableCell>
                      <TableCell>
                        {explorer.popularityScore > 0 && (
                          <span>{explorer.popularityScore.toFixed(1)}</span>
                        )}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex justify-end gap-2">
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => setEditingExplorer(explorer)}
                          >
                            <Edit className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => {
                              if (confirm('Delete this explorer?')) {
                                deleteMutation.mutate(explorer.id)
                              }
                            }}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Edit Dialog */}
      <Dialog
        open={!!editingExplorer}
        onOpenChange={(open) => !open && setEditingExplorer(null)}
      >
        <DialogContent className="max-w-lg">
          {editingExplorer && (
            <ExplorerForm
              initialData={editingExplorer}
              onSubmit={(data) =>
                updateMutation.mutate({ id: editingExplorer.id, data })
              }
              isLoading={updateMutation.isPending}
            />
          )}
        </DialogContent>
      </Dialog>
    </div>
  )
}

interface ExplorerFormProps {
  initialData?: ExplorerResearch
  onSubmit: (data: CreateExplorerRequest) => void
  isLoading: boolean
}

function ExplorerForm({ initialData, onSubmit, isLoading }: ExplorerFormProps) {
  const [formData, setFormData] = useState<CreateExplorerRequest>({
    chainName: initialData?.chainName || '',
    chainType: initialData?.chainType || '',
    explorerName: initialData?.explorerName || '',
    explorerUrl: initialData?.explorerUrl || '',
    explorerType: initialData?.explorerType || '',
    analysis: initialData?.analysis || '',
    researchNotes: initialData?.researchNotes || '',
    strengths: initialData?.strengths || [],
    weaknesses: initialData?.weaknesses || [],
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit(formData)
  }

  return (
    <form onSubmit={handleSubmit}>
      <DialogHeader>
        <DialogTitle>{initialData ? 'Edit Explorer' : 'Add Explorer'}</DialogTitle>
        <DialogDescription>
          {initialData
            ? 'Update explorer research details'
            : 'Add a new blockchain explorer to research'}
        </DialogDescription>
      </DialogHeader>

      <div className="grid gap-4 py-4">
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label htmlFor="chainName">Chain Name *</Label>
            <Input
              id="chainName"
              value={formData.chainName}
              onChange={(e) =>
                setFormData({ ...formData, chainName: e.target.value })
              }
              placeholder="Ethereum"
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="chainType">Chain Type</Label>
            <Select
              value={formData.chainType}
              onValueChange={(value) =>
                setFormData({ ...formData, chainType: value })
              }
            >
              <SelectTrigger>
                <SelectValue placeholder="Select type" />
              </SelectTrigger>
              <SelectContent>
                {CHAIN_TYPES.map((type) => (
                  <SelectItem key={type.value} value={type.value}>
                    {type.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label htmlFor="explorerName">Explorer Name *</Label>
            <Input
              id="explorerName"
              value={formData.explorerName}
              onChange={(e) =>
                setFormData({ ...formData, explorerName: e.target.value })
              }
              placeholder="Etherscan"
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="explorerType">Explorer Type</Label>
            <Select
              value={formData.explorerType}
              onValueChange={(value) =>
                setFormData({ ...formData, explorerType: value })
              }
            >
              <SelectTrigger>
                <SelectValue placeholder="Select type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="official">Official</SelectItem>
                <SelectItem value="third-party">Third-party</SelectItem>
                <SelectItem value="community">Community</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        <div className="space-y-2">
          <Label htmlFor="explorerUrl">Explorer URL *</Label>
          <Input
            id="explorerUrl"
            type="url"
            value={formData.explorerUrl}
            onChange={(e) =>
              setFormData({ ...formData, explorerUrl: e.target.value })
            }
            placeholder="https://etherscan.io"
            required
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="researchNotes">Research Notes</Label>
          <Textarea
            id="researchNotes"
            value={formData.researchNotes}
            onChange={(e) =>
              setFormData({ ...formData, researchNotes: e.target.value })
            }
            placeholder="Notes about this explorer..."
            rows={3}
          />
        </div>
      </div>

      <DialogFooter>
        <Button type="submit" disabled={isLoading}>
          {isLoading && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
          {initialData ? 'Update' : 'Add'} Explorer
        </Button>
      </DialogFooter>
    </form>
  )
}
