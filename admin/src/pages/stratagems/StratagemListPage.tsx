import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { adminApi, Stratagem, Faction } from '../../api/admin';
import { DataTable } from '../../components/DataTable';
import { ImportDialog } from '../../components/ImportDialog';

export function StratagemListPage() {
  const navigate = useNavigate();
  const [data, setData] = useState<Stratagem[]>([]);
  const [factions, setFactions] = useState<Faction[]>([]);
  const [filter, setFilter] = useState('');
  const [loading, setLoading] = useState(true);
  const [showImport, setShowImport] = useState(false);

  const load = () => {
    const params = filter ? { faction_id: filter } : undefined;
    adminApi.stratagems.list(params).then(setData).finally(() => setLoading(false));
  };

  useEffect(() => { adminApi.factions.list().then(setFactions); }, []);
  useEffect(load, [filter]);

  if (loading) return <div className="p-6">Loading...</div>;

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-bold">Stratagems</h2>
        <div className="flex gap-2">
          <button onClick={() => setShowImport(true)} className="px-3 py-1.5 text-sm bg-gray-700 hover:bg-gray-600 rounded">Import CSV</button>
          <button onClick={() => navigate('/stratagems/new')} className="px-3 py-1.5 text-sm bg-amber-600 hover:bg-amber-500 rounded">Create</button>
        </div>
      </div>
      <select value={filter} onChange={(e) => setFilter(e.target.value)} className="mb-4 px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm">
        <option value="">All Factions</option>
        {factions.map((f) => <option key={f.id} value={f.id}>{f.name}</option>)}
      </select>
      <DataTable
        columns={[
          { key: 'name', label: 'Name' },
          { key: 'type', label: 'Type' },
          { key: 'cpCost', label: 'CP' },
          { key: 'phase', label: 'Phase' },
          { key: 'factionId', label: 'Faction' },
        ]}
        data={data}
        getKey={(s) => s.id}
        searchField="name"
        onEdit={(s) => navigate(`/stratagems/${s.id}/edit`)}
        onDelete={async (s) => { await adminApi.stratagems.delete(s.id); load(); }}
      />
      {showImport && (
        <ImportDialog title="Import Stratagems (CSV)" accept=".csv" onImport={adminApi.import.stratagems} onClose={() => setShowImport(false)} onSuccess={load} />
      )}
    </div>
  );
}
