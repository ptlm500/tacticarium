import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { adminApi, Gambit, MissionPack } from "../../api/admin";
import { DataTable } from "../../components/DataTable";

export function GambitListPage() {
  const navigate = useNavigate();
  const [data, setData] = useState<Gambit[]>([]);
  const [packs, setPacks] = useState<MissionPack[]>([]);
  const [filter, setFilter] = useState("");
  const [loading, setLoading] = useState(true);

  const load = () => {
    const params = filter ? { pack_id: filter } : undefined;
    adminApi.gambits
      .list(params)
      .then(setData)
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    adminApi.missionPacks.list().then(setPacks);
  }, []);
  useEffect(load, [filter]);

  if (loading) return <div className="p-6">Loading...</div>;

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-bold">Gambits</h2>
        <button
          onClick={() => navigate("/gambits/new")}
          className="px-3 py-1.5 text-sm bg-amber-600 hover:bg-amber-500 rounded"
        >
          Create
        </button>
      </div>
      <select
        value={filter}
        onChange={(e) => setFilter(e.target.value)}
        className="mb-4 px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm"
      >
        <option value="">All Mission Packs</option>
        {packs.map((p) => (
          <option key={p.id} value={p.id}>
            {p.name}
          </option>
        ))}
      </select>
      <DataTable
        columns={[
          { key: "id", label: "ID" },
          { key: "name", label: "Name" },
          { key: "missionPackId", label: "Pack" },
          { key: "vpValue", label: "VP Value" },
        ]}
        data={data}
        getKey={(g) => g.id}
        searchField="name"
        onEdit={(g) => navigate(`/gambits/${g.id}/edit`)}
        onDelete={async (g) => {
          await adminApi.gambits.delete(g.id);
          load();
        }}
      />
    </div>
  );
}
