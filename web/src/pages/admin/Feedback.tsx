import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Plus, Trash2, Pencil, ArrowUp, ArrowDown, ChevronDown, ChevronUp } from 'lucide-react';
import { fetchAPI } from '../../lib/api';

const questionTypes = ['scale', 'text', 'checkbox', 'select'];

export default function Feedback() {
  const { slug } = useParams<{ slug: string }>();

  const [tab, setTab] = useState<'builder' | 'submissions'>('builder');

  const [questions, setQuestions] = useState<any[]>([]);
  const [editingId, setEditingId] = useState('');
  const [name, setName] = useState('');
  const [type, setType] = useState('scale');
  const [required, setRequired] = useState(false);
  const [fromType, setFromType] = useState('team');
  const [optionsText, setOptionsText] = useState('');

  const [rounds, setRounds] = useState<any[]>([]);
  const [roundId, setRoundId] = useState('');
  const [submissions, setSubmissions] = useState<any[]>([]);
  const [expanded, setExpanded] = useState<string>('');

  const loadQuestions = () => {
    fetchAPI(`/api/t/${slug}/feedback/questions`).then(d => setQuestions(d || [])).catch(console.error);
  };

  useEffect(() => {
    loadQuestions();
    fetchAPI(`/api/t/${slug}/rounds`).then(d => setRounds(d || [])).catch(console.error);
    fetchAPI(`/api/t/${slug}/feedback/submissions`).then(d => setSubmissions(d || [])).catch(console.error);
  }, [slug]);

  const resetForm = () => {
    setEditingId('');
    setName('');
    setType('scale');
    setRequired(false);
    setFromType('team');
    setOptionsText('');
  };

  const startEdit = (q: any) => {
    setEditingId(q.id);
    setName(q.name);
    setType(q.type);
    setRequired(!!q.required);
    setFromType(q.from_type);
    setOptionsText((q.options || []).join('\n'));
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      alert('Question name is required.');
      return;
    }
    const options = optionsText.split('\n').map(o => o.trim()).filter(Boolean);
    const body = {
      type,
      name,
      options,
      required,
      from_type: fromType,
      to_type: 'adjudicator'
    };
    try {
      if (editingId) {
        await fetchAPI(`/api/t/${slug}/feedback/questions/${editingId}`, 'PUT', body);
      } else {
        await fetchAPI(`/api/t/${slug}/feedback/questions`, 'POST', body);
      }
      resetForm();
      loadQuestions();
    } catch (err: any) { alert(err.message); }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Delete this question? Existing submissions keep their stored answers.')) return;
    try {
      await fetchAPI(`/api/t/${slug}/feedback/questions/${id}`, 'DELETE');
      if (editingId === id) resetForm();
      loadQuestions();
    } catch (err: any) { alert(err.message); }
  };

  const handleMove = async (id: string, direction: 'up' | 'down') => {
    try {
      await fetchAPI(`/api/t/${slug}/feedback/questions/${id}/move`, 'POST', { direction });
      loadQuestions();
    } catch (err: any) { alert(err.message); }
  };

  const handleRoundChange = async (rid: string) => {
    setRoundId(rid);
    setExpanded('');
    try {
      const url = rid
        ? `/api/t/${slug}/feedback/submissions?round_id=${encodeURIComponent(rid)}`
        : `/api/t/${slug}/feedback/submissions`;
      const d = await fetchAPI(url);
      setSubmissions(d || []);
    } catch (err: any) { alert(err.message); }
  };

  const qNameById = Object.fromEntries(questions.map(q => [q.id, q.name]));

  const tabBtn = (key: 'builder' | 'submissions', label: string) => (
    <button
      key={key}
      type="button"
      className={`btn ${tab === key ? 'btn-primary' : 'btn-secondary'}`}
      style={{ padding: '6px 14px', fontSize: '0.85rem' }}
      onClick={() => setTab(key)}
    >
      {label}
    </button>
  );

  return (
    <div>
      <h2>Feedback</h2>
      <p style={{ color: 'var(--text-mute)', fontSize: '0.85rem', marginTop: '-0.5rem' }}>
        Build feedback questionnaires and review participant and adjudicator submissions.
      </p>

      <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1.25rem' }}>
        {tabBtn('builder', 'Questionnaire Builder')}
        {tabBtn('submissions', 'Submissions')}
      </div>

      {tab === 'builder' && (
        <>
          <form onSubmit={handleSave} className="card" style={{ maxWidth: '720px', marginBottom: '1.25rem' }}>
            <h3>{editingId ? 'Edit Question' : 'Add Question'}</h3>
            <div className="form-group">
              <label className="label">Question Name</label>
              <input className="input" value={name} onChange={e => setName(e.target.value)} placeholder="e.g. Adjudicator rating" />
            </div>
            <div className="grid grid-cols-2" style={{ gap: '0.75rem' }}>
              <div className="form-group">
                <label className="label">Type</label>
                <select className="input select" value={type} onChange={e => setType(e.target.value)}>
                  {questionTypes.map(t => <option key={t} value={t}>{t}</option>)}
                </select>
              </div>
              <div className="form-group">
                <label className="label">Evaluated By</label>
                <select className="input select" value={fromType} onChange={e => setFromType(e.target.value)}>
                  <option value="team">Teams</option>
                  <option value="adjudicator">Adjudicators</option>
                </select>
              </div>
            </div>
            {(type === 'select' || type === 'scale') && (
              <div className="form-group">
                <label className="label">
                  Options (one per line{type === 'select' ? ', required' : '; optional numeric min/max for scale'})
                </label>
                <textarea className="input" rows={type === 'select' ? 4 : 2} value={optionsText} onChange={e => setOptionsText(e.target.value)} placeholder={type === 'select' ? 'Excellent\nGood\nFair\nPoor' : '0\n10'} />
              </div>
            )}
            <div className="form-group">
              <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', cursor: 'pointer' }}>
                <input type="checkbox" checked={required} onChange={e => setRequired(e.target.checked)} />
                <span>Required</span>
              </label>
            </div>
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <button type="submit" className="btn btn-primary"><Plus size={16} /> {editingId ? 'Save Changes' : 'Add Question'}</button>
              {editingId && (
                <button type="button" className="btn btn-secondary" onClick={resetForm}>Cancel</button>
              )}
            </div>
          </form>

          <div className="card" style={{ maxWidth: '720px' }}>
            <h3>Questions</h3>
            <div style={{ maxHeight: '420px', overflowY: 'auto' }}>
              {questions.map((q, i) => (
                <div key={q.id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.5rem 0', borderBottom: '1px solid var(--border)' }}>
                  <div>
                    <div style={{ fontWeight: '500' }}>
                      {q.name}{q.required && <span style={{ color: 'var(--danger)' }}> *</span>}
                    </div>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-mute)' }}>
                      {q.type} | by {q.from_type === 'team' ? 'teams' : 'adjudicators'} → {q.to_type}
                      {(q.options || []).length > 0 && ` | options: ${q.options.join(', ')}`}
                    </div>
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
                    <button className="btn btn-secondary" style={{ padding: '2px 8px' }} disabled={i === 0} onClick={() => handleMove(q.id, 'up')} aria-label="Move up"><ArrowUp size={13} /></button>
                    <button className="btn btn-secondary" style={{ padding: '2px 8px' }} disabled={i === questions.length - 1} onClick={() => handleMove(q.id, 'down')} aria-label="Move down"><ArrowDown size={13} /></button>
                    <button className="btn btn-secondary" style={{ padding: '2px 8px' }} onClick={() => startEdit(q)}><Pencil size={13} /></button>
                    <button className="btn btn-danger" style={{ padding: '2px 8px' }} onClick={() => handleDelete(q.id)}><Trash2 size={13} /></button>
                  </div>
                </div>
              ))}
              {questions.length === 0 && (
                <div style={{ padding: '2rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>No feedback questions defined yet.</div>
              )}
            </div>
          </div>
        </>
      )}

      {tab === 'submissions' && (
        <div className="card" style={{ maxWidth: '720px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
            <h3>Submissions</h3>
            <select className="input select" style={{ width: '220px' }} value={roundId} onChange={e => handleRoundChange(e.target.value)}>
              <option value="">All rounds</option>
              {rounds.map(r => <option key={r.id} value={r.id}>{r.name}</option>)}
            </select>
          </div>
          {submissions.length === 0 ? (
            <div style={{ padding: '2rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>No feedback submitted yet.</div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              {submissions.map(s => (
                <div key={s.id} style={{ border: '1px solid var(--border)', borderRadius: '8px', overflow: 'hidden' }}>
                  <button
                    type="button"
                    style={{ width: '100%', display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.65rem 0.85rem', background: 'rgba(0,0,0,0.02)', border: 'none', cursor: 'pointer', textAlign: 'left' }}
                    onClick={() => setExpanded(expanded === s.id ? '' : s.id)}
                  >
                    <span style={{ fontWeight: '500' }}>{s.source_name} → {s.target_name}</span>
                    <span style={{ display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
                      {s.score !== null && s.score !== undefined && (
                        <span className="badge badge-info">Score: {s.score.toFixed(1)}</span>
                      )}
                      {expanded === s.id ? <ChevronUp size={15} /> : <ChevronDown size={15} />}
                    </span>
                  </button>
                  {expanded === s.id && (
                    <div style={{ padding: '0.75rem 0.85rem', borderTop: '1px solid var(--border)' }}>
                      {Object.keys(s.answers || {}).length === 0 ? (
                        <div style={{ color: 'var(--text-mute)', fontSize: '0.85rem' }}>No answers recorded.</div>
                      ) : (
                        Object.entries(s.answers).map(([qid, val]) => (
                          <div key={qid} style={{ display: 'flex', gap: '0.75rem', padding: '0.25rem 0', fontSize: '0.85rem' }}>
                            <span style={{ color: 'var(--text-mute)', minWidth: '140px' }}>{qNameById[qid] || qid}</span>
                            <span>{String(val)}</span>
                          </div>
                        ))
                      )}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
