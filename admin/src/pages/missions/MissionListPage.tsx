import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { adminApi, Mission, MissionPack } from "../../api/admin";
import { DataTable } from "../../components/DataTable";
import { ImportDialog } from "../../components/ImportDialog";

export function MissionListPage() {
  const navigate = useNavigate();
  const [data, setData] = useState<Mission[]>([]);
  const [packs, setPacks] = useState<MissionPack[]>([]);
  const [filter, setFilter] = useState("");
  const [loading, setLoading] = useState(true);
  const [showImport, setShowImport] = useState(false);

  const load = () => {
    const params = filter ? { pack_id: filter } : undefined;
    adminApi.missions
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
        <h2 className="text-xl font-bold">Missions</h2>
        <div className="flex gap-2">
          <button
            onClick={() => setShowImport(true)}
            className="px-3 py-1.5 text-sm bg-gray-700 hover:bg-gray-600 rounded"
          >
            Import JSON
          </button>
          <button
            onClick={() => navigate("/missions/new")}
            className="px-3 py-1.5 text-sm bg-amber-600 hover:bg-amber-500 rounded"
          >
            Create
          </button>
        </div>
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
          { key: "scoringTiming", label: "Scoring Timing" },
          {
            key: "scoringRules",
            label: "Rules",
            render: (m) => <span>{m.scoringRules.length} rules</span>,
          },
        ]}
        data={data}
        getKey={(m) => m.id}
        searchField="name"
        onEdit={(m) => navigate(`/missions/${m.id}/edit`)}
        onDelete={async (m) => {
          await adminApi.missions.delete(m.id);
          load();
        }}
      />
      {showImport && (
        <ImportDialog
          title="Import Missions (JSON)"
          accept=".json"
          onImport={adminApi.import.missions}
          onClose={() => setShowImport(false)}
          onSuccess={load}
        />
      )}
    </div>
  );
}
