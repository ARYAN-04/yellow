import { useState, useEffect } from 'react';
import { MessageSquare } from 'lucide-react';
import { fetchAPI } from '../lib/api';

interface FeedbackTargetData {
  questions: any[];
  targets: any[];
}

function QuestionInput({ q, value, onChange }: { q: any; value: string; onChange: (v: string) => void }) {
  switch (q.type) {
    case 'scale':
      return (
        <input
          className="input"
          type="number"
          step="0.5"
          min="0"
          max="10"
          value={value}
          onChange={e => onChange(e.target.value)}
        />
      );
    case 'text':
      return <textarea className="input" rows={3} value={value} onChange={e => onChange(e.target.value)} />;
    case 'checkbox':
      return (
        <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', cursor: 'pointer' }}>
          <input
            type="checkbox"
            checked={value === 'true'}
            onChange={e => onChange(e.target.checked ? 'true' : 'false')}
          />
          <span>Yes</span>
        </label>
      );
    case 'select':
      return (
        <select className="input select" value={value} onChange={e => onChange(e.target.value)}>
          <option value="">Select an option</option>
          {(q.options || []).map((o: string) => (
            <option key={o} value={o}>{o}</option>
          ))}
        </select>
      );
    default:
      return null;
  }
}

export default function FeedbackSection({ token }: { token: string }) {
  const [data, setData] = useState<FeedbackTargetData | null>(null);
  const [selectedTarget, setSelectedTarget] = useState<any | null>(null);
  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [error, setError] = useState('');

  const loadData = () => {
    fetchAPI(`/api/token/${token}/feedback/targets`)
      .then(d => setData({ questions: d.questions || [], targets: d.targets || [] }))
      .catch(err => setError(err.message));
  };

  useEffect(() => {
    loadData();
  }, [token]);

  const openForm = (target: any) => {
    setSelectedTarget(target);
    setAnswers({});
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedTarget) return;
    try {
      await fetchAPI(`/api/token/${token}/feedback`, 'POST', {
        debate_id: selectedTarget.debate_id,
        target_adjudicator_id: selectedTarget.adjudicator_id,
        answers
      });
      setSelectedTarget(null);
      alert('Feedback submitted successfully!');
    } catch (err: any) {
      alert('Failed to submit feedback: ' + err.message);
    }
  };

  if (error || !data) return null;

  return (
    <div className="card" style={{ marginTop: '2rem' }}>
      <h3><MessageSquare size={16} style={{ verticalAlign: '-2px', marginRight: '6px' }} />Give Feedback</h3>
      {data.targets.length === 0 ? (
        <div style={{ padding: '1.5rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>
          No adjudicators are available for feedback yet.
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
          {data.targets.map(t => (
            <div key={`${t.debate_id}-${t.adjudicator_id}`} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.6rem 0.75rem', background: 'rgba(0,0,0,0.02)', borderRadius: '8px' }}>
              <div>
                <div style={{ fontWeight: '500' }}>{t.adjudicator_name}</div>
                <div style={{ fontSize: '0.75rem', color: 'var(--text-mute)' }}>
                  {t.round_name} — {t.venue} | Role: {t.role}
                </div>
              </div>
              <button className="btn btn-primary" style={{ padding: '4px 10px', fontSize: '0.8rem' }} onClick={() => openForm(t)}>
                Give Feedback
              </button>
            </div>
          ))}
        </div>
      )}

      {selectedTarget && (
        <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, background: 'rgba(0,0,0,0.8)', zIndex: 1000, display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
          <div className="card" style={{ maxWidth: '500px', width: '90%', maxHeight: '90vh', overflowY: 'auto' }}>
            <h3>Feedback for {selectedTarget.adjudicator_name}</h3>
            <p style={{ fontSize: '0.8rem', color: 'var(--text-mute)', marginTop: '-0.5rem' }}>
              {selectedTarget.round_name} — {selectedTarget.venue}
            </p>
            <form onSubmit={handleSubmit}>
              {data.questions.map(q => (
                <div key={q.id} className="form-group">
                  <label className="label">
                    {q.name}{q.required ? ' *' : ''}
                    <span style={{ marginLeft: '0.5rem', fontSize: '0.65rem', color: 'var(--text-mute)', textTransform: 'uppercase' }}>{q.type}</span>
                  </label>
                  <QuestionInput q={q} value={answers[q.id] || ''} onChange={v => setAnswers({ ...answers, [q.id]: v })} />
                </div>
              ))}
              <div style={{ display: 'flex', gap: '1rem', marginTop: '1rem' }}>
                <button type="submit" className="btn btn-primary" style={{ flex: 1 }}>Submit Feedback</button>
                <button type="button" className="btn btn-secondary" onClick={() => setSelectedTarget(null)}>Cancel</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
