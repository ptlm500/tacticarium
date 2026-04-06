import { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { adminApi, Gambit, MissionPack } from "../../api/admin";

export function GambitEditPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const isEdit = Boolean(id);

  const [form, setForm] = useState<Gambit>({
    id: "",
    missionPackId: "",
    name: "",
    description: "",
    vpValue: 0,
  });
  const [packs, setPacks] = useState<MissionPack[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    adminApi.missionPacks.list().then(setPacks);
  }, []);
  useEffect(() => {
    if (id)
      adminApi.gambits
        .get(id)
        .then(setForm)
        .catch(() => navigate("/gambits"));
  }, [id, navigate]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setError(null);
    try {
      if (isEdit) await adminApi.gambits.update(id!, form);
      else await adminApi.gambits.create(form);
      navigate("/gambits");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Save failed");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="p-6 max-w-lg">
      <h2 className="text-xl font-bold mb-4">{isEdit ? "Edit" : "Create"} Gambit</h2>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm text-gray-400 mb-1">ID</label>
          <input
            type="text"
            value={form.id}
            onChange={(e) => setForm({ ...form, id: e.target.value })}
            disabled={isEdit}
            required
            className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm disabled:opacity-50"
          />
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Mission Pack</label>
          <select
            value={form.missionPackId}
            onChange={(e) => setForm({ ...form, missionPackId: e.target.value })}
            required
            className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm"
          >
            <option value="">Select...</option>
            {packs.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Name</label>
          <input
            type="text"
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
            required
            className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm"
          />
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">VP Value</label>
          <input
            type="number"
            min={0}
            value={form.vpValue}
            onChange={(e) => setForm({ ...form, vpValue: parseInt(e.target.value) || 0 })}
            className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm"
          />
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Description</label>
          <textarea
            value={form.description}
            onChange={(e) => setForm({ ...form, description: e.target.value })}
            rows={4}
            className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm"
          />
        </div>
        {error && <p className="text-sm text-red-400">{error}</p>}
        <div className="flex gap-2">
          <button
            type="submit"
            disabled={saving}
            className="px-4 py-2 bg-amber-600 hover:bg-amber-500 rounded text-sm disabled:opacity-50"
          >
            {saving ? "Saving..." : "Save"}
          </button>
          <button
            type="button"
            onClick={() => navigate("/gambits")}
            className="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded text-sm"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}
