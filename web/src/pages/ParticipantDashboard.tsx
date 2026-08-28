import { useState, useEffect } from 'react';
import { Link, useParams } from 'react-router-dom';
import { TrendingUp, CheckCircle2, Clock, UserCheck, Users, User, MapPin } from 'lucide-react';
import { fetchAPI } from '../lib/api';
import FeedbackSection from '../components/FeedbackSection';
import TrajectoryModal from '../components/TrajectoryModal';

export default function ParticipantDashboard() {
  const { token } = useParams<{ token: string }>();
  const [tokenInfo, setTokenInfo] = useState<any>(null);
  const [activeTab, setActiveTab] = useState<'schedule' | 'speakers'>('schedule');
  const [debates, setDebates] = useState<any[]>([]);
  const [standings, setStandings] = useState<any[]>([]);
  const [speakerStandings, setSpeakerStandings] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [trajectoryModal, setTrajectoryModal] = useState<{ type: 'team' | 'speaker'; id: string } | null>(null);

  // Check-in state
  const [checkinInfo, setCheckinInfo] = useState<{ checked_in: boolean; checked_in_at: string } | null>(null);
  const [checkinLoading, setCheckinLoading] = useState(false);

  const loadData = () => {
    setLoading(true);
    fetchAPI(`/api/token/${token}`)
      .then(info => {
        setTokenInfo(info);
        const loadDebates = fetchAPI(`/api/token/${token}/debates`);
        const loadStandings = fetchAPI(`/api/t/${info.slug}/standings`).catch(() => []);
        const loadSpeakerStandings = fetchAPI(`/api/t/${info.slug}/standings/speakers`).catch(() => []);
        const loadCheckin = fetchAPI(`/api/token/${token}/checkin`).catch(() => null);
        return Promise.all([loadDebates, loadStandings, loadSpeakerStandings, loadCheckin, info]);
      })
      .then(([d, s, sp, chk]) => {
        setDebates(d || []);
        setStandings(s || []);
        setSpeakerStandings(sp || []);
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

  const myStanding = standings.find((team: any) => team.team_id === tokenInfo?.owner_id);
  const myRank = standings.findIndex((team: any) => team.team_id === tokenInfo?.owner_id) + 1;
  const teamSpeakers = tokenInfo?.speakers || [];

  return (
    <div className="container" style={{ maxWidth: '740px' }}>
      {/* Team Profile & Header */}
      <div className="card" style={{ marginBottom: '1.5rem', textAlign: 'center' }}>
        <span className="badge badge-info" style={{ marginBottom: '0.75rem', textTransform: 'uppercase' }}>Debater Portal</span>
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
                  {checkinInfo.checked_in ? 'Your Team is Checked In' : 'Not Checked In'}
                </div>
                <div style={{ fontSize: '0.75rem', color: 'var(--text-mute)' }}>
                  {checkinInfo.checked_in ? 'Confirmed available for draw pairing' : 'Please check in so tab organizers pair your team'}
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

        {/* Tab Buttons */}
        <div style={{ display: 'inline-flex', border: '1px solid var(--border)', borderRadius: '8px', overflow: 'hidden', marginTop: '1.25rem' }}>
          <button
            className="btn btn-secondary"
            onClick={() => setActiveTab('schedule')}
            style={{
              border: 'none', borderRadius: 0, padding: '6px 16px', fontSize: '0.85rem',
              background: activeTab === 'schedule' ? 'var(--accent)' : undefined,
              color: activeTab === 'schedule' ? '#fff' : undefined
            }}
          >
            My Team &amp; Schedule
          </button>
          <button
            className="btn btn-secondary"
            onClick={() => setActiveTab('speakers')}
            style={{
              border: 'none', borderRadius: 0, padding: '6px 16px', fontSize: '0.85rem',
              background: activeTab === 'speakers' ? 'var(--accent)' : undefined,
              color: activeTab === 'speakers' ? '#fff' : undefined
            }}
          >
            Speaker Tab
          </button>
        </div>
      </div>

      {activeTab === 'schedule' && (
        <>
          {/* Team Speakers Roster Card */}
          {teamSpeakers.length > 0 && (
            <div className="card" style={{ marginBottom: '1.5rem' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <Users size={18} />
                  <h3 style={{ margin: 0, fontSize: '1rem' }}>Team Speaker Roster</h3>
                </div>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '0.75rem' }}>
                {teamSpeakers.map((sp: any) => (
                  <div
                    key={sp.id}
                    style={{
                      padding: '0.75rem',
                      background: 'rgba(0,0,0,0.02)',
                      border: '1px solid var(--border)',
                      borderRadius: '6px',
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center'
                    }}
                  >
                    <div>
                      <div style={{ fontWeight: 600, fontSize: '0.9rem', color: 'var(--text-h)' }}>
                        {sp.name}
                      </div>
                      <div style={{ display: 'flex', gap: '0.3rem', marginTop: '0.25rem', flexWrap: 'wrap' }}>
                        {sp.is_novice && <span className="badge badge-warning" style={{ fontSize: '0.65rem' }}>Novice</span>}
                        {sp.is_esl && <span className="badge badge-info" style={{ fontSize: '0.65rem' }}>ESL</span>}
                        {sp.is_efl && <span className="badge badge-info" style={{ fontSize: '0.65rem' }}>EFL</span>}
                        {!sp.is_novice && !sp.is_esl && !sp.is_efl && <span className="badge badge-secondary" style={{ fontSize: '0.65rem' }}>Open</span>}
                      </div>
                    </div>

                    <button
                      type="button"
                      className="btn btn-secondary"
                      style={{ padding: '4px 8px', fontSize: '0.75rem', display: 'inline-flex', alignItems: 'center', gap: '0.25rem' }}
                      onClick={() => setTrajectoryModal({ type: 'speaker', id: sp.id })}
                      title="View individual trajectory"
                    >
                      <User size={13} /> Trajectory
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Cumulative Standing Card */}
          {myStanding && (
            <div className="card" style={{ marginBottom: '1.5rem', textAlign: 'center' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <h3 style={{ margin: 0 }}>Cumulative Standing</h3>
                <button
                  type="button"
                  className="btn btn-secondary"
                  style={{ fontSize: '0.8rem', padding: '0.35rem 0.75rem', display: 'inline-flex', alignItems: 'center', gap: '0.3rem' }}
                  onClick={() => setTrajectoryModal({ type: 'team', id: tokenInfo.owner_id })}
                >
                  <TrendingUp size={14} /> Full Team Trajectory
                </button>
              </div>
              <div className="grid grid-cols-3" style={{ marginTop: '1rem' }}>
                <div style={{ borderRight: '1px solid var(--border)' }}>
                  <div style={{ fontSize: '0.75rem', textTransform: 'uppercase', color: 'var(--text-mute)', marginBottom: '0.25rem' }}>Rank</div>
                  <div style={{ fontSize: '1.5rem', fontWeight: '700', color: 'var(--accent)' }}>#{myRank}</div>
                </div>
                <div style={{ borderRight: '1px solid var(--border)' }}>
                  <div style={{ fontSize: '0.75rem', textTransform: 'uppercase', color: 'var(--text-mute)', marginBottom: '0.25rem' }}>Wins / Points</div>
                  <div style={{ fontSize: '1.5rem', fontWeight: '700', color: 'var(--accent)' }}>{myStanding.points}</div>
                </div>
                <div>
                  <div style={{ fontSize: '0.75rem', textTransform: 'uppercase', color: 'var(--text-mute)', marginBottom: '0.25rem' }}>Speaker Points</div>
                  <div style={{ fontSize: '1.5rem', fontWeight: '700', color: 'var(--accent)' }}>{myStanding.speaker_points.toFixed(1)}</div>
                </div>
              </div>
            </div>
          )}

          {/* Match Timeline / Schedule */}
          <div className="card">
            <h3>Match Schedule &amp; Results</h3>
            {debates.length === 0 ? (
              <div style={{ padding: '2rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>
                No matches released for your team yet. Check back when the round draw is published!
              </div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
                {debates.map(d => (
                  <div key={d.id} style={{ padding: '1.25rem', background: 'rgba(0,0,0,0.01)', border: '1px solid var(--border)', borderRadius: '8px' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '1px solid var(--border)', paddingBottom: '0.5rem', marginBottom: '0.75rem' }}>
                      <span style={{ fontWeight: '700', color: 'var(--text-h)', fontSize: '0.95rem' }}>{d.round_name}</span>
                      <span style={{ fontSize: '0.85rem', color: 'var(--accent)', fontWeight: '600', display: 'inline-flex', alignItems: 'center', gap: '0.25rem' }}>
                        <MapPin size={14} /> {d.venue}
                      </span>
                    </div>

                    {d.motion && (
                      <div style={{ marginBottom: '0.75rem', padding: '0.6rem 0.8rem', background: 'rgba(0,0,0,0.02)', borderRadius: '6px', fontSize: '0.85rem' }}>
                        <strong style={{ color: 'var(--text-mute)', fontSize: '0.75rem', textTransform: 'uppercase' }}>Motion: </strong>
                        <span>{d.motion}</span>
                      </div>
                    )}

                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
                      <span style={{ fontSize: '0.85rem', color: 'var(--text-mute)' }}>Your Position:</span>
                      <span className="badge badge-info" style={{ textTransform: 'uppercase' }}>{d.side}</span>
                    </div>

                    <div style={{ fontSize: '0.8rem', color: 'var(--text-mute)', marginBottom: '0.4rem' }}>
                      <strong>Room Opponents: </strong>
                      {d.teams.filter((t: any) => t.team_id !== tokenInfo.owner_id).map((t: any) => `${t.team_name} (${t.side})`).join(' • ') || 'None'}
                    </div>

                    {(d.chair || d.panellists?.length > 0) && (
                      <div style={{ fontSize: '0.8rem', color: 'var(--text-mute)', marginBottom: '0.4rem' }}>
                        <strong>Panel: </strong>
                        {d.chair ? `[Chair] ${d.chair}` : ''}
                        {d.panellists?.length > 0 ? ` • ${d.panellists.join(', ')}` : ''}
                      </div>
                    )}

                    {d.points !== undefined && (
                      <div style={{ marginTop: '0.75rem', paddingTop: '0.75rem', borderTop: '1px dashed var(--border)' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: d.speaker_scores?.length > 0 ? '0.5rem' : '0' }}>
                          <span style={{ fontSize: '0.85rem', color: 'var(--text-mute)', fontWeight: 600 }}>Ballot Decision:</span>
                          <span style={{ fontWeight: '700', color: '#16a34a', fontSize: '0.9rem' }}>
                            {d.points} Points | {d.speaker_points?.toFixed(1)} Total Speaker Pts
                          </span>
                        </div>

                        {/* Individual Speaker Breakdown */}
                        {d.speaker_scores?.length > 0 && (
                          <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap', marginTop: '0.4rem' }}>
                            {d.speaker_scores.map((sc: any, idx: number) => (
                              <div
                                key={idx}
                                style={{
                                  padding: '0.25rem 0.6rem',
                                  background: 'rgba(0,0,0,0.03)',
                                  borderRadius: '4px',
                                  fontSize: '0.8rem',
                                  display: 'inline-flex',
                                  alignItems: 'center',
                                  gap: '0.35rem'
                                }}
                              >
                                <span style={{ color: 'var(--text-mute)' }}>
                                  {sc.role ? `${sc.role} (${sc.speaker_name})` : sc.speaker_name}{sc.is_reply ? ' (R)' : ''}:
                                </span>
                                <strong style={{ color: 'var(--text-h)' }}>{sc.score.toFixed(1)}</strong>
                              </div>
                            ))}
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>

          <FeedbackSection token={token!} />
        </>
      )}

      {activeTab === 'speakers' && (
        <div className="card">
          <h3>Speaker Standings</h3>
          <div className="table-wrapper">
            <table className="table">
              <thead>
                <tr>
                  <th>Rank</th>
                  <th>Speaker</th>
                  <th>Team</th>
                  <th>Speeches</th>
                  <th>Total</th>
                  <th>Avg</th>
                </tr>
              </thead>
              <tbody>
                {speakerStandings.map(sp => {
                  const isMyTeam = sp.team_id === tokenInfo?.owner_id;
                  return (
                    <tr key={sp.speaker_id} style={{ background: isMyTeam ? 'rgba(59,130,246,0.08)' : undefined }}>
                      <td>{sp.rank}</td>
                      <td style={{ fontWeight: isMyTeam ? '600' : '500', color: isMyTeam ? 'var(--accent)' : 'var(--text-h)' }}>
                        <button
                          type="button"
                          onClick={() => setTrajectoryModal({ type: 'speaker', id: sp.speaker_id })}
                          style={{
                            background: 'none',
                            border: 'none',
                            padding: 0,
                            textAlign: 'left',
                            cursor: 'pointer',
                            fontWeight: 600,
                            color: isMyTeam ? 'var(--accent)' : 'var(--text-h)',
                            textDecoration: 'underline',
                            textDecorationStyle: 'dotted'
                          }}
                          title="View Speaker Trajectory"
                        >
                          {sp.speaker_name}
                        </button>
                        {isMyTeam && <span className="badge badge-info" style={{ marginLeft: '0.4rem', fontSize: '0.65rem' }}>YOU</span>}
                      </td>
                      <td>{sp.team_name}</td>
                      <td>{sp.speech_count}</td>
                      <td style={{ fontWeight: '600' }}>{sp.total_score.toFixed(1)}</td>
                      <td>{sp.average_score.toFixed(2)}</td>
                    </tr>
                  );
                })}
                {speakerStandings.length === 0 && (
                  <tr>
                    <td colSpan={6} style={{ textAlign: 'center', padding: '2rem 0', color: 'var(--text-mute)' }}>
                      No speaker standings published yet.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Trajectory Modal */}
      {trajectoryModal && (
        <TrajectoryModal
          slug={tokenInfo?.slug || ''}
          type={trajectoryModal.type}
          id={trajectoryModal.id}
          onClose={() => setTrajectoryModal(null)}
        />
      )}
    </div>
  );
}
