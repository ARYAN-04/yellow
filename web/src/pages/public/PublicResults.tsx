import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { FileText, MapPin, Accessibility, CalendarDays } from 'lucide-react';
import { fetchAPI } from '../../lib/api';
import PublicNav from '../../components/PublicNav';
import TrajectoryModal from '../../components/TrajectoryModal';

export default function PublicResults() {
  const { slug } = useParams<{ slug: string }>();
  const [selectedRoundId, setSelectedRoundId] = useState<string>('');
  const [trajectoryModal, setTrajectoryModal] = useState<{ type: 'team' | 'speaker' | 'adjudicator'; id: string } | null>(null);

  const { data: rounds = [], isLoading: loadingRounds } = useQuery<any[]>({
    queryKey: ['rounds', slug],
    queryFn: () => fetchAPI(`/api/t/${slug}/rounds`),
  });

  // Filter to rounds with released results
  const releasedRounds = (rounds || []).filter(r => r.results_released);

  useEffect(() => {
    if (releasedRounds.length > 0 && !selectedRoundId) {
      setSelectedRoundId(releasedRounds[releasedRounds.length - 1].id);
    }
  }, [releasedRounds, selectedRoundId]);

  const { data: ballots = [], isLoading: loadingBallots } = useQuery<any[]>({
    queryKey: ['round-ballots', slug, selectedRoundId],
    queryFn: () => fetchAPI(`/api/t/${slug}/rounds/${selectedRoundId}/ballots`),
    enabled: !!selectedRoundId,
  });

  const { data: draw = [] } = useQuery<any[]>({
    queryKey: ['draw', slug, selectedRoundId],
    queryFn: () => fetchAPI(`/api/t/${slug}/rounds/${selectedRoundId}/draw`),
    enabled: !!selectedRoundId,
  });

  // Map confirmed results by debate ID
  const debateResultsMap = new Map<string, any>();
  (ballots || []).forEach(b => {
    if (b.status === 'confirmed') {
      debateResultsMap.set(b.debate_id, b);
    }
  });

  const selectedRound = (rounds || []).find(r => r.id === selectedRoundId);

  return (
    <div style={{ minHeight: '100vh', background: 'var(--bg)' }}>
      <PublicNav title="Public Results" />

      <div className="container" style={{ maxWidth: '1000px', paddingBottom: '3rem' }}>
        <div className="card">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem', flexWrap: 'wrap', gap: '1rem' }}>
            <div>
              <h2 style={{ margin: 0, display: 'inline-flex', alignItems: 'center', gap: '0.4rem', fontSize: '1.4rem' }}>
                <FileText size={22} color="var(--accent)" /> Released Round Results
              </h2>
              <p style={{ fontSize: '0.85rem', color: 'var(--text-mute)', margin: '0.35rem 0 0 0' }}>
                Official confirmed debate outcomes and speaker scores by round.
              </p>
            </div>

            {/* Round Selector Tabs */}
            {releasedRounds.length > 0 && (
              <div className="tabs" style={{ margin: 0 }}>
                {releasedRounds.map(r => (
                  <button
                    key={r.id}
                    className={`tab-btn ${selectedRoundId === r.id ? 'active' : ''}`}
                    onClick={() => setSelectedRoundId(r.id)}
                  >
                    <CalendarDays size={13} style={{ marginRight: '4px', verticalAlign: 'middle' }} />
                    {r.name}
                  </button>
                ))}
              </div>
            )}
          </div>

          {loadingRounds ? (
            <div style={{ padding: '3rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>Loading rounds...</div>
          ) : releasedRounds.length === 0 ? (
            <div style={{ padding: '3rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>
              No round results have been officially released to the public yet.
            </div>
          ) : loadingBallots ? (
            <div style={{ padding: '3rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>Loading results for {selectedRound?.name}...</div>
          ) : (draw || []).length === 0 ? (
            <div style={{ padding: '3rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>
              No debate matches found for this round.
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              {(draw || []).map((d: any) => {
                const ballot = debateResultsMap.get(d.id);
                const resultsByTeam = new Map<string, any>();
                (ballot?.results || []).forEach((r: any) => {
                  resultsByTeam.set(r.team_id, r);
                });

                return (
                  <div
                    key={d.id}
                    className="card"
                    style={{
                      border: '1px solid var(--border)',
                      padding: '1.25rem',
                      background: '#fff',
                    }}
                  >
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem', borderBottom: '1px solid var(--border)', paddingBottom: '0.5rem' }}>
                      <span style={{ fontWeight: 600, color: 'var(--text-h)', display: 'inline-flex', alignItems: 'center', gap: '0.4rem' }}>
                        <MapPin size={15} /> {d.venue}
                        {d.venue_accessible && (
                          <span title="Wheelchair Accessible Venue" style={{ display: 'inline-flex', alignItems: 'center', gap: '0.2rem', fontSize: '0.65rem', background: '#dbeafe', color: '#1d4ed8', padding: '1px 6px', borderRadius: '999px', fontWeight: 600 }}>
                            <Accessibility size={11} /> Accessible
                          </span>
                        )}
                      </span>

                      {ballot ? (
                        <span className="badge badge-success" style={{ fontSize: '0.7rem' }}>
                          Confirmed
                        </span>
                      ) : (
                        <span className="badge" style={{ fontSize: '0.7rem', background: 'rgba(0,0,0,0.05)', color: 'var(--text-mute)' }}>
                          Pending
                        </span>
                      )}
                    </div>

                    {/* Teams in Debate */}
                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: '0.75rem' }}>
                      {(d.teams || []).map((t: any) => {
                        const res = resultsByTeam.get(t.team_id);
                        const isWin = res && res.points > 0;

                        return (
                          <div
                            key={t.team_id}
                            style={{
                              padding: '0.75rem',
                              borderRadius: '6px',
                              border: '1px solid var(--border)',
                              background: isWin ? 'rgba(22,163,74,0.04)' : 'rgba(0,0,0,0.01)',
                            }}
                          >
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.35rem' }}>
                              <button
                                type="button"
                                onClick={() => setTrajectoryModal({ type: 'team', id: t.team_id })}
                                style={{
                                  background: 'none',
                                  border: 'none',
                                  padding: 0,
                                  textAlign: 'left',
                                  cursor: 'pointer',
                                  fontWeight: 600,
                                  color: 'var(--text-h)',
                                  fontSize: '0.9rem',
                                  textDecoration: 'underline',
                                  textDecorationStyle: 'dotted',
                                }}
                              >
                                {t.team_name}
                              </button>
                              <span className="badge badge-info" style={{ textTransform: 'uppercase', fontSize: '0.65rem' }}>
                                {t.side}
                              </span>
                            </div>

                            {res ? (
                              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: '0.8rem', marginTop: '0.4rem', color: 'var(--text-mute)' }}>
                                <span style={{ fontWeight: 700, color: isWin ? '#16a34a' : 'var(--text-h)' }}>
                                  {res.points} {res.points === 1 ? 'pt' : 'pts'}
                                </span>
                                <span>{res.speaker_points?.toFixed(1) || '0.0'} spkrs</span>
                              </div>
                            ) : (
                              <div style={{ fontSize: '0.75rem', color: 'var(--text-mute)', marginTop: '0.3rem' }}>
                                Scores pending release
                              </div>
                            )}

                            {/* Speaker scores breakdown if available */}
                            {res?.speaker_scores?.length > 0 && (
                              <div style={{ marginTop: '0.4rem', paddingTop: '0.4rem', borderTop: '1px solid var(--border)', display: 'flex', flexDirection: 'column', gap: '0.2rem' }}>
                                {res.speaker_scores.map((sc: any, idx: number) => (
                                  <div key={idx} style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.75rem' }}>
                                    <span style={{ color: 'var(--text-mute)' }}>
                                      {sc.role || 'Speaker'}:
                                    </span>
                                    <span style={{ fontWeight: 600 }}>{sc.score?.toFixed(1)}</span>
                                  </div>
                                ))}
                              </div>
                            )}
                          </div>
                        );
                      })}
                    </div>

                    <div style={{ marginTop: '0.75rem', fontSize: '0.78rem', color: 'var(--text-mute)' }}>
                      <strong>Panel: </strong>
                      {(d.adjudicators || []).map((a: any) => `${a.adjudicator_name}${a.role === 'chair' ? ' (Chair)' : ''}`).join(', ') || 'None assigned'}
                    </div>
                  </div>
                );
              })}
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
