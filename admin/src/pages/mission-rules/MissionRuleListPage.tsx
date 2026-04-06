import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { adminApi, MissionRule, MissionPack } from "../../api/admin";
import { DataTable } from "../../components/DataTable";

export function MissionRuleListPage() {
  const navigate = useNavigate();
  const [data, setData] = useState<MissionRule[]>([]);
  const [packs, setPacks] = useState<MissionPack[]>([]);
  const [filter, setFilter] = useState("");
  const [loading, setLoading] = useState(true);

  const load = () => {
    const params = filter ? { pack_id: filter } : undefined;
    adminApi.missionRules
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
        <h2 className="text-xl font-bold">Mission Rules</h2>
        <button
          onClick={() => navigate("/mission-rules/new")}
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
        ]}
        data={data}
        getKey={(r) => r.id}
        searchField="name"
        onEdit={(r) => navigate(`/mission-rules/${r.id}/edit`)}
        onDelete={async (r) => {
          await adminApi.missionRules.delete(r.id);
          load();
        }}
      />
    </div>
  );
}
