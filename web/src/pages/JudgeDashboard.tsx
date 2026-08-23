import { useState, useEffect } from 'react';
import React from 'react';
import { Link, useParams } from 'react-router-dom';
import { fetchAPI } from '../lib/api';
import FeedbackSection from '../components/FeedbackSection';

export default function JudgeDashboard() {
  const { token } = useParams<{ token: string }>();
  const [tokenInfo, setTokenInfo] = useState<any>(null);
  const [debates, setDebates] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [selectedDebate, setSelectedDebate] = useState<any | null>(null);
  const [ballotResults, setBallotResults] = useState<Record<string, { points: number, speakerPoints: number }>>({});

  const loadData = () => {
    setLoading(true);
    fetchAPI(`/api/token/${token}`)
      .then(info => {
        setTokenInfo(info);
        return fetchAPI(`/api/token/${token}/debates`);
      })
      .then(d => {
        setDebates(d || []);
        setLoading(false);
      })
      .catch(err => {
        setError(err.message);
        setLoading(false);
      });
  };

  useEffect(() => {
    loadData();
  }, [token]);

  const openBallotForm = (debate: any) => {
    setSelectedDebate(debate);
    const initialResults: Record<string, { points: number, speakerPoints: number }> = {};
    debate.teams.forEach((t: any) => {
      initialResults[t.team_id] = { points: 0, speakerPoints: 150 };
    });
    setBallotResults(initialResults);
  };

  const handleBallotSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedDebate) return;

    const resultsPayload = Object.keys(ballotResults).map(tid => ({
      team_id: tid,
      points: ballotResults[tid].points,
      speaker_points: ballotResults[tid].speakerPoints
    }));

    try {
      await fetchAPI(`/api/token/${token}/debates/${selectedDebate.id}/ballots`, 'POST', {
        results: resultsPayload
      });
      setSelectedDebate(null);
      loadData();
      alert("Ballot submitted successfully! It is pending organizer confirmation.");
    } catch (err: any) {
      alert("Failed to submit ballot: " + err.message);
    }
  };

  if (loading) {
    return <div className="container" style={{ textAlign: 'center', marginTop: '6rem' }}>Loading Portal...</div>;
  }

  if (error) {
    return (
      <div className="container" style={{ maxWidth: '420px', marginTop: '6rem' }}>
        <div className="card" style={{ textAlign: 'center', borderColor: 'var(--danger)' }}>
          <h2 style={{ color: 'var(--danger)' }}>Access Error</h2>
          <p>{error}</p>
          <div style={{ marginTop: '1.5rem' }}>
            <Link to="/" className="btn btn-secondary">Go Home</Link>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="container" style={{ maxWidth: '720px' }}>
      <div className="card" style={{ marginBottom: '2rem', textAlign: 'center' }}>
        <span className="badge badge-warning" style={{ marginBottom: '0.75rem', textTransform: 'uppercase' }}>Adjudicator Portal</span>
        <h2 style={{ margin: '0 0 0.25rem 0' }}>{tokenInfo.owner_name}</h2>
        <p style={{ color: 'var(--text-mute)', fontSize: '0.85rem' }}>Tournament: {tokenInfo.slug} | Scoped Access Link</p>
      </div>

      <div className="card">
        <h3>Your Assigned Debates</h3>
        {debates.length === 0 ? (
          <div style={{ padding: '2rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>
            No debates assigned to you in active rounds.
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
            {debates.map(d => (
              <div key={d.id} className="card" style={{ padding: '1.25rem', background: 'rgba(0,0,0,0.01)', border: '1px solid var(--border)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '1px solid var(--border)', paddingBottom: '0.5rem', marginBottom: '0.75rem' }}>
                  <span style={{ fontWeight: '600', color: 'var(--text-h)' }}>{d.round_name} — {d.venue}</span>
                  <span className="badge badge-info" style={{ textTransform: 'uppercase' }}>Role: {d.role}</span>
                </div>

                <div className="grid grid-cols-2" style={{ gap: '0.5rem', marginBottom: '1rem' }}>
                  {d.teams.map((t: any) => (
                    <div key={t.team_id} style={{ display: 'flex', justifyContent: 'space-between', padding: '0.4rem', background: 'rgba(0,0,0,0.02)', borderRadius: '6px', fontSize: '0.85rem' }}>
                      <span>{t.team_name}</span>
                      <span style={{ textTransform: 'uppercase', color: 'var(--text-mute)' }}>{t.side}</span>
                    </div>
                  ))}
                </div>

                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderTop: '1px solid var(--border)', paddingTop: '0.75rem' }}>
                  <span style={{ fontSize: '0.8rem', color: 'var(--text-mute)' }}>
                    Ballot status: <span style={{ fontWeight: '600', color: d.ballot_status === 'confirmed' ? 'var(--success)' : d.ballot_status === 'submitted' ? 'var(--warning)' : 'var(--text-mute)' }}>{d.ballot_status}</span>
                  </span>
                  {d.ballot_status === 'unsubmitted' && d.role === 'chair' && (
                    <button className="btn btn-primary" style={{ padding: '4px 10px', fontSize: '0.8rem' }} onClick={() => openBallotForm(d)}>
                      Enter Ballot
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <FeedbackSection token={token!} />

      {selectedDebate && (
        <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, background: 'rgba(0,0,0,0.8)', zIndex: 1000, display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
          <div className="card" style={{ maxWidth: '500px', width: '90%', maxHeight: '90vh', overflowY: 'auto' }}>
            <h3>Submit Ballot: {selectedDebate.round_name} ({selectedDebate.venue})</h3>
            <form onSubmit={handleBallotSubmit}>
              {selectedDebate.teams.map((t: any) => (
                <div key={t.team_id} style={{ padding: '1rem', borderBottom: '1px solid var(--border)', marginBottom: '1rem' }}>
                  <div style={{ fontWeight: '600', color: 'var(--text-h)', marginBottom: '0.5rem' }}>
                    {t.team_name} ({t.side.toUpperCase()})
                  </div>
                  <div className="grid grid-cols-2" style={{ gap: '0.75rem' }}>
                    <div>
                      <label className="label">Wins / Rank Points</label>
                      <select
                        className="input select"
                        value={ballotResults[t.team_id]?.points || 0}
                        onChange={e => setBallotResults({
                          ...ballotResults,
                          [t.team_id]: { ...ballotResults[t.team_id], points: Number(e.target.value) }
                        })}
                      >
                        <option value="3">3 Points (1st Place)</option>
                        <option value="2">2 Points (2nd Place)</option>
                        <option value="1">1 Point (3rd Place)</option>
                        <option value="0">0 Points (4th Place)</option>
                      </select>
                    </div>
                    <div>
                      <label className="label">Total Speaker Points</label>
                      <input
                        type="number"
                        step="0.5"
                        className="input"
                        value={ballotResults[t.team_id]?.speakerPoints || 150}
                        onChange={e => setBallotResults({
                          ...ballotResults,
                          [t.team_id]: { ...ballotResults[t.team_id], speakerPoints: Number(e.target.value) }
                        })}
                      />
                    </div>
                  </div>
                </div>
              ))}

              <div style={{ display: 'flex', gap: '1rem', marginTop: '1rem' }}>
                <button type="submit" className="btn btn-primary" style={{ flex: 1 }}>Submit Ballot</button>
                <button type="button" className="btn btn-secondary" onClick={() => setSelectedDebate(null)}>Cancel</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
