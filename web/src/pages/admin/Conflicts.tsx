import { useState, useEffect } from 'react';
import { useParams, useOutletContext } from 'react-router-dom';
import { Plus, Trash2 } from 'lucide-react';
import { fetchAPI, type AdminContext } from '../../lib/api';

const entityTypes = [
  { value: 'adjudicator', label: 'Adjudicator' },
  { value: 'team', label: 'Team' },
  { value: 'speaker', label: 'Speaker' },
  { value: 'institution', label: 'Institution' },
];

interface EntityPickerProps {
  label: string;
  types: string[];
  entityType: string;
  onTypeChange: (t: string) => void;
  entityId: string;
  onEntityChange: (id: string) => void;
  teams: any[];
  adjudicators: any[];
  institutions: any[];
}

function EntityPicker({ label, types, entityType, onTypeChange, entityId, onEntityChange, teams, adjudicators, institutions }: EntityPickerProps) {
  const options: { id: string; name: string }[] =
    entityType === 'adjudicator' ? adjudicators.map(a => ({ id: a.id, name: a.name }))
      : entityType === 'team' ? teams.map(t => ({ id: t.id, name: t.name }))
        : entityType === 'speaker' ? teams.flatMap(t => (t.speakers || []).map((s: any) => ({ id: s.id, name: `${s.name} (${t.name})` })))
          : institutions.map(i => ({ id: i.id, name: i.name }));

  return (
    <div className="form-group" style={{ display: 'flex', gap: '0.5rem' }}>
      <select className="input select" style={{ width: '150px' }} value={entityType} onChange={e => { onTypeChange(e.target.value); onEntityChange(''); }} aria-label={`${label} type`}>
        {types.map(t => <option key={t} value={t}>{entityTypes.find(x => x.value === t)?.label}</option>)}
      </select>
      <select className="input select" value={entityId} onChange={e => onEntityChange(e.target.value)} aria-label={label}>
        <option value={`Select ${label}`}>Select {label}</option>
        {options.map(o => <option key={o.id} value={o.id}>{o.name}</option>)}
      </select>
    </div>
  );
}

export default function Conflicts() {
  const { slug } = useParams<{ slug: string }>();
  const { isReadOnly, isAssistant } = useOutletContext<AdminContext>();
  const isRestricted = isReadOnly || isAssistant;

  const [conflicts, setConflicts] = useState<any[]>([]);
  const [teams, setTeams] = useState<any[]>([]);
  const [adjudicators, setAdjudicators] = useState<any[]>([]);
  const [institutions, setInstitutions] = useState<any[]>([]);

  const [subjectType, setSubjectType] = useState('adjudicator');
  const [subjectId, setSubjectId] = useState('');
  const [targetType, setTargetType] = useState('team');
  const [targetId, setTargetId] = useState('');
  const [weight, setWeight] = useState('soft');

  const loadData = () => {
    fetchAPI(`/api/t/${slug}/conflicts`).then(d => setConflicts(d || [])).catch(console.error);
    fetchAPI(`/api/t/${slug}/teams`).then(d => setTeams(d || [])).catch(console.error);
    fetchAPI(`/api/t/${slug}/adjudicators`).then(d => setAdjudicators(d || [])).catch(console.error);
    fetchAPI(`/api/t/${slug}/institutions`).then(d => setInstitutions(d || [])).catch(console.error);
  };

  useEffect(() => {
    loadData();
  }, [slug]);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!subjectId || !targetId) {
      alert('Both entities must be selected.');
      return;
    }
    try {
      await fetchAPI(`/api/t/${slug}/conflicts`, 'POST', {
        subject_type: subjectType,
        subject_id: subjectId,
        target_type: targetType,
        target_id: targetId,
        weight
      });
      setSubjectId('');
      setTargetId('');
      loadData();
    } catch (err: any) { alert(err.message); }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Delete this conflict?")) return;
    try {
      await fetchAPI(`/api/t/${slug}/conflicts/${id}`, 'DELETE');
      loadData();
    } catch (err: any) { alert(err.message); }
  };

  const weightBadge = (w: string) => (
    <span style={{
      fontSize: '0.65rem',
      fontWeight: 600,
      padding: '2px 8px',
      borderRadius: '999px',
      border: `1px solid ${w === 'hard' ? '#dc2626' : 'var(--border)'}`,
      background: w === 'hard' ? '#dc2626' : 'transparent',
      color: w === 'hard' ? '#fff' : 'var(--text-mute)'
    }}>
      {w.toUpperCase()}
    </span>
  );

  return (
    <div>
      <h2>Conflicts</h2>
      <p style={{ color: 'var(--text-mute)', fontSize: '0.85rem', marginTop: '-0.5rem' }}>
        Declared clashes between adjudicators, teams, speakers, or institutions. Hard conflicts are avoided strictly during draw pairing; soft conflicts are minimized.
      </p>

      {!isRestricted && (
        <form onSubmit={handleCreate} className="card" style={{ maxWidth: '720px', marginBottom: '1.25rem' }}>
          <h3>Add Conflict</h3>
          <EntityPicker
            label="Subject"
            types={['adjudicator', 'team']}
            entityType={subjectType}
            onTypeChange={setSubjectType}
            entityId={subjectId}
            onEntityChange={setSubjectId}
            teams={teams}
            adjudicators={adjudicators}
            institutions={institutions}
          />
          <EntityPicker
            label="Target"
            types={entityTypes.map(t => t.value)}
            entityType={targetType}
            onTypeChange={setTargetType}
            entityId={targetId}
            onEntityChange={setTargetId}
            teams={teams}
            adjudicators={adjudicators}
            institutions={institutions}
          />
          <div className="form-group">
            <select className="input select" style={{ width: '150px' }} value={weight} onChange={e => setWeight(e.target.value)} aria-label="Weight">
              <option value="soft">Soft</option>
              <option value="hard">Hard</option>
            </select>
          </div>
          <button type="submit" className="btn btn-primary"><Plus size={16} /> Add Conflict</button>
        </form>
      )}

      <div className="card" style={{ maxWidth: '720px' }}>
        <h3>Existing Conflicts</h3>
        <div style={{ maxHeight: '420px', overflowY: 'auto' }}>
          {conflicts.map(c => (
            <div key={c.id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.5rem 0', borderBottom: '1px solid var(--border)' }}>
              <div>
                <div style={{ fontWeight: '500' }}>{c.subject_name} ↔ {c.target_name}</div>
                <div style={{ fontSize: '0.75rem', color: 'var(--text-mute)' }}>
                  {c.subject_type.replace('_', ' ')} → {c.target_type.replace('_', ' ')}
                </div>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                {weightBadge(c.weight)}
                {!isRestricted && (
                  <button className="btn btn-danger" style={{ padding: '2px 8px', fontSize: '0.75rem' }} onClick={() => handleDelete(c.id)}>
                    <Trash2 size={13} />
                  </button>
                )}
              </div>
            </div>
          ))}
          {conflicts.length === 0 && (
            <div style={{ padding: '2rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>No conflicts declared yet.</div>
          )}
        </div>
      </div>
    </div>
  );
}
