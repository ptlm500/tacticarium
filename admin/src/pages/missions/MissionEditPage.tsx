import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { adminApi, Mission, MissionPack, ScoringAction } from '../../api/admin';

const emptyRule: ScoringAction = { label: '', vp: 0, minRound: 0, description: '', scoringTiming: '' };

export function MissionEditPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const isEdit = Boolean(id);

  const [form, setForm] = useState<Mission>({ id: '', missionPackId: '', name: '', lore: '', description: '', scoringRules: [], scoringTiming: 'end_of_command_phase' });
  const [packs, setPacks] = useState<MissionPack[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => { adminApi.missionPacks.list().then(setPacks); }, []);
  useEffect(() => {
    if (id) adminApi.missions.get(id).then(setForm).catch(() => navigate('/missions'));
  }, [id, navigate]);

  const updateRule = (index: number, field: keyof ScoringAction, value: string | number) => {
    const rules = [...form.scoringRules];
    rules[index] = { ...rules[index], [field]: value };
    setForm({ ...form, scoringRules: rules });
  };

  const addRule = () => setForm({ ...form, scoringRules: [...form.scoringRules, { ...emptyRule }] });
  const removeRule = (index: number) => setForm({ ...form, scoringRules: form.scoringRules.filter((_, i) => i !== index) });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setError(null);
    try {
      const data = {
        ...form,
        scoringRules: form.scoringRules.map((r) => ({
          ...r,
          minRound: r.minRound || 0,
        })),
      };
      if (isEdit) {
        await adminApi.missions.update(id!, data);
      } else {
        await adminApi.missions.create(data);
      }
      navigate('/missions');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Save failed');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="p-6 max-w-2xl">
      <h2 className="text-xl font-bold mb-4">{isEdit ? 'Edit' : 'Create'} Mission</h2>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm text-gray-400 mb-1">ID</label>
          <input type="text" value={form.id} onChange={(e) => setForm({ ...form, id: e.target.value })} disabled={isEdit} required className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm disabled:opacity-50" />
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Mission Pack</label>
          <select value={form.missionPackId} onChange={(e) => setForm({ ...form, missionPackId: e.target.value })} required className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm">
            <option value="">Select...</option>
            {packs.map((p) => <option key={p.id} value={p.id}>{p.name}</option>)}
          </select>
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Name</label>
          <input type="text" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm" />
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Lore</label>
          <textarea value={form.lore || ''} onChange={(e) => setForm({ ...form, lore: e.target.value })} rows={2} className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm" />
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Description</label>
          <textarea value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} rows={4} className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm" />
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Scoring Timing</label>
          <select value={form.scoringTiming} onChange={(e) => setForm({ ...form, scoringTiming: e.target.value })} className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm">
            <option value="end_of_command_phase">End of Command Phase</option>
            <option value="end_of_turn">End of Turn</option>
            <option value="end_of_round">End of Round</option>
          </select>
        </div>

        <div>
          <div className="flex items-center justify-between mb-2">
            <label className="text-sm text-gray-400">Scoring Rules</label>
            <button type="button" onClick={addRule} className="text-xs text-amber-400 hover:text-amber-300">+ Add Rule</button>
          </div>
          {form.scoringRules.map((rule, i) => (
            <div key={i} className="mb-3 p-3 bg-gray-800 rounded border border-gray-700">
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs text-gray-500">Rule {i + 1}</span>
                <button type="button" onClick={() => removeRule(i)} className="text-xs text-red-400 hover:text-red-300">Remove</button>
              </div>
              <div className="grid grid-cols-2 gap-2 mb-2">
                <div>
                  <label className="block text-xs text-gray-500 mb-1">Label</label>
                  <input type="text" value={rule.label} onChange={(e) => updateRule(i, 'label', e.target.value)} className="w-full px-2 py-1 bg-gray-700 border border-gray-600 rounded text-xs" />
                </div>
                <div>
                  <label className="block text-xs text-gray-500 mb-1">VP</label>
                  <input type="number" value={rule.vp} onChange={(e) => updateRule(i, 'vp', parseInt(e.target.value) || 0)} className="w-full px-2 py-1 bg-gray-700 border border-gray-600 rounded text-xs" />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-2 mb-2">
                <div>
                  <label className="block text-xs text-gray-500 mb-1">Min Round (0 = any)</label>
                  <input type="number" min={0} max={5} value={rule.minRound || 0} onChange={(e) => updateRule(i, 'minRound', parseInt(e.target.value) || 0)} className="w-full px-2 py-1 bg-gray-700 border border-gray-600 rounded text-xs" />
                </div>
                <div>
                  <label className="block text-xs text-gray-500 mb-1">Scoring Timing Override</label>
                  <input type="text" value={rule.scoringTiming || ''} onChange={(e) => updateRule(i, 'scoringTiming', e.target.value)} placeholder="(uses mission default)" className="w-full px-2 py-1 bg-gray-700 border border-gray-600 rounded text-xs" />
                </div>
              </div>
              <div>
                <label className="block text-xs text-gray-500 mb-1">Description</label>
                <input type="text" value={rule.description || ''} onChange={(e) => updateRule(i, 'description', e.target.value)} className="w-full px-2 py-1 bg-gray-700 border border-gray-600 rounded text-xs" />
              </div>
            </div>
          ))}
          {form.scoringRules.length === 0 && <p className="text-xs text-gray-500">No scoring rules. Click "+ Add Rule" to add one.</p>}
        </div>

        {error && <p className="text-sm text-red-400">{error}</p>}
        <div className="flex gap-2">
          <button type="submit" disabled={saving} className="px-4 py-2 bg-amber-600 hover:bg-amber-500 rounded text-sm disabled:opacity-50">{saving ? 'Saving...' : 'Save'}</button>
          <button type="button" onClick={() => navigate('/missions')} className="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded text-sm">Cancel</button>
        </div>
      </form>
    </div>
  );
}
