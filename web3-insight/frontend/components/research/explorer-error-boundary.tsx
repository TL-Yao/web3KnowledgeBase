'use client'

import { Component, type ReactNode } from 'react'
import { AlertCircle } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'

interface Props {
  children: ReactNode
}

interface State {
  hasError: boolean
  error?: Error
}

export class ExplorerErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: any) {
    console.error('Explorer panel error:', error, errorInfo)
  }

  render() {
    if (this.state.hasError) {
      return (
        <Card>
          <CardContent className="py-12">
            <div className="flex flex-col items-center justify-center gap-4">
              <AlertCircle className="w-12 h-12 text-red-400" />
              <p className="text-lg font-medium text-red-600">浏览器调研面板出错</p>
              <p className="text-sm text-muted-foreground">
                {this.state.error?.message || '未知错误'}
              </p>
              <Button onClick={() => this.setState({ hasError: false })}>
                重试
              </Button>
            </div>
          </CardContent>
        </Card>
      )
    }

    return this.props.children
  }
}
