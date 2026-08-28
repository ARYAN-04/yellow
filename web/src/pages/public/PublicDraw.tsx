import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Shuffle, MapPin, Accessibility, CalendarDays, Users } from 'lucide-react';
import { fetchAPI } from '../../lib/api';
import PublicNav from '../../components/PublicNav';
import TrajectoryModal from '../../components/TrajectoryModal';

export default function PublicDraw() {
  const { slug } = useParams<{ slug: string }>();
  const [selectedRoundId, setSelectedRoundId] = useState<string>('');
  const [trajectoryModal, setTrajectoryModal] = useState<{ type: 'team' | 'speaker' | 'adjudicator'; id: string } | null>(null);

  const { data: rounds = [], isLoading: loadingRounds } = useQuery<any[]>({
    queryKey: ['rounds', slug],
    queryFn: () => fetchAPI(`/api/t/${slug}/rounds`),
  });

  // Filter to rounds with released draw
  const releasedRounds = (rounds || []).filter(r => r.draw_released);

  useEffect(() => {
    if (releasedRounds.length > 0 && !selectedRoundId) {
      setSelectedRoundId(releasedRounds[releasedRounds.length - 1].id);
    }
  }, [releasedRounds, selectedRoundId]);

  const { data: draw = [], isLoading: loadingDraw } = useQuery<any[]>({
    queryKey: ['draw', slug, selectedRoundId],
    queryFn: () => fetchAPI(`/api/t/${slug}/rounds/${selectedRoundId}/draw`),
    enabled: !!selectedRoundId,
  });

  const selectedRound = (rounds || []).find(r => r.id === selectedRoundId);

  return (
    <div style={{ minHeight: '100vh', background: 'var(--bg)' }}>
      <PublicNav title="Public Draw" />

      <div className="container" style={{ maxWidth: '1000px', paddingBottom: '3rem' }}>
        <div className="card">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem', flexWrap: 'wrap', gap: '1rem' }}>
            <div>
              <h2 style={{ margin: 0, display: 'inline-flex', alignItems: 'center', gap: '0.4rem', fontSize: '1.4rem' }}>
                <Shuffle size={22} color="var(--accent)" /> Released Round Matchups &amp; Draw
              </h2>
              <p style={{ fontSize: '0.85rem', color: 'var(--text-mute)', margin: '0.35rem 0 0 0' }}>
                Official team room allocations, side assignments, and adjudicator panels.
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
              No round draws have been released to the public yet.
            </div>
          ) : loadingDraw ? (
            <div style={{ padding: '3rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>Loading draw for {selectedRound?.name}...</div>
          ) : (draw || []).length === 0 ? (
            <div style={{ padding: '3rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>
              No debate matches found for this round.
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              {(draw || []).map((d: any) => (
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
                  </div>

                  {/* BP 4-Team / 2-Team Matchup Cards */}
                  <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '0.75rem' }}>
                    {(d.teams || []).map((t: any) => (
                      <div
                        key={t.team_id}
                        style={{
                          display: 'flex',
                          justifyContent: 'space-between',
                          alignItems: 'center',
                          padding: '0.6rem 0.75rem',
                          borderRadius: '6px',
                          background: 'rgba(0,0,0,0.02)',
                          border: '1px solid var(--border)',
                        }}
                      >
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
                            fontSize: '0.88rem',
                            textDecoration: 'underline',
                            textDecorationStyle: 'dotted',
                          }}
                        >
                          {t.team_name} {t.pull_up ? '(PU)' : ''}
                        </button>
                        <span className="badge badge-info" style={{ textTransform: 'uppercase', fontSize: '0.65rem' }}>
                          {t.side}
                        </span>
                      </div>
                    ))}
                  </div>

                  {/* Panel Breakdown */}
                  <div style={{ marginTop: '0.75rem', paddingTop: '0.6rem', borderTop: '1px solid rgba(0,0,0,0.06)', fontSize: '0.8rem', color: 'var(--text-mute)', display: 'flex', alignItems: 'center', gap: '0.4rem', flexWrap: 'wrap' }}>
                    <Users size={14} />
                    <strong>Panel: </strong>
                    {(d.adjudicators || []).map((a: any) => (
                      <span key={a.id ?? a.adjudicator_id} style={{ display: 'inline-flex', alignItems: 'center', gap: '0.2rem' }}>
                        <button
                          type="button"
                          onClick={() => setTrajectoryModal({ type: 'adjudicator', id: a.adjudicator_id })}
                          style={{
                            background: 'none',
                            border: 'none',
                            padding: 0,
                            cursor: 'pointer',
                            color: 'var(--text-h)',
                            fontWeight: a.role === 'chair' ? 700 : 500,
                            textDecoration: 'underline',
                            textDecorationStyle: 'dotted',
                          }}
                        >
                          {a.adjudicator_name}
                        </button>
                        {a.role === 'chair' && <span className="badge badge-secondary" style={{ fontSize: '0.6rem' }}>Ⓒ</span>}
                        {a.role === 'trainee' && <span className="badge" style={{ fontSize: '0.6rem', background: '#f3f4f6' }}>T</span>}
                      </span>
                    ))}
                    {(d.adjudicators || []).length === 0 && <span>No adjudicators allocated</span>}
                  </div>
                </div>
              ))}
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
