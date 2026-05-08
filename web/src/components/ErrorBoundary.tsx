import { Component, type ReactNode } from 'react'
import { AlertTriangle, RefreshCw } from 'lucide-react'

interface Props {
  children: ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('ErrorBoundary caught:', error, errorInfo)
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null })
  }

  render() {
    if (this.state.hasError) {
      return (
        <div
          className="flex flex-col items-center justify-center py-20 px-6"
          style={{ color: 'var(--color-foreground)' }}
        >
          <AlertTriangle
            size={40}
            className="mb-4"
            style={{ color: 'var(--color-warning)' }}
          />
          <h2
            className="text-lg font-semibold mb-2"
            style={{ fontFamily: 'var(--font-heading)' }}
          >
            页面加载出错
          </h2>
          <p
            className="text-sm mb-4 text-center max-w-md"
            style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
          >
            {this.state.error?.message || '发生了未知错误'}
          </p>
          <button
            onClick={this.handleReset}
            className="flex items-center gap-1.5 px-4 py-2 text-sm font-medium rounded-lg transition-colors cursor-pointer"
            style={{
              backgroundColor: 'var(--color-primary)',
              color: 'var(--color-primary-foreground)',
              fontFamily: 'var(--font-heading)',
            }}
          >
            <RefreshCw size={14} />
            重试
          </button>
        </div>
      )
    }

    return this.props.children
  }
}