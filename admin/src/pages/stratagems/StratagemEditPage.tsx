import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { adminApi, Stratagem, Faction, Detachment } from '../../api/admin';

export function StratagemEditPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const isEdit = Boolean(id);

  const [form, setForm] = useState<Stratagem>({ id: '', factionId: '', detachmentId: '', name: '', type: '', cpCost: 1, legend: '', turn: '', phase: '', description: '' });
  const [factions, setFactions] = useState<Faction[]>([]);
  const [detachments, setDetachments] = useState<Detachment[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => { adminApi.factions.list().then(setFactions); }, []);
  useEffect(() => {
    if (form.factionId) {
      adminApi.detachments.list({ faction_id: form.factionId }).then(setDetachments);
    } else {
      setDetachments([]);
    }
  }, [form.factionId]);
  useEffect(() => {
    if (id) adminApi.stratagems.get(id).then(setForm).catch(() => navigate('/stratagems'));
  }, [id, navigate]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setError(null);
    try {
      if (isEdit) {
        await adminApi.stratagems.update(id!, form);
      } else {
        await adminApi.stratagems.create(form);
      }
      navigate('/stratagems');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Save failed');
    } finally {
      setSaving(false);
    }
  };

  const set = (field: keyof Stratagem, value: string | number) => setForm({ ...form, [field]: value });

  return (
    <div className="p-6 max-w-lg">
      <h2 className="text-xl font-bold mb-4">{isEdit ? 'Edit' : 'Create'} Stratagem</h2>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm text-gray-400 mb-1">ID</label>
          <input type="text" value={form.id} onChange={(e) => set('id', e.target.value)} disabled={isEdit} required className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm disabled:opacity-50" />
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Faction</label>
          <select value={form.factionId} onChange={(e) => set('factionId', e.target.value)} className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm">
            <option value="">None (Core)</option>
            {factions.map((f) => <option key={f.id} value={f.id}>{f.name}</option>)}
          </select>
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Detachment</label>
          <select value={form.detachmentId || ''} onChange={(e) => set('detachmentId', e.target.value)} className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm">
            <option value="">None</option>
            {detachments.map((d) => <option key={d.id} value={d.id}>{d.name}</option>)}
          </select>
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Name</label>
          <input type="text" value={form.name} onChange={(e) => set('name', e.target.value)} required className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm" />
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm text-gray-400 mb-1">Type</label>
            <select value={form.type} onChange={(e) => set('type', e.target.value)} className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm">
              <option value="">Select...</option>
              <option value="Battle Tactic">Battle Tactic</option>
              <option value="Epic Deed">Epic Deed</option>
              <option value="Strategic Ploy">Strategic Ploy</option>
              <option value="Wargear">Wargear</option>
            </select>
          </div>
          <div>
            <label className="block text-sm text-gray-400 mb-1">CP Cost</label>
            <input type="number" min={0} max={3} value={form.cpCost} onChange={(e) => set('cpCost', parseInt(e.target.value) || 0)} className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm" />
          </div>
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm text-gray-400 mb-1">Turn</label>
            <select value={form.turn} onChange={(e) => set('turn', e.target.value)} className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm">
              <option value="">Any</option>
              <option value="your">Your Turn</option>
              <option value="opponents">Opponent's Turn</option>
              <option value="either">Either Turn</option>
            </select>
          </div>
          <div>
            <label className="block text-sm text-gray-400 mb-1">Phase</label>
            <select value={form.phase} onChange={(e) => set('phase', e.target.value)} className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm">
              <option value="">Any</option>
              <option value="command">Command</option>
              <option value="movement">Movement</option>
              <option value="shooting">Shooting</option>
              <option value="charge">Charge</option>
              <option value="fight">Fight</option>
            </select>
          </div>
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Legend</label>
          <input type="text" value={form.legend || ''} onChange={(e) => set('legend', e.target.value)} className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm" />
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Description</label>
          <textarea value={form.description} onChange={(e) => set('description', e.target.value)} rows={4} className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm" />
        </div>
        {error && <p className="text-sm text-red-400">{error}</p>}
        <div className="flex gap-2">
          <button type="submit" disabled={saving} className="px-4 py-2 bg-amber-600 hover:bg-amber-500 rounded text-sm disabled:opacity-50">{saving ? 'Saving...' : 'Save'}</button>
          <button type="button" onClick={() => navigate('/stratagems')} className="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded text-sm">Cancel</button>
        </div>
      </form>
    </div>
  );
}
