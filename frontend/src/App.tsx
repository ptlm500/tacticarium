import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { AuthContext, useAuthProvider, useAuth } from "./hooks/useAuth";
import { ErrorBoundary } from "./components/ErrorBoundary";
import { LoginPage } from "./pages/LoginPage";
import { LobbyPage } from "./pages/LobbyPage";
import { GameSetupPage } from "./pages/GameSetupPage";
import { GamePage } from "./pages/GamePage";
import { GameHistoryPage } from "./pages/GameHistoryPage";
import { GameDetailPage } from "./pages/GameDetailPage";
import { JoinRedirect } from "./pages/JoinRedirect";
import { AuthCallbackPage } from "./pages/AuthCallbackPage";

function AuthGuard({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center">
        <p>Loading...</p>
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
      <AuthContext.Provider value={auth}>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/auth/callback" element={<AuthCallbackPage />} />
            <Route
              path="/"
              element={
                <AuthGuard>
                  <LobbyPage />
                </AuthGuard>
              }
            />
            <Route
              path="/game/:id/setup"
              element={
                <AuthGuard>
                  <GameSetupPage />
                </AuthGuard>
              }
            />
            <Route
              path="/game/:id"
              element={
                <AuthGuard>
                  <GamePage />
                </AuthGuard>
              }
            />
            <Route
              path="/history"
              element={
                <AuthGuard>
                  <GameHistoryPage />
                </AuthGuard>
              }
            />
            <Route
              path="/history/:id"
              element={
                <AuthGuard>
                  <GameDetailPage />
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
      </AuthContext.Provider>
    </ErrorBoundary>
  );
}

export default App;
