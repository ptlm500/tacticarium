import { Link } from 'react-router-dom';

const sections = [
  { to: '/factions', label: 'Factions', desc: 'Manage army factions and Wahapedia links' },
  { to: '/detachments', label: 'Detachments', desc: 'Manage faction detachments' },
  { to: '/stratagems', label: 'Stratagems', desc: 'Manage stratagems across factions and detachments' },
  { to: '/mission-packs', label: 'Mission Packs', desc: 'Manage mission pack collections' },
  { to: '/missions', label: 'Missions', desc: 'Manage primary missions and scoring rules' },
  { to: '/secondaries', label: 'Secondaries', desc: 'Manage secondary objectives and scoring options' },
  { to: '/gambits', label: 'Gambits', desc: 'Manage gambit cards' },
  { to: '/challenger-cards', label: 'Challenger Cards', desc: 'Manage challenger mode cards' },
  { to: '/mission-rules', label: 'Mission Rules', desc: 'Manage mission twists and rules' },
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
