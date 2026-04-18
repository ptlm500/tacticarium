import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { adminApi, Detachment, Faction } from "../../api/admin";
import { DataTable } from "../../components/DataTable";
import { ImportDialog } from "../../components/ImportDialog";

export function DetachmentListPage() {
  const navigate = useNavigate();
  const [data, setData] = useState<Detachment[]>([]);
  const [factions, setFactions] = useState<Faction[]>([]);
  const [filter, setFilter] = useState("");
  const [gameModeFilter, setGameModeFilter] = useState("");
  const [loading, setLoading] = useState(true);
  const [showImport, setShowImport] = useState(false);

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

  const filtered = gameModeFilter
    ? data.filter((d) => (d.gameMode ?? "core") === gameModeFilter)
    : data;

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-bold">Detachments</h2>
        <div className="flex gap-2">
          <button
            onClick={() => setShowImport(true)}
            className="px-3 py-1.5 text-sm bg-gray-700 hover:bg-gray-600 rounded"
          >
            Import CSV
          </button>
          <button
            onClick={() => navigate("/detachments/new")}
            className="px-3 py-1.5 text-sm bg-amber-600 hover:bg-amber-500 rounded"
          >
            Create
          </button>
        </div>
      </div>
      <div className="flex gap-2 mb-4">
        <select
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          className="px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm"
        >
          <option value="">All Factions</option>
          {factions.map((f) => (
            <option key={f.id} value={f.id}>
              {f.name}
            </option>
          ))}
        </select>
        <select
          value={gameModeFilter}
          onChange={(e) => setGameModeFilter(e.target.value)}
          className="px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm"
        >
          <option value="">All Game Modes</option>
          <option value="core">Core</option>
          <option value="boarding_actions">Boarding Actions</option>
        </select>
      </div>
      <DataTable
        columns={[
          { key: "id", label: "ID" },
          { key: "factionId", label: "Faction" },
          { key: "name", label: "Name" },
          {
            key: "gameMode",
            label: "Game Mode",
            render: (d) => d.gameMode ?? "core",
          },
        ]}
        data={filtered}
        getKey={(d) => d.id}
        searchField="name"
        onEdit={(d) => navigate(`/detachments/${d.id}/edit`)}
        onDelete={async (d) => {
          await adminApi.detachments.delete(d.id);
          load();
        }}
      />
      {showImport && (
        <ImportDialog
          title="Import Detachments (CSV)"
          accept=".csv"
          onImport={adminApi.import.detachments}
          onClose={() => setShowImport(false)}
          onSuccess={load}
        />
      )}
    </div>
  );
}
