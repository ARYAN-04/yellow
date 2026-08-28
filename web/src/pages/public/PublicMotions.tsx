import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Megaphone, Info, CalendarDays, BarChart2, CheckCircle2 } from 'lucide-react';
import { fetchAPI } from '../../lib/api';
import PublicNav from '../../components/PublicNav';

export default function PublicMotions() {
  const { slug } = useParams<{ slug: string }>();
  const [selectedRoundId, setSelectedRoundId] = useState<string>('');
  const [showStats, setShowStats] = useState(false);

  const { data: rounds = [], isLoading: loadingRounds } = useQuery<any[]>({
    queryKey: ['rounds', slug],
    queryFn: () => fetchAPI(`/api/t/${slug}/rounds`),
  });

  useEffect(() => {
    if (rounds.length > 0 && !selectedRoundId) {
      setSelectedRoundId(rounds[rounds.length - 1].id);
    }
  }, [rounds, selectedRoundId]);

  const { data: motions = [], isLoading: loadingMotions } = useQuery<any[]>({
    queryKey: ['round-motions', slug, selectedRoundId],
    queryFn: () => fetchAPI(`/api/t/${slug}/rounds/${selectedRoundId}/motions`),
    enabled: !!selectedRoundId,
  });

  const { data: motionStats = [] } = useQuery<any[]>({
    queryKey: ['motion-stats', slug],
    queryFn: () => fetchAPI(`/api/t/${slug}/motions/statistics`),
    enabled: showStats,
  });

  // Filter only released motions for public view
  const releasedMotions = (motions || []).filter(m => !!m.released_at);
  const selectedRound = (rounds || []).find(r => r.id === selectedRoundId);

  return (
    <div style={{ minHeight: '100vh', background: 'var(--bg)' }}>
      <PublicNav title="Public Motions" />

      <div className="container" style={{ maxWidth: '1000px', paddingBottom: '3rem' }}>
        <div className="card">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem', flexWrap: 'wrap', gap: '1rem' }}>
            <div>
              <h2 style={{ margin: 0, display: 'inline-flex', alignItems: 'center', gap: '0.4rem', fontSize: '1.4rem' }}>
                <Megaphone size={22} color="var(--accent)" /> Tournament Motions &amp; Info Slides
              </h2>
              <p style={{ fontSize: '0.85rem', color: 'var(--text-mute)', margin: '0.35rem 0 0 0' }}>
                Official motion releases and contextual background info slides by round.
              </p>
            </div>

            <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center', flexWrap: 'wrap' }}>
              <button
                className={`btn ${showStats ? 'btn-primary' : 'btn-secondary'}`}
                onClick={() => setShowStats(!showStats)}
                style={{ fontSize: '0.8rem', padding: '0.35rem 0.75rem', display: 'inline-flex', alignItems: 'center', gap: '0.35rem' }}
              >
                <BarChart2 size={14} /> Motion Balance Stats
              </button>

              {/* Round Selector Tabs */}
              {(rounds || []).length > 0 && !showStats && (
                <div className="tabs" style={{ margin: 0 }}>
                  {rounds.map(r => (
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
          </div>

          {showStats ? (
            <div>
              <h4 style={{ margin: '0 0 1rem 0' }}>Tournament Motion Balance Statistics</h4>
              {(!motionStats || motionStats.length === 0) ? (
                <div style={{ padding: '2.5rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>
                  No confirmed debate statistics available yet.
                </div>
              ) : (
                <div className="table-wrapper">
                  <table className="table">
                    <thead>
                      <tr>
                        <th>Motion</th>
                        <th>Round</th>
                        <th>Debates</th>
                        <th>Side Win Rates</th>
                      </tr>
                    </thead>
                    <tbody>
                      {motionStats.map((s: any) => (
                        <tr key={s.motion_id}>
                          <td style={{ fontWeight: 600, color: 'var(--text-h)', maxWidth: '380px' }}>
                            {s.reference ? `[${s.reference}] ` : ''}{s.text}
                          </td>
                          <td>{s.round_name}</td>
                          <td style={{ fontWeight: 700 }}>{s.total_debates}</td>
                          <td>
                            {s.total_debates === 0 ? (
                              <span style={{ color: 'var(--text-mute)' }}>No confirmed ballots</span>
                            ) : (
                              <div style={{ display: 'flex', gap: '0.4rem', flexWrap: 'wrap' }}>
                                {Object.entries(s.side_percentages || {}).map(([side, pct]: [string, any]) => (
                                  <span key={side} className="badge badge-info" style={{ fontSize: '0.75rem' }}>
                                    {side}: {pct}% ({s.side_wins?.[side] || 0} wins)
                                  </span>
                                ))}
                              </div>
                            )}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          ) : loadingRounds || loadingMotions ? (
            <div style={{ padding: '3rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>Loading motions...</div>
          ) : releasedMotions.length === 0 ? (
            <div style={{ padding: '3rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>
              No motions have been released for {selectedRound?.name || 'this round'} yet.
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
              {releasedMotions.map((m: any) => (
                <div
                  key={m.id}
                  className="card"
                  style={{
                    border: '1px solid var(--border)',
                    padding: '1.5rem',
                    background: '#fff',
                  }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.75rem' }}>
                    <span className="badge badge-info" style={{ fontWeight: 700 }}>
                      Motion #{m.seq}
                    </span>
                    {m.reference && (
                      <span style={{ fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-mute)' }}>
                        [{m.reference}]
                      </span>
                    )}
                    <span style={{ fontSize: '0.75rem', color: '#15803d', display: 'inline-flex', alignItems: 'center', gap: '0.2rem', marginLeft: 'auto' }}>
                      <CheckCircle2 size={12} /> Released {new Date(m.released_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                    </span>
                  </div>

                  <div style={{ fontSize: '1.25rem', fontWeight: 700, color: 'var(--text-h)', lineHeight: 1.45, marginBottom: m.info_slide ? '1rem' : 0 }}>
                    {m.text}
                  </div>

                  {m.info_slide && (
                    <div
                      style={{
                        marginTop: '1rem',
                        padding: '1rem',
                        background: 'rgba(0,0,0,0.02)',
                        borderLeft: '3px solid var(--accent)',
                        borderRadius: '6px',
                      }}
                    >
                      <div style={{ fontSize: '0.8rem', fontWeight: 700, color: 'var(--text-h)', display: 'inline-flex', alignItems: 'center', gap: '0.35rem', marginBottom: '0.4rem' }}>
                        <Info size={14} color="var(--accent)" /> Context / Info Slide
                      </div>
                      <div style={{ fontSize: '0.9rem', color: 'var(--text)', whiteSpace: 'pre-wrap', lineHeight: 1.5 }}>
                        {m.info_slide}
                      </div>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
