import { useState, useEffect, useMemo } from 'react';
import { useOutletContext } from 'react-router-dom';
import { RefreshCw, Save, CheckCircle2 } from 'lucide-react';
import { fetchAPI, type RoundContext } from '../../../lib/api';

interface AvailabilityRow {
  entity_type: string;
  entity_id: string;
  name: string;
  is_available: boolean | null;
  checked_in: boolean;
}

function EntityList({ title, rows = [], onToggle }: { title: string; rows: AvailabilityRow[]; onToggle: (row: AvailabilityRow) => void }) {
  return (
    <div className="card" style={{ flex: 1, minWidth: '280px' }}>
      <h3>{title}</h3>
      <div style={{ maxHeight: '420px', overflowY: 'auto' }}>
        {(rows || []).map(row => (
          <label key={row.entity_id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.55rem 0', borderBottom: '1px solid var(--border)', cursor: 'pointer' }}>
            <span style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <span>{row.name}</span>
              {row.checked_in && (
                <span className="badge badge-success" title="Checked in"><CheckCircle2 size={11} /> In</span>
              )}
            </span>
            <input
              type="checkbox"
              checked={row.is_available !== false}
              onChange={() => onToggle(row)}
            />
          </label>
        ))}
        {(!rows || rows.length === 0) && (
          <div style={{ padding: '1.5rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>Nothing registered yet.</div>
        )}
      </div>
    </div>
  );
}

export default function RoundAvailability() {
  const { slug, roundId, isReadOnly } = useOutletContext<RoundContext>();
  const [rows, setRows] = useState<AvailabilityRow[]>([]);
  const [saving, setSaving] = useState(false);
  const [syncing, setSyncing] = useState(false);

  const load = () => {
    fetchAPI(`/api/t/${slug}/rounds/${roundId}/availability`).then(d => setRows(d || [])).catch(console.error);
  };

  useEffect(() => {
    load();
  }, [slug, roundId]);

  const toggle = (row: AvailabilityRow) => {
    setRows(prev => prev.map(r =>
      r.entity_type === row.entity_type && r.entity_id === row.entity_id
        ? { ...r, is_available: r.is_available === false }
        : r
    ));
  };

  const save = async () => {
    setSaving(true);
    try {
      const updates = rows.map(r => ({ entity_type: r.entity_type, entity_id: r.entity_id, is_available: r.is_available !== false }));
      await fetchAPI(`/api/t/${slug}/rounds/${roundId}/availability`, 'PUT', updates);
      load();
    } catch (err: any) { alert(err.message); } finally { setSaving(false); }
  };

  const sync = async () => {
    setSyncing(true);
    try {
      await fetchAPI(`/api/t/${slug}/rounds/${roundId}/availability/sync`, 'POST');
      load();
    } catch (err: any) { alert(err.message); } finally { setSyncing(false); }
  };

  const teams = useMemo(() => rows.filter(r => r.entity_type === 'team'), [rows]);
  const adjudicators = useMemo(() => rows.filter(r => r.entity_type === 'adjudicator'), [rows]);

  return (
    <div>
      <p style={{ color: 'var(--text-mute)', fontSize: '0.85rem', marginTop: 0 }}>
        Unchecked teams and adjudicators are excluded when the draw for this round is generated.
      </p>

      {!isReadOnly && (
        <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem' }}>
          <button type="button" className="btn btn-secondary" onClick={sync} disabled={syncing}>
            <RefreshCw size={15} /> {syncing ? 'Syncing…' : 'Sync from Check-ins'}
          </button>
          <button type="button" className="btn btn-primary" onClick={save} disabled={saving}>
            <Save size={15} /> {saving ? 'Saving…' : 'Save Availability'}
          </button>
        </div>
      )}

      {isReadOnly && (
        <div className="badge badge-success" style={{ marginBottom: '1rem' }}>
          Archived record — availability editing is disabled.
        </div>
      )}

      <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap', maxWidth: '900px' }}>
        <EntityList title="Teams" rows={teams} onToggle={isReadOnly ? () => {} : toggle} />
        <EntityList title="Adjudicators" rows={adjudicators} onToggle={isReadOnly ? () => {} : toggle} />
      </div>
    </div>
  );
}
