import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { adminApi, Secondary, MissionPack } from "../../api/admin";
import { DataTable } from "../../components/DataTable";

export function SecondaryListPage() {
  const navigate = useNavigate();
  const [data, setData] = useState<Secondary[]>([]);
  const [packs, setPacks] = useState<MissionPack[]>([]);
  const [filter, setFilter] = useState("");
  const [loading, setLoading] = useState(true);

  const load = () => {
    const params = filter ? { pack_id: filter } : undefined;
    adminApi.secondaries
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
        <h2 className="text-xl font-bold">Secondaries</h2>
        <button
          onClick={() => navigate("/secondaries/new")}
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
          { key: "maxVp", label: "Max VP" },
          { key: "isFixed", label: "Fixed", render: (s) => (s.isFixed ? "Yes" : "No") },
          {
            key: "scoringTiming",
            label: "Timing",
            render: (s) =>
              (s.scoringTiming ?? "end_of_own_turn") === "end_of_opponent_turn"
                ? "Opp turn"
                : "Own turn",
          },
          {
            key: "scoringOptions",
            label: "Options",
            render: (s) => <span>{(s.scoringOptions ?? []).length}</span>,
          },
        ]}
        data={data}
        getKey={(s) => s.id}
        searchField="name"
        onEdit={(s) => navigate(`/secondaries/${s.id}/edit`)}
        onDelete={async (s) => {
          await adminApi.secondaries.delete(s.id);
          load();
        }}
      />
    </div>
  );
}
