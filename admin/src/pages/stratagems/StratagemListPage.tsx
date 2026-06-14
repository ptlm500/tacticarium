import { useState, useEffect } from "react";
import { adminApi, Stratagem, Faction } from "../../api/admin";
import { DataTable } from "../../components/DataTable";
import { ImportDialog } from "../../components/ImportDialog";

export function StratagemListPage() {
  const [data, setData] = useState<Stratagem[]>([]);
  const [factions, setFactions] = useState<Faction[]>([]);
  const [filter, setFilter] = useState("");
  const [loading, setLoading] = useState(true);
  const [showImport, setShowImport] = useState(false);

  const load = () => {
    const params = filter ? { faction_id: filter } : undefined;
    adminApi.stratagems
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
        <h2 className="text-xl font-bold">Stratagems</h2>
        <div className="flex gap-2">
          <button
            onClick={() => setShowImport(true)}
            className="px-3 py-1.5 text-sm bg-gray-700 hover:bg-gray-600 rounded"
          >
            Import CSV
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
      </div>
      <DataTable
        columns={[
          { key: "name", label: "Name" },
          { key: "type", label: "Type", render: (s) => s.type ?? "-" },
          { key: "category", label: "Category", render: (s) => s.category ?? "-" },
          { key: "cpCost", label: "CP" },
          {
            key: "phases",
            label: "Phases",
            render: (s) => (s.phases ?? []).join(", ") || "-",
          },
          { key: "factionId", label: "Faction", render: (s) => s.factionId ?? "-" },
        ]}
        data={data}
        getKey={(s) => s.id}
        searchField="name"
      />
      {showImport && (
        <ImportDialog
          title="Import Stratagems (CSV)"
          accept=".csv"
          onImport={adminApi.import.stratagems}
          onClose={() => setShowImport(false)}
          onSuccess={load}
        />
      )}
    </div>
  );
}
