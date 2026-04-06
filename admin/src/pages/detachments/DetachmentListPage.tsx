import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { adminApi, Detachment, Faction } from "../../api/admin";
import { DataTable } from "../../components/DataTable";

export function DetachmentListPage() {
  const navigate = useNavigate();
  const [data, setData] = useState<Detachment[]>([]);
  const [factions, setFactions] = useState<Faction[]>([]);
  const [filter, setFilter] = useState("");
  const [loading, setLoading] = useState(true);

  const load = () => {
    const params = filter ? { faction_id: filter } : undefined;
    adminApi.detachments
      .list(params)
      .then(setData)
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    adminApi.factions.list().then(setFactions);
  }, []);
  useEffect(load, [filter]);

  if (loading) return <div className="p-6">Loading...</div>;

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-bold">Detachments</h2>
        <button
          onClick={() => navigate("/detachments/new")}
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
        <option value="">All Factions</option>
        {factions.map((f) => (
          <option key={f.id} value={f.id}>
            {f.name}
          </option>
        ))}
      </select>
      <DataTable
        columns={[
          { key: "id", label: "ID" },
          { key: "factionId", label: "Faction" },
          { key: "name", label: "Name" },
        ]}
        data={data}
        getKey={(d) => d.id}
        searchField="name"
        onEdit={(d) => navigate(`/detachments/${d.id}/edit`)}
        onDelete={async (d) => {
          await adminApi.detachments.delete(d.id);
          load();
        }}
      />
    </div>
  );
}
