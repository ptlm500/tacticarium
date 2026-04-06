import { NavLink, Outlet } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";

const navItems = [
  { to: "/factions", label: "Factions" },
  { to: "/detachments", label: "Detachments" },
  { to: "/stratagems", label: "Stratagems" },
  { to: "/mission-packs", label: "Mission Packs" },
  { to: "/missions", label: "Missions" },
  { to: "/secondaries", label: "Secondaries" },
  { to: "/gambits", label: "Gambits" },
  { to: "/challenger-cards", label: "Challenger Cards" },
  { to: "/mission-rules", label: "Mission Rules" },
];

export function Layout() {
  const { user, logout } = useAuth();

  return (
    <div className="flex h-screen bg-gray-900 text-gray-100">
      <aside className="w-56 flex-shrink-0 bg-gray-800 border-r border-gray-700 flex flex-col">
        <div className="p-4 border-b border-gray-700">
          <h1 className="text-lg font-bold text-amber-400">Tacticarium</h1>
          <p className="text-xs text-gray-400">Admin Panel</p>
        </div>
        <nav className="flex-1 overflow-y-auto p-2">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                `block px-3 py-2 rounded text-sm ${
                  isActive ? "bg-amber-600/20 text-amber-400" : "text-gray-300 hover:bg-gray-700"
                }`
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
        <div className="p-3 border-t border-gray-700">
          <p className="text-xs text-gray-400 truncate">{user?.githubUser}</p>
          <button onClick={logout} className="mt-1 text-xs text-red-400 hover:text-red-300">
            Sign out
          </button>
        </div>
      </aside>
      <main className="flex-1 overflow-y-auto">
        <Outlet />
      </main>
    </div>
  );
}
