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
import { StratagemEditPage } from "./pages/stratagems/StratagemEditPage";
import { MissionPackListPage } from "./pages/mission-packs/MissionPackListPage";
import { MissionPackEditPage } from "./pages/mission-packs/MissionPackEditPage";
import { MissionListPage } from "./pages/missions/MissionListPage";
import { MissionEditPage } from "./pages/missions/MissionEditPage";
import { SecondaryListPage } from "./pages/secondaries/SecondaryListPage";
import { SecondaryEditPage } from "./pages/secondaries/SecondaryEditPage";
import { GambitListPage } from "./pages/gambits/GambitListPage";
import { GambitEditPage } from "./pages/gambits/GambitEditPage";
import { ChallengerCardListPage } from "./pages/challenger-cards/ChallengerCardListPage";
import { ChallengerCardEditPage } from "./pages/challenger-cards/ChallengerCardEditPage";
import { MissionRuleListPage } from "./pages/mission-rules/MissionRuleListPage";
import { MissionRuleEditPage } from "./pages/mission-rules/MissionRuleEditPage";

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
            <Route path="/stratagems/new" element={<StratagemEditPage />} />
            <Route path="/stratagems/:id/edit" element={<StratagemEditPage />} />
            <Route path="/mission-packs" element={<MissionPackListPage />} />
            <Route path="/mission-packs/new" element={<MissionPackEditPage />} />
            <Route path="/mission-packs/:id/edit" element={<MissionPackEditPage />} />
            <Route path="/missions" element={<MissionListPage />} />
            <Route path="/missions/new" element={<MissionEditPage />} />
            <Route path="/missions/:id/edit" element={<MissionEditPage />} />
            <Route path="/secondaries" element={<SecondaryListPage />} />
            <Route path="/secondaries/new" element={<SecondaryEditPage />} />
            <Route path="/secondaries/:id/edit" element={<SecondaryEditPage />} />
            <Route path="/gambits" element={<GambitListPage />} />
            <Route path="/gambits/new" element={<GambitEditPage />} />
            <Route path="/gambits/:id/edit" element={<GambitEditPage />} />
            <Route path="/challenger-cards" element={<ChallengerCardListPage />} />
            <Route path="/challenger-cards/new" element={<ChallengerCardEditPage />} />
            <Route path="/challenger-cards/:id/edit" element={<ChallengerCardEditPage />} />
            <Route path="/mission-rules" element={<MissionRuleListPage />} />
            <Route path="/mission-rules/new" element={<MissionRuleEditPage />} />
            <Route path="/mission-rules/:id/edit" element={<MissionRuleEditPage />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </AuthContext.Provider>
  );
}

export default App;
