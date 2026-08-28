import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { TrendingUp, Users, User, Award } from 'lucide-react';
import { fetchAPI } from '../../lib/api';
import PublicNav from '../../components/PublicNav';
import TrajectoryModal from '../../components/TrajectoryModal';

const categoryTabs = [
  { key: '', label: 'Open' },
  { key: 'novice', label: 'Novice' },
  { key: 'esl', label: 'ESL' },
  { key: 'efl', label: 'EFL' },
];

export default function PublicStandings() {
  const { slug } = useParams<{ slug: string }>();
  const [activeTab, setActiveTab] = useState<'teams' | 'speakers' | 'adjudicators'>('teams');
  const [standings, setStandings] = useState<any[]>([]);
  const [speakerStandings, setSpeakerStandings] = useState<any[]>([]);
  const [adjudicatorStandings, setAdjudicatorStandings] = useState<any[]>([]);
  const [category, setCategory] = useState('');
  const [isTrimmed, setIsTrimmed] = useState(false);
  const [loading, setLoading] = useState(false);
  const [trajectoryModal, setTrajectoryModal] = useState<{ type: 'team' | 'speaker' | 'adjudicator'; id: string } | null>(null);

  useEffect(() => {
    setLoading(true);
    if (activeTab === 'teams') {
      const qs = category ? `?category=${category}` : '';
      fetchAPI(`/api/t/${slug}/standings${qs}`)
        .then(d => setStandings(d || []))
        .catch(console.error)
        .finally(() => setLoading(false));
    } else if (activeTab === 'speakers') {
      const params = new URLSearchParams();
      if (category) params.set('category', category);
      if (isTrimmed) params.set('trimmed', 'true');
      const qs = params.toString() ? `?${params.toString()}` : '';
      fetchAPI(`/api/t/${slug}/standings/speakers${qs}`)
        .then(d => setSpeakerStandings(d || []))
        .catch(console.error)
        .finally(() => setLoading(false));
    } else {
      fetchAPI(`/api/t/${slug}/standings/adjudicators`)
        .then(d => setAdjudicatorStandings(d || []))
        .catch(console.error)
        .finally(() => setLoading(false));
    }
  }, [slug, activeTab, category, isTrimmed]);

  return (
    <div style={{ minHeight: '100vh', background: 'var(--bg)' }}>
      <PublicNav title="Public Standings" />

      <div className="container" style={{ maxWidth: '1000px', paddingBottom: '3rem' }}>
        <div className="card">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem', flexWrap: 'wrap', gap: '1rem' }}>
            <div>
              <h2 style={{ margin: 0, display: 'inline-flex', alignItems: 'center', gap: '0.4rem', fontSize: '1.4rem' }}>
                <TrendingUp size={22} color="var(--accent)" /> Official Tournament Standings
              </h2>
              <p style={{ fontSize: '0.85rem', color: 'var(--text-mute)', margin: '0.35rem 0 0 0' }}>
                Live results and speaker rankings updated across confirmed preliminary rounds.
              </p>
            </div>

            <div style={{ display: 'inline-flex', border: '1px solid var(--border)', borderRadius: '8px', overflow: 'hidden' }}>
              <button
                className="btn btn-secondary"
                onClick={() => { setActiveTab('teams'); setCategory(''); }}
                style={{
                  border: 'none', borderRadius: 0, padding: '6px 14px', fontSize: '0.8rem',
                  background: activeTab === 'teams' ? 'var(--accent)' : undefined,
                  color: activeTab === 'teams' ? '#fff' : undefined,
                }}
              >
                <Users size={13} style={{ marginRight: '4px', verticalAlign: 'middle' }} />
                Team Standings
              </button>
              <button
                className="btn btn-secondary"
                onClick={() => { setActiveTab('speakers'); setCategory(''); }}
                style={{
                  border: 'none', borderRadius: 0, padding: '6px 14px', fontSize: '0.8rem',
                  background: activeTab === 'speakers' ? 'var(--accent)' : undefined,
                  color: activeTab === 'speakers' ? '#fff' : undefined,
                }}
              >
                <User size={13} style={{ marginRight: '4px', verticalAlign: 'middle' }} />
                Speaker Standings
              </button>
              <button
                className="btn btn-secondary"
                onClick={() => { setActiveTab('adjudicators'); setCategory(''); }}
                style={{
                  border: 'none', borderRadius: 0, padding: '6px 14px', fontSize: '0.8rem',
                  background: activeTab === 'adjudicators' ? 'var(--accent)' : undefined,
                  color: activeTab === 'adjudicators' ? '#fff' : undefined,
                }}
              >
                <Award size={13} style={{ marginRight: '4px', verticalAlign: 'middle' }} />
                Adjudicators
              </button>
            </div>
          </div>

          {activeTab !== 'adjudicators' && (
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem', flexWrap: 'wrap', gap: '0.75rem' }}>
              <div className="tabs" style={{ margin: 0 }}>
                {categoryTabs.map(t => (
                  <button key={t.key} className={`tab-btn ${category === t.key ? 'active' : ''}`} onClick={() => setCategory(t.key)}>
                    {t.label}
                  </button>
                ))}
              </div>
              {activeTab === 'speakers' && (
                <label style={{ display: 'inline-flex', alignItems: 'center', gap: '0.4rem', fontSize: '0.85rem', cursor: 'pointer' }}>
                  <input
                    type="checkbox"
                    checked={isTrimmed}
                    onChange={e => setIsTrimmed(e.target.checked)}
                  />
                  <span>Trimmed Mean (Drop High/Low)</span>
                </label>
              )}
            </div>
          )}

          {loading ? (
            <div style={{ padding: '3rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>Loading standings...</div>
          ) : activeTab === 'teams' ? (
            <div className="table-wrapper">
              <table className="table">
                <thead>
                  <tr>
                    <th>Rank</th>
                    <th>Team</th>
                    <th>Institution</th>
                    <th>Wins / Points</th>
                    <th>Speaker Points</th>
                    <th>Margin</th>
                  </tr>
                </thead>
                <tbody>
                  {(standings || []).map((team, idx) => (
                    <tr key={team.team_id}>
                      <td style={{ fontWeight: 600, color: 'var(--accent)' }}>{idx + 1}</td>
                      <td style={{ fontWeight: 500, color: 'var(--text-h)' }}>
                        <button
                          type="button"
                          onClick={() => setTrajectoryModal({ type: 'team', id: team.team_id })}
                          style={{
                            background: 'none',
                            border: 'none',
                            padding: 0,
                            textAlign: 'left',
                            cursor: 'pointer',
                            fontWeight: 600,
                            color: 'var(--text-h)',
                            textDecoration: 'underline',
                            textDecorationStyle: 'dotted',
                          }}
                          title="View Team Trajectory"
                        >
                          {team.team_name}
                        </button>
                      </td>
                      <td>{team.institution_code || 'Independent'}</td>
                      <td style={{ fontWeight: 700, color: 'var(--accent)' }}>{team.points}</td>
                      <td>{team.speaker_points?.toFixed(1) || '0.0'}</td>
                      <td>{(team.margin || 0).toFixed(1)}</td>
                    </tr>
                  ))}
                  {(!standings || standings.length === 0) && (
                    <tr>
                      <td colSpan={6} style={{ textAlign: 'center', padding: '2.5rem 0', color: 'var(--text-mute)' }}>
                        No published team standings available yet.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          ) : activeTab === 'speakers' ? (
            <div className="table-wrapper">
              <table className="table">
                <thead>
                  <tr>
                    <th>Rank</th>
                    <th>Speaker</th>
                    <th>Team</th>
                    <th>Institution</th>
                    <th>Speeches</th>
                    <th>Total Score</th>
                    <th>Avg Score</th>
                    <th>Trimmed Score</th>
                  </tr>
                </thead>
                <tbody>
                  {(speakerStandings || []).map((sp) => (
                    <tr key={sp.speaker_id}>
                      <td style={{ fontWeight: 600, color: 'var(--accent)' }}>{sp.rank}</td>
                      <td style={{ fontWeight: 500, color: 'var(--text-h)' }}>
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
                            color: 'var(--text-h)',
                            textDecoration: 'underline',
                            textDecorationStyle: 'dotted',
                          }}
                          title="View Speaker Trajectory"
                        >
                          {sp.speaker_name}
                        </button>
                        {sp.is_novice && <span className="badge badge-info" style={{ marginLeft: '0.4rem', fontSize: '0.65rem' }}>NOV</span>}
                        {sp.is_esl && <span className="badge badge-warning" style={{ marginLeft: '0.4rem', fontSize: '0.65rem' }}>ESL</span>}
                        {sp.is_efl && <span className="badge badge-warning" style={{ marginLeft: '0.4rem', fontSize: '0.65rem' }}>EFL</span>}
                      </td>
                      <td>{sp.team_name}</td>
                      <td>{sp.institution_code || 'Independent'}</td>
                      <td>{sp.speech_count}</td>
                      <td style={{ fontWeight: 700, color: 'var(--accent)' }}>{sp.total_score?.toFixed(1) || '0.0'}</td>
                      <td>{sp.average_score?.toFixed(2) || '0.00'}</td>
                      <td style={{ fontWeight: isTrimmed ? 700 : 'normal', color: isTrimmed ? 'var(--accent)' : undefined }}>
                        {sp.trimmed_score?.toFixed(2) || '0.00'}
                      </td>
                    </tr>
                  ))}
                  {(!speakerStandings || speakerStandings.length === 0) && (
                    <tr>
                      <td colSpan={8} style={{ textAlign: 'center', padding: '2.5rem 0', color: 'var(--text-mute)' }}>
                        No speaker standings available yet.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="table-wrapper">
              <table className="table">
                <thead>
                  <tr>
                    <th>Rank</th>
                    <th>Adjudicator</th>
                    <th>Institution</th>
                    <th>Test Score</th>
                    <th>Rating</th>
                    <th>Debates Judged</th>
                    <th>Avg Feedback</th>
                    <th>Action</th>
                  </tr>
                </thead>
                <tbody>
                  {(adjudicatorStandings || []).map((adj) => (
                    <tr key={adj.id}>
                      <td style={{ fontWeight: 600, color: 'var(--accent)' }}>{adj.rank}</td>
                      <td style={{ fontWeight: 500, color: 'var(--text-h)' }}>
                        <button
                          type="button"
                          onClick={() => setTrajectoryModal({ type: 'adjudicator', id: adj.id })}
                          style={{
                            background: 'none',
                            border: 'none',
                            padding: 0,
                            textAlign: 'left',
                            cursor: 'pointer',
                            fontWeight: 600,
                            color: 'var(--text-h)',
                            textDecoration: 'underline',
                            textDecorationStyle: 'dotted',
                          }}
                          title="View Judging Trajectory"
                        >
                          {adj.name}
                        </button>
                      </td>
                      <td>{adj.institution_code || adj.institution_name || 'Independent'}</td>
                      <td>{adj.test_score?.toFixed(1) || '0.0'}</td>
                      <td style={{ fontWeight: 600, color: adj.feedback_rating ? '#16a34a' : 'var(--text-h)' }}>
                        {adj.feedback_rating != null ? adj.feedback_rating.toFixed(2) : adj.test_score?.toFixed(1) || '-'}
                      </td>
                      <td>
                        {adj.debates_count} debates ({adj.chairs_count}C / {adj.panels_count}P / {adj.trainees_count}T)
                      </td>
                      <td>
                        {adj.average_feedback_score != null ? (
                          <span className="badge badge-info" style={{ fontWeight: 600 }}>
                            {adj.average_feedback_score.toFixed(2)}
                          </span>
                        ) : (
                          <span style={{ color: 'var(--text-mute)' }}>-</span>
                        )}
                      </td>
                      <td>
                        <button
                          type="button"
                          className="btn btn-secondary"
                          style={{ padding: '3px 8px', fontSize: '0.75rem' }}
                          onClick={() => setTrajectoryModal({ type: 'adjudicator', id: adj.id })}
                        >
                          Trajectory
                        </button>
                      </td>
                    </tr>
                  ))}
                  {(!adjudicatorStandings || adjudicatorStandings.length === 0) && (
                    <tr>
                      <td colSpan={8} style={{ textAlign: 'center', padding: '2.5rem 0', color: 'var(--text-mute)' }}>
                        No adjudicators recorded yet.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {trajectoryModal && (
        <TrajectoryModal
          slug={slug || ''}
          type={trajectoryModal.type}
          id={trajectoryModal.id}
          onClose={() => setTrajectoryModal(null)}
        />
      )}
    </div>
  );
}
