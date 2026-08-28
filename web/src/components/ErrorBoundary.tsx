import { Component, type ErrorInfo, type ReactNode } from 'react';
import { AlertTriangle, RotateCcw, ArrowLeft } from 'lucide-react';

interface Props {
  children: ReactNode;
  fallbackTitle?: string;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export default class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null,
  };

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Uncaught error caught by ErrorBoundary:', error, errorInfo);
  }

  private handleReset = () => {
    this.setState({ hasError: false, error: null });
  };

  private handleGoBack = () => {
    window.history.back();
  };

  public render() {
    if (this.state.hasError) {
      return (
        <div style={{ padding: '2rem 1rem', maxWidth: '640px', margin: '2rem auto' }}>
          <div
            className="card"
            style={{
              border: '1px solid #fca5a5',
              background: '#fff5f5',
              padding: '2rem',
              textAlign: 'center',
            }}
          >
            <div
              style={{
                width: '48px',
                height: '48px',
                borderRadius: '50%',
                background: '#fee2e2',
                color: '#ef4444',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                margin: '0 auto 1rem auto',
              }}
            >
              <AlertTriangle size={24} />
            </div>

            <h3 style={{ margin: '0 0 0.5rem 0', color: '#991b1b', fontSize: '1.25rem' }}>
              {this.props.fallbackTitle || 'Something went wrong'}
            </h3>

            <p style={{ color: '#7f1d1d', fontSize: '0.9rem', marginBottom: '1.5rem', lineHeight: 1.5 }}>
              {this.state.error?.message || 'An unexpected error occurred while loading this view.'}
            </p>

            <div style={{ display: 'flex', gap: '0.75rem', justifyContent: 'center', flexWrap: 'wrap' }}>
              <button
                type="button"
                className="btn btn-secondary"
                onClick={this.handleGoBack}
                style={{ display: 'inline-flex', alignItems: 'center', gap: '0.35rem', fontSize: '0.85rem' }}
              >
                <ArrowLeft size={15} /> Go Back
              </button>
              <button
                type="button"
                className="btn btn-primary"
                onClick={this.handleReset}
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: '0.35rem',
                  fontSize: '0.85rem',
                  background: '#dc2626',
                  borderColor: '#dc2626',
                }}
              >
                <RotateCcw size={15} /> Try Again
              </button>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
