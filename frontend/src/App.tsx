import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { AuthContext, useAuthProvider, useAuth } from "./hooks/useAuth";
import { createQueryClient } from "./queryClient";
import { ErrorBoundary } from "./components/ErrorBoundary";
import { QueryErrorBoundary } from "./components/QueryErrorBoundary";
import { ThemeProvider } from "./components/ThemeProvider";
import { Toaster } from "./components/ui/sonner";
import { Spinner } from "./components/ui/spinner";
import { LoginPage } from "./pages/LoginPage";
import { LobbyPage } from "./pages/LobbyPage";
import { GameSetupPage } from "./pages/GameSetupPage";
import { GamePage } from "./pages/GamePage";
import { GameHistoryPage } from "./pages/GameHistoryPage";
import { GameDetailPage } from "./pages/GameDetailPage";
import { JoinRedirect } from "./pages/JoinRedirect";
import { AuthCallbackPage } from "./pages/AuthCallbackPage";

const queryClient = createQueryClient();

function AuthGuard({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div className="min-h-screen bg-background text-foreground flex items-center justify-center">
        <Spinner className="text-primary" />
      </div>
    );
  }

  if (!user) {
    return <Navigate to="/login" />;
  }

  return <>{children}</>;
}

function App() {
  const auth = useAuthProvider();

  return (
    <ErrorBoundary>
      <ThemeProvider>
        <AuthContext.Provider value={auth}>
          <QueryClientProvider client={queryClient}>
            <BrowserRouter>
              <Routes>
                <Route path="/login" element={<LoginPage />} />
                <Route path="/auth/callback" element={<AuthCallbackPage />} />
                <Route
                  path="/"
                  element={
                    <AuthGuard>
                      <QueryErrorBoundary>
                        <LobbyPage />
                      </QueryErrorBoundary>
                    </AuthGuard>
                  }
                />
                <Route
                  path="/game/:id/setup"
                  element={
                    <AuthGuard>
                      <QueryErrorBoundary>
                        <GameSetupPage />
                      </QueryErrorBoundary>
                    </AuthGuard>
                  }
                />
                <Route
                  path="/game/:id"
                  element={
                    <AuthGuard>
                      <QueryErrorBoundary>
                        <GamePage />
                      </QueryErrorBoundary>
                    </AuthGuard>
                  }
                />
                <Route
                  path="/history"
                  element={
                    <AuthGuard>
                      <QueryErrorBoundary>
                        <GameHistoryPage />
                      </QueryErrorBoundary>
                    </AuthGuard>
                  }
                />
                <Route
                  path="/history/:id"
                  element={
                    <AuthGuard>
                      <QueryErrorBoundary>
                        <GameDetailPage />
                      </QueryErrorBoundary>
                    </AuthGuard>
                  }
                />
                <Route
                  path="/join/:code"
                  element={
                    <AuthGuard>
                      <JoinRedirect />
                    </AuthGuard>
                  }
                />
              </Routes>
            </BrowserRouter>
            <Toaster position="top-right" />
            <ReactQueryDevtools initialIsOpen={false} />
          </QueryClientProvider>
        </AuthContext.Provider>
      </ThemeProvider>
    </ErrorBoundary>
  );
}

export default App;
