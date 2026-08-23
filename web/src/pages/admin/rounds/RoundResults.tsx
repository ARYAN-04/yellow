import { useState } from 'react';
import { useOutletContext } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { MapPin, RefreshCw } from 'lucide-react';
import { fetchAPI, type RoundContext } from '../../../lib/api';

interface ScoreEntry {
  points: number,
  speakerPoints: number,
}

const statusChipStyle: Record<string, React.CSSProperties> = {
  none: { background: 'rgba(0,0,0,0.05)', color: 'var(--text-mute)', border: '1px solid var(--border)' },
  draft: { background: 'rgba(113,113,122,0.12)', color: '#52525b', border: '1px solid rgba(82,82,91,0.3)' },
  submitted: { background: 'rgba(37,99,235,0.1)', color: '#1d4ed8', border: '1px solid rgba(37,99,235,0.25)' },
  discrepancy: { background: 'rgba(220,38,38,0.1)', color: '#b91c1c', border: '1px solid rgba(220,38,38,0.3)' },
  confirmed: { background: 'rgba(22,163,74,0.08)', color: 'var(--success)', border: '1px solid rgba(22,163,74,0.15)' },
};

function BallotStatusChip({ status }: { status: string }) {
  const key = statusChipStyle[status] ? status : 'none';
  const label = status === 'none' ? 'None' : status.charAt(0).toUpperCase() + status.slice(1);
  return <span className="badge" style={statusChipStyle[key]}>{label}</span>;
}

export default function RoundResults() {
  const { slug, roundId, isReadOnly } = useOutletContext<RoundContext>();
  const queryClient = useQueryClient();

  const [view, setView] = useState<'enter' | 'registry'>('enter');
  const [debateBallot, setDebateBallot] = useState<any | null>(null);
  const [ballotResults, setBallotResults] = useState<Record<string, ScoreEntry>>({});
  const [isSplit, setIsSplit] = useState(false);
  const [splitScores, setSplitScores] = useState<Record<string, Record<string, ScoreEntry>>>({});
  const [isDoubleEntry, setIsDoubleEntry] = useState(false);
  const [entryGroup, setEntryGroup] = useState('');
  const [discrepancyDiffs, setDiscrepancyDiffs] = useState<any[] | null>(null);

  const { data: draw = [], isLoading } = useQuery<any[]>({
    queryKey: ['draw', slug, roundId],
    queryFn: () => fetchAPI(`/api/t/${slug}/rounds/${roundId}/draw`),
  });

  const { data: registry = [], refetch: refetchRegistry } = useQuery<any[]>({
    queryKey: ['round-ballots', slug, roundId],
    queryFn: () => fetchAPI(`/api/t/${slug}/rounds/${roundId}/ballots`),
    enabled: view === 'registry',
  });

  const openBallotForm = (debate: any) => {
    setDebateBallot(debate);
    setIsSplit(false);
    setIsDoubleEntry(false);
    setDiscrepancyDiffs(null);
    setEntryGroup(`eg-${Date.now().toString(36)}`);
    const initialResults: Record<string, ScoreEntry> = {};
    debate.teams.forEach((t: any) => {
      initialResults[t.team_id] = { points: 0, speakerPoints: 150 };
    });
    setBallotResults(initialResults);
    const initialSplit: Record<string, Record<string, ScoreEntry>> = {};
    debate.adjudicators
      .filter((a: any) => a.role !== 'trainee')
      .forEach((a: any) => {
        initialSplit[a.adjudicator_id] = {};
        debate.teams.forEach((t: any) => {
          initialSplit[a.adjudicator_id][t.team_id] = { points: 0, speakerPoints: 150 };
        });
      });
    setSplitScores(initialSplit);
  };

  const submitBallot = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!debateBallot) return;

    let resultsPayload: any[];
    if (isSplit) {
      resultsPayload = [];
      Object.keys(splitScores).forEach(adjId => {
        debateBallot.teams.forEach((t: any) => {
          const entry = splitScores[adjId][t.team_id];
          resultsPayload.push({
            team_id: t.team_id,
            adjudicator_id: adjId,
            points: entry.points,
            speaker_points: entry.speakerPoints
          });
        });
      });
    } else {
      resultsPayload = Object.keys(ballotResults).map(tid => ({
        team_id: tid,
        points: ballotResults[tid].points,
        speaker_points: ballotResults[tid].speakerPoints
      }));
    }

    try {
      const res = await fetchAPI(`/api/t/${slug}/debates/${debateBallot.id}/ballots`, 'POST', {
        submitter_type: 'organizer',
        submitter_id: 'admin',
        is_split: isSplit,
        entry_group: isDoubleEntry ? entryGroup : '',
        results: resultsPayload
      });

      await fetchAPI(`/api/t/${slug}/ballots/${res.id}/confirm`, 'POST');

      setDebateBallot(null);
      queryClient.invalidateQueries({ queryKey: ['draw', slug, roundId] });
      queryClient.invalidateQueries({ queryKey: ['standings'] });
      queryClient.invalidateQueries({ queryKey: ['round-ballots', slug, roundId] });
      alert("Ballot entered and confirmed!");
    } catch (err: any) {
      if (err.data?.error === 'discrepancy') {
        setDiscrepancyDiffs(err.data.diffs || []);
      } else {
        alert("Failed to submit ballot: " + err.message);
      }
      queryClient.invalidateQueries({ queryKey: ['round-ballots', slug, roundId] });
    }
  };

  return (
    <div>
      <div className="card" style={{ marginBottom: '1.5rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h3 style={{ margin: 0 }}>Results &amp; Ballots</h3>
          <p style={{ fontSize: '0.85rem', color: 'var(--text-mute)', margin: '0.5rem 0 0 0' }}>
            Record ballots per debate or review the ballot registry.
          </p>
        </div>
        <div style={{ display: 'inline-flex', border: '1px solid var(--border)', borderRadius: '8px', overflow: 'hidden' }}>
          {[['enter', 'Enter Results'], ['registry', 'Ballot Registry']].map(([key, label]) => (
            <button
              key={key}
              className="btn btn-secondary"
              onClick={() => setView(key as 'enter' | 'registry')}
              style={{
                border: 'none', borderRadius: 0, padding: '6px 14px', fontSize: '0.8rem',
                background: view === key ? 'var(--accent)' : undefined,
                color: view === key ? '#fff' : undefined
              }}
            >
              {label}
            </button>
          ))}
        </div>
      </div>

      {view === 'registry' && (
        <div className="card">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
            <h4 style={{ margin: 0 }}>Ballot Registry</h4>
            <button className="btn btn-secondary" style={{ padding: '4px 10px', fontSize: '0.75rem' }} onClick={() => refetchRegistry()}>
              <RefreshCw size={13} style={{ marginRight: '4px' }} /> Refresh
            </button>
          </div>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.85rem' }}>
            <thead>
              <tr style={{ textAlign: 'left', borderBottom: '1px solid var(--border)', color: 'var(--text-mute)' }}>
                <th style={{ padding: '0.5rem' }}>Venue</th>
                <th style={{ padding: '0.5rem' }}>Status</th>
                <th style={{ padding: '0.5rem' }}>Submitter</th>
                <th style={{ padding: '0.5rem' }}>Type</th>
                <th style={{ padding: '0.5rem' }}>Group</th>
              </tr>
            </thead>
            <tbody>
              {registry.length === 0 ? (
                <tr><td colSpan={5} style={{ padding: '1.5rem', textAlign: 'center', color: 'var(--text-mute)' }}>No ballots recorded for this round yet.</td></tr>
              ) : (
                registry.map(b => (
                  <tr key={b.id} style={{ borderBottom: '1px solid var(--border)' }}>
                    <td style={{ padding: '0.5rem' }}>{b.debate_venue}</td>
                    <td style={{ padding: '0.5rem' }}><BallotStatusChip status={b.status} /></td>
                    <td style={{ padding: '0.5rem' }}>{b.submitter_name || b.submitter_type}</td>
                    <td style={{ padding: '0.5rem' }}>{b.is_split ? <span className="badge badge-warning">Split</span> : <span className="badge badge-info">Consensus</span>}</td>
                    <td style={{ padding: '0.5rem', fontFamily: 'monospace', fontSize: '0.75rem' }}>{b.entry_group || '-'}</td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}

      {view === 'enter' && (isLoading ? (
        <div style={{ padding: '3rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>Loading debates...</div>
      ) : draw.length === 0 ? (
        <div style={{ padding: '3rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>
          No matches generated for this round yet. Generate the draw first.
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
          {draw.map(d => (
            <div key={d.id} className="card" style={{ padding: '1.25rem', background: 'rgba(0,0,0,0.01)', border: '1px solid var(--border)' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '1px solid var(--border)', paddingBottom: '0.5rem', marginBottom: '0.75rem' }}>
                <span style={{ fontWeight: '600', color: 'var(--text-h)', display: 'inline-flex', alignItems: 'center', gap: '0.4rem' }}>
                  <MapPin size={16} /> {d.venue}
                </span>
                {!isReadOnly && (
                  <button className="btn btn-secondary" style={{ padding: '4px 10px', fontSize: '0.75rem' }} onClick={() => openBallotForm(d)}>
                    Record Ballot
                  </button>
                )}
              </div>

              {/* BP Grid representation */}
              <div className="grid grid-cols-2" style={{ gap: '0.75rem' }}>
                {d.teams.map((t: any) => (
                  <div key={t.team_id} style={{ display: 'flex', justifyContent: 'space-between', padding: '0.5rem', background: 'rgba(0,0,0,0.02)', borderRadius: '6px' }}>
                    <span style={{ fontSize: '0.85rem' }}>{t.team_name}</span>
                    <span className="badge badge-info" style={{ textTransform: 'uppercase' }}>{t.side}</span>
                  </div>
                ))}
              </div>

              <div style={{ marginTop: '0.75rem', fontSize: '0.8rem', color: 'var(--text-mute)' }}>
                Chair Panel: {d.adjudicators.filter((a: any) => a.role === 'chair').map((a: any) => a.adjudicator_name).join(', ') || 'No chair assigned'}
              </div>
            </div>
          ))}
        </div>
      ))}

      {/* --- BALLOT ENTRY MODAL --- */}
      {debateBallot && !isReadOnly && (
        <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, background: 'rgba(0,0,0,0.8)', zIndex: 1000, display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
          <div className="card" style={{ maxWidth: '560px', width: '90%', maxHeight: '90vh', overflowY: 'auto' }}>
            <h3>Enter Ballot: {debateBallot.venue}</h3>
            <form onSubmit={submitBallot}>
              <div style={{ display: 'flex', gap: '1.5rem', marginBottom: '1rem' }}>
                <label style={{ display: 'inline-flex', alignItems: 'center', gap: '0.4rem', fontSize: '0.85rem' }}>
                  <input type="checkbox" checked={isSplit} onChange={e => setIsSplit(e.target.checked)} />
                  Split panel
                </label>
                <label style={{ display: 'inline-flex', alignItems: 'center', gap: '0.4rem', fontSize: '0.85rem' }}>
                  <input type="checkbox" checked={isDoubleEntry} onChange={e => setIsDoubleEntry(e.target.checked)} />
                  Double-entry
                </label>
              </div>

              {isDoubleEntry && (
                <div style={{ marginBottom: '1rem' }}>
                  <label className="label">Entry Group ID</label>
                  <input
                    className="input"
                    value={entryGroup}
                    onChange={e => setEntryGroup(e.target.value)}
                    placeholder="Shared id linking paired drafts"
                    required
                  />
                </div>
              )}

              {!isSplit && debateBallot.teams.map((t: any) => (
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

              {isSplit && debateBallot.adjudicators.filter((a: any) => a.role !== 'trainee').map((a: any) => (
                <div key={a.adjudicator_id} style={{ padding: '1rem', borderBottom: '1px solid var(--border)', marginBottom: '1rem' }}>
                  <div style={{ fontWeight: '600', color: 'var(--text-h)', marginBottom: '0.5rem' }}>
                    {a.adjudicator_name} <span className="badge badge-info" style={{ marginLeft: '0.4rem', textTransform: 'uppercase' }}>{a.role}</span>
                  </div>
                  {debateBallot.teams.map((t: any) => (
                    <div key={t.team_id} className="grid grid-cols-2" style={{ gap: '0.75rem', marginTop: '0.5rem' }}>
                      <div>
                        <label className="label">{t.team_name} — Points</label>
                        <select
                          className="input select"
                          value={splitScores[a.adjudicator_id]?.[t.team_id]?.points || 0}
                          onChange={e => setSplitScores({
                            ...splitScores,
                            [a.adjudicator_id]: {
                              ...splitScores[a.adjudicator_id],
                              [t.team_id]: { ...splitScores[a.adjudicator_id]?.[t.team_id], points: Number(e.target.value) }
                            }
                          })}
                        >
                          <option value="3">3 Points (1st Place)</option>
                          <option value="2">2 Points (2nd Place)</option>
                          <option value="1">1 Point (3rd Place)</option>
                          <option value="0">0 Points (4th Place)</option>
                        </select>
                      </div>
                      <div>
                        <label className="label">{t.team_name} — Speaker</label>
                        <input
                          type="number"
                          step="0.5"
                          className="input"
                          value={splitScores[a.adjudicator_id]?.[t.team_id]?.speakerPoints || 150}
                          onChange={e => setSplitScores({
                            ...splitScores,
                            [a.adjudicator_id]: {
                              ...splitScores[a.adjudicator_id],
                              [t.team_id]: { ...splitScores[a.adjudicator_id]?.[t.team_id], speakerPoints: Number(e.target.value) }
                            }
                          })}
                        />
                      </div>
                    </div>
                  ))}
                </div>
              ))}

              {discrepancyDiffs && (
                <div style={{ background: 'rgba(220,38,38,0.06)', border: '1px solid rgba(220,38,38,0.35)', borderRadius: '8px', padding: '0.75rem', margin: '0.75rem 0' }}>
                  <strong style={{ color: '#b91c1c', fontSize: '0.85rem' }}>Discrepancy detected — drafts do not match:</strong>
                  <ul style={{ margin: '0.5rem 0 0 0', paddingLeft: '1.2rem', fontSize: '0.8rem', color: '#b91c1c' }}>
                    {discrepancyDiffs.map((d: any, i: number) => (
                      <li key={i}>
                        Team {d.team_id}{d.adjudicator_id ? ` / judge ${d.adjudicator_id}` : ''}: {d.field.replace('_', ' ')} is{' '}
                        <strong>{d.ballot_a === null ? 'missing' : d.ballot_a}</strong> vs <strong>{d.ballot_b === null ? 'missing' : d.ballot_b}</strong>.
                      </li>
                    ))}
                  </ul>
                  <div style={{ fontSize: '0.75rem', color: 'var(--text-mute)', marginTop: '0.4rem' }}>
                    Correct the scores above and resubmit with the same entry group to replace this draft.
                  </div>
                </div>
              )}

              <div style={{ display: 'flex', gap: '1rem', marginTop: '1rem' }}>
                <button type="submit" className="btn btn-primary" style={{ flex: 1 }}>Submit &amp; Confirm</button>
                <button type="button" className="btn btn-secondary" onClick={() => { setDebateBallot(null); setDiscrepancyDiffs(null); }}>Cancel</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
