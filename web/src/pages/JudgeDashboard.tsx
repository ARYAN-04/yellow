import { useState, useEffect } from 'react';
import React from 'react';
import { Link, useParams } from 'react-router-dom';
import { CheckCircle2, Clock, MapPin, UserCheck, FileText, History } from 'lucide-react';
import { fetchAPI } from '../lib/api';
import { getRoleSlotsForSide, type RoleSlot } from '../lib/roles';
import FeedbackSection from '../components/FeedbackSection';

interface SpeakerScoreFormState {
  speakerId: string;
  speakerName?: string;
  score: number;
  isReply: boolean;
  role: string;
  speechOrder: number;
}

export default function JudgeDashboard() {
  const { token } = useParams<{ token: string }>();
  const [tokenInfo, setTokenInfo] = useState<any>(null);
  const [debates, setDebates] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [activeTab, setActiveTab] = useState<'assigned' | 'history'>('assigned');

  // Check-in state
  const [checkinInfo, setCheckinInfo] = useState<{ checked_in: boolean; checked_in_at: string } | null>(null);
  const [checkinLoading, setCheckinLoading] = useState(false);

  // Enter ballot state
  const [selectedDebate, setSelectedDebate] = useState<any | null>(null);
  const [ballotResults, setBallotResults] = useState<Record<string, { points: number; speakerPoints: number }>>({});
  const [teamSpeakerRoles, setTeamSpeakerRoles] = useState<Record<string, SpeakerScoreFormState[]>>({});

  const loadData = () => {
    setLoading(true);
    fetchAPI(`/api/token/${token}`)
      .then(info => {
        setTokenInfo(info);
        return Promise.all([
          fetchAPI(`/api/token/${token}/debates`),
          fetchAPI(`/api/token/${token}/checkin`).catch(() => null)
        ]);
      })
      .then(([d, chk]) => {
        setDebates(d || []);
        if (chk) setCheckinInfo(chk);
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

  const toggleCheckin = async () => {
    if (!checkinInfo) return;
    setCheckinLoading(true);
    try {
      const updated = await fetchAPI(`/api/token/${token}/checkin`, 'POST', {
        checked_in: !checkinInfo.checked_in
      });
      setCheckinInfo(updated);
    } catch (err: any) {
      alert("Failed to update check-in: " + err.message);
    } finally {
      setCheckinLoading(false);
    }
  };

  const openBallotForm = (debate: any) => {
    setSelectedDebate(debate);
    const initialResults: Record<string, { points: number; speakerPoints: number }> = {};
    const initialRoles: Record<string, SpeakerScoreFormState[]> = {};

    debate.teams.forEach((t: any) => {
      const slots: RoleSlot[] = getRoleSlotsForSide(t.side, debate.teams.length);
      const speakers: any[] = t.speakers || [];

      const roleStates: SpeakerScoreFormState[] = slots.map((slot, idx) => {
        const assignedSpeaker = speakers[idx] || speakers[0] || null;
        return {
          speakerId: assignedSpeaker ? assignedSpeaker.id : '',
          speakerName: assignedSpeaker ? assignedSpeaker.name : '',
          score: 75.0,
          isReply: Boolean(slot.isReply),
          role: slot.role,
          speechOrder: slot.order || idx + 1,
        };
      });

      initialRoles[t.team_id] = roleStates;

      // Substantive total
      let sum = 0;
      roleStates.forEach(r => {
        if (!r.isReply) sum += r.score;
      });

      initialResults[t.team_id] = { points: 0, speakerPoints: sum || 150 };
    });

    setBallotResults(initialResults);
    setTeamSpeakerRoles(initialRoles);
  };

  const handleRoleSpeakerChange = (teamId: string, roleIndex: number, speakerId: string) => {
    const currentList = [...(teamSpeakerRoles[teamId] || [])];
    if (!currentList[roleIndex]) return;

    const team = selectedDebate?.teams.find((t: any) => t.team_id === teamId);
    const sp = team?.speakers?.find((s: any) => s.id === speakerId);

    currentList[roleIndex] = {
      ...currentList[roleIndex],
      speakerId: speakerId,
      speakerName: sp ? sp.name : '',
    };

    setTeamSpeakerRoles({ ...teamSpeakerRoles, [teamId]: currentList });
  };

  const handleRoleScoreChange = (teamId: string, roleIndex: number, score: number) => {
    const currentList = [...(teamSpeakerRoles[teamId] || [])];
    if (!currentList[roleIndex]) return;

    currentList[roleIndex] = {
      ...currentList[roleIndex],
      score: score,
    };

    const nextRoles = { ...teamSpeakerRoles, [teamId]: currentList };
    setTeamSpeakerRoles(nextRoles);

    // Live update total substantive speaker points
    let sum = 0;
    currentList.forEach(r => {
      if (!r.isReply) sum += r.score;
    });

    setBallotResults(prev => ({
      ...prev,
      [teamId]: { ...prev[teamId], speakerPoints: sum }
    }));
  };

  const handleBallotSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedDebate) return;

    // Validate that all role slots have a speaker assigned
    for (const t of selectedDebate.teams) {
      const roles = teamSpeakerRoles[t.team_id] || [];
      for (const r of roles) {
        if (!r.speakerId) {
          alert(`Please select a speaker for role ${r.role} in team ${t.team_name}.`);
          return;
        }
      }
    }

    const resultsPayload = Object.keys(ballotResults).map(tid => {
      const roles = teamSpeakerRoles[tid] || [];
      return {
        team_id: tid,
        points: ballotResults[tid].points,
        speaker_points: ballotResults[tid].speakerPoints,
        speaker_scores: roles.map(r => ({
          speaker_id: r.speakerId,
          score: Number(r.score),
          is_reply: r.isReply,
          speech_order: r.speechOrder,
          role: r.role,
        })),
      };
    });

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

  const activeDebates = debates.filter(d => d.ballot_status === 'unsubmitted');
  const pastDebates = debates.filter(d => d.ballot_status !== 'unsubmitted');

  return (
    <div className="container" style={{ maxWidth: '740px' }}>
      {/* Header Profile Card */}
      <div className="card" style={{ marginBottom: '1.5rem', textAlign: 'center' }}>
        <span className="badge badge-warning" style={{ marginBottom: '0.75rem', textTransform: 'uppercase' }}>Adjudicator Portal</span>
        <h2 style={{ margin: '0 0 0.25rem 0' }}>{tokenInfo.owner_name}</h2>
        <p style={{ color: 'var(--text-mute)', fontSize: '0.85rem' }}>Tournament: {tokenInfo.slug} | Scoped Access Link</p>

        {/* Self Check-In Card */}
        {checkinInfo && (
          <div style={{
            marginTop: '1.25rem',
            padding: '0.85rem 1.25rem',
            background: checkinInfo.checked_in ? 'rgba(22, 163, 74, 0.08)' : 'rgba(234, 88, 12, 0.08)',
            border: `1px solid ${checkinInfo.checked_in ? 'rgba(22, 163, 74, 0.25)' : 'rgba(234, 88, 12, 0.25)'}`,
            borderRadius: '8px',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center'
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem', textAlign: 'left' }}>
              {checkinInfo.checked_in ? (
                <CheckCircle2 size={20} color="#16a34a" />
              ) : (
                <Clock size={20} color="#ea580c" />
              )}
              <div>
                <div style={{ fontWeight: 600, fontSize: '0.9rem', color: checkinInfo.checked_in ? '#16a34a' : '#ea580c' }}>
                  {checkinInfo.checked_in ? 'You are Checked In' : 'Not Checked In'}
                </div>
                <div style={{ fontSize: '0.75rem', color: 'var(--text-mute)' }}>
                  {checkinInfo.checked_in ? 'Confirmed available for round allocations' : 'Please check in so organizers know you are present'}
                </div>
              </div>
            </div>

            <button
              onClick={toggleCheckin}
              disabled={checkinLoading}
              className={`btn ${checkinInfo.checked_in ? 'btn-secondary' : 'btn-primary'}`}
              style={{ padding: '6px 14px', fontSize: '0.8rem', display: 'inline-flex', alignItems: 'center', gap: '0.35rem' }}
            >
              <UserCheck size={14} />
              {checkinLoading ? 'Updating...' : checkinInfo.checked_in ? 'Check Out' : 'Check In Now'}
            </button>
          </div>
        )}

        {/* Tab View Buttons */}
        <div style={{ display: 'inline-flex', border: '1px solid var(--border)', borderRadius: '8px', overflow: 'hidden', marginTop: '1.25rem' }}>
          <button
            className="btn btn-secondary"
            onClick={() => setActiveTab('assigned')}
            style={{
              border: 'none', borderRadius: 0, padding: '6px 16px', fontSize: '0.85rem',
              background: activeTab === 'assigned' ? 'var(--accent)' : undefined,
              color: activeTab === 'assigned' ? '#fff' : undefined
            }}
          >
            Assigned Debates ({activeDebates.length})
          </button>
          <button
            className="btn btn-secondary"
            onClick={() => setActiveTab('history')}
            style={{
              border: 'none', borderRadius: 0, padding: '6px 16px', fontSize: '0.85rem',
              background: activeTab === 'history' ? 'var(--accent)' : undefined,
              color: activeTab === 'history' ? '#fff' : undefined
            }}
          >
            Past Debates &amp; History ({pastDebates.length})
          </button>
        </div>
      </div>

      {/* Assigned Debates Tab */}
      {activeTab === 'assigned' && (
        <div className="card">
          <h3>Your Current Debates</h3>
          {activeDebates.length === 0 ? (
            <div style={{ padding: '2rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>
              No pending debate ballots assigned to you right now. Check back when the next round draw is released!
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
              {activeDebates.map(d => (
                <div key={d.id} className="card" style={{ padding: '1.25rem', background: 'rgba(0,0,0,0.01)', border: '1px solid var(--border)' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '1px solid var(--border)', paddingBottom: '0.5rem', marginBottom: '0.75rem' }}>
                    <span style={{ fontWeight: '600', color: 'var(--text-h)', display: 'inline-flex', alignItems: 'center', gap: '0.35rem' }}>
                      <MapPin size={15} /> {d.round_name} — {d.venue}
                    </span>
                    <span className="badge badge-info" style={{ textTransform: 'uppercase' }}>Role: {d.role}</span>
                  </div>

                  {d.motion && (
                    <div style={{ marginBottom: '0.75rem', padding: '0.6rem 0.8rem', background: 'rgba(0,0,0,0.02)', borderRadius: '6px', fontSize: '0.85rem' }}>
                      <strong style={{ color: 'var(--text-mute)', fontSize: '0.75rem', textTransform: 'uppercase' }}>Motion: </strong>
                      <span>{d.motion}</span>
                    </div>
                  )}

                  <div className="grid grid-cols-2" style={{ gap: '0.5rem', marginBottom: '1rem' }}>
                    {d.teams.map((t: any) => (
                      <div key={t.team_id} style={{ padding: '0.5rem', background: 'rgba(0,0,0,0.02)', borderRadius: '6px', fontSize: '0.85rem' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.2rem' }}>
                          <span style={{ fontWeight: 600 }}>{t.team_name}</span>
                          <span className="badge badge-info" style={{ textTransform: 'uppercase', fontSize: '0.7rem' }}>{t.side}</span>
                        </div>
                        {t.speakers?.length > 0 && (
                          <div style={{ fontSize: '0.75rem', color: 'var(--text-mute)' }}>
                            Roster: {t.speakers.map((s: any) => s.name).join(', ')}
                          </div>
                        )}
                      </div>
                    ))}
                  </div>

                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderTop: '1px solid var(--border)', paddingTop: '0.75rem' }}>
                    <span style={{ fontSize: '0.8rem', color: 'var(--text-mute)' }}>
                      Ballot status: <span style={{ fontWeight: '600', color: 'var(--text-mute)' }}>Unsubmitted</span>
                    </span>
                    {d.role === 'chair' && (
                      <button className="btn btn-primary" style={{ padding: '6px 14px', fontSize: '0.8rem', display: 'inline-flex', alignItems: 'center', gap: '0.35rem' }} onClick={() => openBallotForm(d)}>
                        <FileText size={14} /> Enter Official Ballot
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Past Debates / History Tab */}
      {activeTab === 'history' && (
        <div className="card">
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.5rem' }}>
            <History size={18} />
            <h3 style={{ margin: 0 }}>Debate Adjudication History</h3>
          </div>
          <p style={{ fontSize: '0.85rem', color: 'var(--text-mute)', marginBottom: '1.25rem' }}>
            Overview of all past rounds you have judged in this tournament.
          </p>

          {pastDebates.length === 0 ? (
            <div style={{ padding: '2rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>
              No completed debate records yet.
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              {pastDebates.map(d => (
                <div key={d.id} style={{ padding: '1rem', background: 'rgba(0,0,0,0.01)', border: '1px solid var(--border)', borderRadius: '8px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
                    <span style={{ fontWeight: 700, fontSize: '0.95rem', color: 'var(--text-h)' }}>
                      {d.round_name} • {d.venue}
                    </span>
                    <div style={{ display: 'flex', gap: '0.4rem', alignItems: 'center' }}>
                      <span className="badge badge-info" style={{ textTransform: 'uppercase', fontSize: '0.7rem' }}>
                        Role: {d.role}
                      </span>
                      <span
                        className="badge"
                        style={{
                          background: d.ballot_status === 'confirmed' ? 'rgba(22,163,74,0.1)' : 'rgba(37,99,235,0.1)',
                          color: d.ballot_status === 'confirmed' ? 'var(--success)' : '#1d4ed8',
                          fontSize: '0.7rem',
                          textTransform: 'capitalize'
                        }}
                      >
                        {d.ballot_status}
                      </span>
                    </div>
                  </div>

                  {d.motion && (
                    <div style={{ fontSize: '0.8rem', color: 'var(--text-mute)', marginBottom: '0.5rem' }}>
                      <strong>Motion: </strong>{d.motion}
                    </div>
                  )}

                  <div style={{ fontSize: '0.8rem', color: 'var(--text-mute)' }}>
                    Teams: {d.teams.map((t: any) => `${t.team_name} (${t.side})`).join(' • ')}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Adjudicator Feedback Form */}
      <FeedbackSection token={token!} />

      {/* --- ENTER BALLOT MODAL --- */}
      {selectedDebate && (
        <div style={{
          position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
          background: 'rgba(0,0,0,0.8)', zIndex: 1000,
          display: 'flex', justifyContent: 'center', alignItems: 'center',
          padding: '1rem'
        }}>
          <div className="card" style={{ maxWidth: '640px', width: '100%', maxHeight: '92vh', overflowY: 'auto' }}>
            <div style={{ borderBottom: '1px solid var(--border)', paddingBottom: '0.75rem', marginBottom: '1.25rem' }}>
              <h3 style={{ margin: '0 0 0.25rem 0' }}>Submit Official Ballot</h3>
              <div style={{ fontSize: '0.85rem', color: 'var(--text-mute)' }}>
                {selectedDebate.round_name} • Venue: <strong>{selectedDebate.venue}</strong>
              </div>
            </div>

            <form onSubmit={handleBallotSubmit}>
              {selectedDebate.teams.map((t: any) => {
                const roles = teamSpeakerRoles[t.team_id] || [];
                const speakers = t.speakers || [];
                const isTwoTeam = selectedDebate.teams.length === 2;

                return (
                  <div key={t.team_id} style={{ padding: '1rem', border: '1px solid var(--border)', borderRadius: '8px', marginBottom: '1.25rem', background: 'rgba(0,0,0,0.01)' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
                      <span style={{ fontWeight: 700, fontSize: '1rem', color: 'var(--text-h)' }}>{t.team_name}</span>
                      <span className="badge badge-info" style={{ textTransform: 'uppercase' }}>{t.side}</span>
                    </div>

                    {/* Rank & Substantive points */}
                    <div className="grid grid-cols-2" style={{ gap: '0.75rem', marginBottom: '1rem' }}>
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
                          {isTwoTeam ? (
                            <>
                              <option value="1">1 Point (Win)</option>
                              <option value="0">0 Points (Loss)</option>
                            </>
                          ) : (
                            <>
                              <option value="3">3 Points (1st Place)</option>
                              <option value="2">2 Points (2nd Place)</option>
                              <option value="1">1 Point (3rd Place)</option>
                              <option value="0">0 Points (4th Place)</option>
                            </>
                          )}
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

                    {/* Role Slots Scoring */}
                    <div style={{ background: 'rgba(0,0,0,0.02)', padding: '0.75rem', borderRadius: '6px', border: '1px solid var(--border)' }}>
                      <div style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-mute)', marginBottom: '0.6rem', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                        Speaker-Wise Scores &amp; Role Positions
                      </div>

                      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.6rem' }}>
                        {roles.map((r, rIdx) => (
                          <div key={rIdx} style={{ display: 'grid', gridTemplateColumns: '1.2fr 2fr 1fr', gap: '0.5rem', alignItems: 'center' }}>
                            <div style={{ display: 'inline-flex', alignItems: 'center', gap: '0.3rem' }}>
                              <span style={{ fontWeight: 600, fontSize: '0.85rem', color: 'var(--text-h)' }}>{r.role}</span>
                              {r.isReply && <span className="badge badge-warning" style={{ fontSize: '0.65rem' }}>Reply</span>}
                            </div>

                            <select
                              className="input select"
                              style={{ fontSize: '0.8rem', padding: '4px 8px' }}
                              value={r.speakerId}
                              onChange={e => handleRoleSpeakerChange(t.team_id, rIdx, e.target.value)}
                              required
                            >
                              <option value="">-- Select Speaker --</option>
                              {speakers.map((sp: any) => (
                                <option key={sp.id} value={sp.id}>
                                  {sp.name} {sp.is_novice ? '(Novice)' : ''}
                                </option>
                              ))}
                            </select>

                            <input
                              type="number"
                              step="0.5"
                              min="0"
                              max="100"
                              className="input"
                              style={{ fontSize: '0.85rem', padding: '4px 8px', textAlign: 'right' }}
                              value={r.score}
                              onChange={e => handleRoleScoreChange(t.team_id, rIdx, Number(e.target.value))}
                              required
                            />
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>
                );
              })}

              <div style={{ display: 'flex', gap: '1rem', marginTop: '1.25rem' }}>
                <button type="submit" className="btn btn-primary" style={{ flex: 1 }}>Submit Official Ballot</button>
                <button type="button" className="btn btn-secondary" onClick={() => setSelectedDebate(null)}>Cancel</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
