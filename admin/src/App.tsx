import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { AuthContext, useAuthProvider, useAuth } from "./hooks/useAuth";
import { Layout } from "./components/Layout";
import { LoginPage } from "./pages/LoginPage";
import { AuthCallbackPage } from "./pages/AuthCallbackPage";
import { DashboardPage } from "./pages/DashboardPage";
import { FactionListPage } from "./pages/factions/FactionListPage";
import { FactionEditPage } from "./pages/factions/FactionEditPage";
import { DetachmentListPage } from "./pages/detachments/DetachmentListPage";
import { DetachmentEditPage } from "./pages/detachments/DetachmentEditPage";
import { StratagemListPage } from "./pages/stratagems/StratagemListPage";

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
    <AuthContext.Provider value={auth}>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/auth/callback" element={<AuthCallbackPage />} />
          <Route
            element={
              <AuthGuard>
                <Layout />
              </AuthGuard>
            }
          >
            <Route path="/" element={<DashboardPage />} />
            <Route path="/factions" element={<FactionListPage />} />
            <Route path="/factions/new" element={<FactionEditPage />} />
            <Route path="/factions/:id/edit" element={<FactionEditPage />} />
            <Route path="/detachments" element={<DetachmentListPage />} />
            <Route path="/detachments/new" element={<DetachmentEditPage />} />
            <Route path="/detachments/:id/edit" element={<DetachmentEditPage />} />
            <Route path="/stratagems" element={<StratagemListPage />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </AuthContext.Provider>
  );
}

export default App;
