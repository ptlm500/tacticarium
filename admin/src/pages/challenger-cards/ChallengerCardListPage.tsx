import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { adminApi, ChallengerCard, MissionPack } from '../../api/admin';
import { DataTable } from '../../components/DataTable';

export function ChallengerCardListPage() {
  const navigate = useNavigate();
  const [data, setData] = useState<ChallengerCard[]>([]);
  const [packs, setPacks] = useState<MissionPack[]>([]);
  const [filter, setFilter] = useState('');
  const [loading, setLoading] = useState(true);

  const load = () => {
    const params = filter ? { pack_id: filter } : undefined;
    adminApi.challengerCards.list(params).then(setData).finally(() => setLoading(false));
  };

  useEffect(() => { adminApi.missionPacks.list().then(setPacks); }, []);
  useEffect(load, [filter]);

  if (loading) return <div className="p-6">Loading...</div>;

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-bold">Challenger Cards</h2>
        <button onClick={() => navigate('/challenger-cards/new')} className="px-3 py-1.5 text-sm bg-amber-600 hover:bg-amber-500 rounded">Create</button>
      </div>
      <select value={filter} onChange={(e) => setFilter(e.target.value)} className="mb-4 px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm">
        <option value="">All Mission Packs</option>
        {packs.map((p) => <option key={p.id} value={p.id}>{p.name}</option>)}
      </select>
      <DataTable
        columns={[
          { key: 'id', label: 'ID' },
          { key: 'name', label: 'Name' },
          { key: 'missionPackId', label: 'Pack' },
        ]}
        data={data}
        getKey={(c) => c.id}
        searchField="name"
        onEdit={(c) => navigate(`/challenger-cards/${c.id}/edit`)}
        onDelete={async (c) => { await adminApi.challengerCards.delete(c.id); load(); }}
      />
    </div>
  );
}
