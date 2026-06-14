import { Link } from "react-router-dom";

const sections = [
  { to: "/factions", label: "Factions", desc: "Manage army factions" },
  { to: "/detachments", label: "Detachments", desc: "Manage faction detachments" },
  {
    to: "/stratagems",
    label: "Stratagems",
    desc: "View stratagems across factions and detachments",
  },
];

export function DashboardPage() {
  return (
    <div className="p-6">
      <h2 className="text-2xl font-bold mb-6">Dashboard</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {sections.map((s) => (
          <Link
            key={s.to}
            to={s.to}
            className="block p-4 bg-gray-800 rounded-lg border border-gray-700 hover:border-amber-500/50 transition-colors"
          >
            <h3 className="font-semibold text-amber-400">{s.label}</h3>
            <p className="text-sm text-gray-400 mt-1">{s.desc}</p>
          </Link>
        ))}
      </div>
    </div>
  );
}
