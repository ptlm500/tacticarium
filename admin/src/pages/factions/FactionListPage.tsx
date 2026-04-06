import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { adminApi, Faction } from '../../api/admin';
import { DataTable } from '../../components/DataTable';
import { ImportDialog } from '../../components/ImportDialog';

export function FactionListPage() {
  const navigate = useNavigate();
  const [data, setData] = useState<Faction[]>([]);
  const [loading, setLoading] = useState(true);
  const [showImport, setShowImport] = useState(false);

  const load = () => {
    adminApi.factions.list().then(setData).finally(() => setLoading(false));
  };

  useEffect(load, []);

  if (loading) return <div className="p-6">Loading...</div>;

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-bold">Factions</h2>
        <div className="flex gap-2">
          <button onClick={() => setShowImport(true)} className="px-3 py-1.5 text-sm bg-gray-700 hover:bg-gray-600 rounded">
            Import CSV
          </button>
          <button onClick={() => navigate('/factions/new')} className="px-3 py-1.5 text-sm bg-amber-600 hover:bg-amber-500 rounded">
            Create
          </button>
        </div>
      </div>
      <DataTable
        columns={[
          { key: 'id', label: 'ID' },
          { key: 'name', label: 'Name' },
          { key: 'wahapediaLink', label: 'Link', render: (f) => f.wahapediaLink ? <a href={f.wahapediaLink} target="_blank" rel="noreferrer" className="text-blue-400 hover:underline text-xs truncate block max-w-xs">{f.wahapediaLink}</a> : '-' },
        ]}
        data={data}
        getKey={(f) => f.id}
        searchField="name"
        onEdit={(f) => navigate(`/factions/${f.id}/edit`)}
        onDelete={async (f) => { await adminApi.factions.delete(f.id); load(); }}
      />
      {showImport && (
        <ImportDialog
          title="Import Factions (CSV)"
          accept=".csv"
          onImport={adminApi.import.factions}
          onClose={() => setShowImport(false)}
          onSuccess={load}
        />
      )}
    </div>
  );
}
