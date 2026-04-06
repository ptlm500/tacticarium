import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { adminApi, MissionPack } from "../../api/admin";
import { DataTable } from "../../components/DataTable";

export function MissionPackListPage() {
  const navigate = useNavigate();
  const [data, setData] = useState<MissionPack[]>([]);
  const [loading, setLoading] = useState(true);

  const load = () => {
    adminApi.missionPacks
      .list()
      .then(setData)
      .finally(() => setLoading(false));
  };

  useEffect(load, []);

  if (loading) return <div className="p-6">Loading...</div>;

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-bold">Mission Packs</h2>
        <button
          onClick={() => navigate("/mission-packs/new")}
          className="px-3 py-1.5 text-sm bg-amber-600 hover:bg-amber-500 rounded"
        >
          Create
        </button>
      </div>
      <DataTable
        columns={[
          { key: "id", label: "ID" },
          { key: "name", label: "Name" },
          { key: "description", label: "Description" },
        ]}
        data={data}
        getKey={(p) => p.id}
        searchField="name"
        onEdit={(p) => navigate(`/mission-packs/${p.id}/edit`)}
        onDelete={async (p) => {
          await adminApi.missionPacks.delete(p.id);
          load();
        }}
      />
    </div>
  );
}
