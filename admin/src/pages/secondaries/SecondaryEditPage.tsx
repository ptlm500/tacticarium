import { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { adminApi, Secondary, MissionPack, ScoringOption, DrawRestriction } from "../../api/admin";

const emptyOption: ScoringOption = { label: "", vp: 0, mode: "" };

export function SecondaryEditPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const isEdit = Boolean(id);

  const [form, setForm] = useState<Secondary>({
    id: "",
    missionPackId: "",
    name: "",
    lore: "",
    description: "",
    maxVp: 15,
    isFixed: false,
    scoringOptions: [],
  });
  const [packs, setPacks] = useState<MissionPack[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    adminApi.missionPacks.list().then(setPacks);
  }, []);
  useEffect(() => {
    if (id)
      adminApi.secondaries
        .get(id)
        .then(setForm)
        .catch(() => navigate("/secondaries"));
  }, [id, navigate]);

  const options = form.scoringOptions ?? [];

  const updateOption = (index: number, field: keyof ScoringOption, value: string | number) => {
    const updated = [...options];
    updated[index] = { ...updated[index], [field]: value };
    setForm({ ...form, scoringOptions: updated });
  };

  const addOption = () => setForm({ ...form, scoringOptions: [...options, { ...emptyOption }] });
  const removeOption = (index: number) =>
    setForm({ ...form, scoringOptions: options.filter((_, i) => i !== index) });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setError(null);
    try {
      if (isEdit) {
        await adminApi.secondaries.update(id!, form);
      } else {
        await adminApi.secondaries.create(form);
      }
      navigate("/secondaries");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Save failed");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="p-6 max-w-2xl">
      <h2 className="text-xl font-bold mb-4">{isEdit ? "Edit" : "Create"} Secondary</h2>
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
          <label className="block text-sm text-gray-400 mb-1">Lore</label>
          <textarea
            value={form.lore || ""}
            onChange={(e) => setForm({ ...form, lore: e.target.value })}
            rows={2}
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
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm text-gray-400 mb-1">Max VP</label>
            <input
              type="number"
              min={0}
              value={form.maxVp}
              onChange={(e) => setForm({ ...form, maxVp: parseInt(e.target.value) || 0 })}
              className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm"
            />
          </div>
          <div className="flex items-end pb-2">
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={form.isFixed}
                onChange={(e) => setForm({ ...form, isFixed: e.target.checked })}
                className="rounded"
              />
              Fixed Secondary
            </label>
          </div>
        </div>

        <div className="p-3 bg-gray-800 rounded border border-gray-700">
          <label className="flex items-center gap-2 text-sm mb-2">
            <input
              type="checkbox"
              checked={Boolean(form.drawRestriction)}
              onChange={(e) =>
                setForm({
                  ...form,
                  drawRestriction: e.target.checked ? { round: 1, mode: "mandatory" } : undefined,
                })
              }
              className="rounded"
            />
            Has draw restriction
          </label>
          <p className="text-xs text-gray-500 mb-2">
            Tactical mode only. Triggers the "When Drawn" rule on the given battle round.
          </p>
          {form.drawRestriction && (
            <div className="grid grid-cols-2 gap-2">
              <div>
                <label className="block text-xs text-gray-500 mb-1">Round</label>
                <input
                  type="number"
                  min={1}
                  value={form.drawRestriction.round}
                  onChange={(e) =>
                    setForm({
                      ...form,
                      drawRestriction: {
                        ...(form.drawRestriction as DrawRestriction),
                        round: parseInt(e.target.value) || 1,
                      },
                    })
                  }
                  className="w-full px-2 py-1 bg-gray-700 border border-gray-600 rounded text-xs"
                />
              </div>
              <div>
                <label className="block text-xs text-gray-500 mb-1">Mode</label>
                <select
                  value={form.drawRestriction.mode}
                  onChange={(e) =>
                    setForm({
                      ...form,
                      drawRestriction: {
                        ...(form.drawRestriction as DrawRestriction),
                        mode: e.target.value,
                      },
                    })
                  }
                  className="w-full px-2 py-1 bg-gray-700 border border-gray-600 rounded text-xs"
                >
                  <option value="mandatory">Mandatory (auto-reshuffle)</option>
                  <option value="optional">Optional (player choice)</option>
                </select>
              </div>
            </div>
          )}
        </div>

        <div>
          <div className="flex items-center justify-between mb-2">
            <label className="text-sm text-gray-400">Scoring Options</label>
            <button
              type="button"
              onClick={addOption}
              className="text-xs text-amber-400 hover:text-amber-300"
            >
              + Add Option
            </button>
          </div>
          {options.map((opt, i) => (
            <div key={i} className="mb-2 p-3 bg-gray-800 rounded border border-gray-700">
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs text-gray-500">Option {i + 1}</span>
                <button
                  type="button"
                  onClick={() => removeOption(i)}
                  className="text-xs text-red-400 hover:text-red-300"
                >
                  Remove
                </button>
              </div>
              <div className="grid grid-cols-3 gap-2">
                <div>
                  <label className="block text-xs text-gray-500 mb-1">Label</label>
                  <input
                    type="text"
                    value={opt.label}
                    onChange={(e) => updateOption(i, "label", e.target.value)}
                    className="w-full px-2 py-1 bg-gray-700 border border-gray-600 rounded text-xs"
                  />
                </div>
                <div>
                  <label className="block text-xs text-gray-500 mb-1">VP</label>
                  <input
                    type="number"
                    value={opt.vp}
                    onChange={(e) => updateOption(i, "vp", parseInt(e.target.value) || 0)}
                    className="w-full px-2 py-1 bg-gray-700 border border-gray-600 rounded text-xs"
                  />
                </div>
                <div>
                  <label className="block text-xs text-gray-500 mb-1">Mode</label>
                  <select
                    value={opt.mode || ""}
                    onChange={(e) => updateOption(i, "mode", e.target.value)}
                    className="w-full px-2 py-1 bg-gray-700 border border-gray-600 rounded text-xs"
                  >
                    <option value="">Both</option>
                    <option value="fixed">Fixed only</option>
                    <option value="tactical">Tactical only</option>
                  </select>
                </div>
              </div>
            </div>
          ))}
          {options.length === 0 && <p className="text-xs text-gray-500">No scoring options.</p>}
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
            onClick={() => navigate("/secondaries")}
            className="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded text-sm"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}
