import { Component, type ReactNode } from "react";
import { QueryClient, useQueryClient } from "@tanstack/react-query";

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

class QueryErrorBoundaryInner extends Component<Props & { queryClient: QueryClient }, State> {
  state: State = { hasError: false, error: null };

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  handleRetry = () => {
    this.props.queryClient.resetQueries();
    this.setState({ hasError: false, error: null });
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center">
          <div className="text-center p-8 space-y-4 max-w-md">
            <h1 className="text-2xl font-bold text-red-400">Something went wrong</h1>
            <p className="text-gray-400">
              {this.state.error?.message || "An unexpected error occurred."}
            </p>
            <div className="flex gap-3 justify-center">
              <button
                onClick={this.handleRetry}
                className="bg-indigo-600 hover:bg-indigo-700 text-white font-semibold py-2 px-6 rounded-lg transition-colors"
              >
                Retry
              </button>
              <a
                href="/"
                className="bg-gray-700 hover:bg-gray-600 text-white font-semibold py-2 px-6 rounded-lg transition-colors inline-block"
              >
                Back to Lobby
              </a>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

export function QueryErrorBoundary({ children }: Props) {
  const queryClient = useQueryClient();
  return <QueryErrorBoundaryInner queryClient={queryClient}>{children}</QueryErrorBoundaryInner>;
}
