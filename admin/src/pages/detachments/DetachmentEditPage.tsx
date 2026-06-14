import { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { adminApi, Detachment, Faction } from "../../api/admin";

export function DetachmentEditPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const isEdit = Boolean(id);

  const [form, setForm] = useState<Detachment>({
    id: "",
    factionId: "",
    name: "",
    forceDispositions: [],
  });
  const [factions, setFactions] = useState<Faction[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    adminApi.factions.list().then(setFactions);
  }, []);
  useEffect(() => {
    if (id)
      adminApi.detachments
        .get(id)
        .then((d) => setForm({ ...d, forceDispositions: d.forceDispositions ?? [] }))
        .catch(() => navigate("/detachments"));
  }, [id, navigate]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setError(null);
    try {
      if (isEdit) {
        await adminApi.detachments.update(id!, form);
      } else {
        await adminApi.detachments.create(form);
      }
      navigate("/detachments");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Save failed");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="p-6 max-w-lg">
      <h2 className="text-xl font-bold mb-4">{isEdit ? "Edit" : "Create"} Detachment</h2>
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
          <label className="block text-sm text-gray-400 mb-1">Faction</label>
          <select
            value={form.factionId}
            onChange={(e) => setForm({ ...form, factionId: e.target.value })}
            required
            className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm"
          >
            <option value="">Select faction...</option>
            {factions.map((f) => (
              <option key={f.id} value={f.id}>
                {f.name}
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
          <label className="block text-sm text-gray-400 mb-1">Detachment Points</label>
          <input
            type="number"
            value={form.detachmentPoints ?? ""}
            onChange={(e) =>
              setForm({
                ...form,
                detachmentPoints: e.target.value === "" ? undefined : Number(e.target.value),
              })
            }
            className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm"
          />
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Force Dispositions</label>
          <textarea
            value={(form.forceDispositions ?? []).join("\n")}
            onChange={(e) =>
              setForm({
                ...form,
                forceDispositions: e.target.value
                  .split("\n")
                  .map((s) => s.trim())
                  .filter(Boolean),
              })
            }
            rows={3}
            className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm"
          />
          <p className="text-xs text-gray-500 mt-1">One disposition ID per line.</p>
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
            onClick={() => navigate("/detachments")}
            className="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded text-sm"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}
