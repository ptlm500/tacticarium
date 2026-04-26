import { Component, type ReactNode } from "react";
import { QueryClient, useQueryClient } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";

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
    void this.props.queryClient.resetQueries();
    this.setState({ hasError: false, error: null });
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex min-h-screen items-center justify-center bg-background text-foreground">
          <div className="max-w-md space-y-4 p-8 text-center">
            <h1 className="text-2xl font-bold text-destructive">Something went wrong</h1>
            <p className="text-muted-foreground">
              {this.state.error?.message || "An unexpected error occurred."}
            </p>
            <div className="flex justify-center gap-3">
              <Button
                type="button"
                onClick={this.handleRetry}
                className="font-mono uppercase tracking-widest"
              >
                Retry
              </Button>
              <Button asChild variant="outline" className="font-mono uppercase tracking-widest">
                <a href="/">Back to Lobby</a>
              </Button>
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
