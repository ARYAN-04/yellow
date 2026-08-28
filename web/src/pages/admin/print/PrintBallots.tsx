import { useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Printer, ArrowLeft } from 'lucide-react';
import { fetchAPI } from '../../../lib/api';
import { getRoleSlotsForSide } from '../../../lib/roles';

export default function PrintBallots() {
  const { slug, roundId } = useParams<{ slug: string; roundId: string }>();
  const navigate = useNavigate();

  const { data: tournaments = [] } = useQuery<any[]>({
    queryKey: ['tournaments'],
    queryFn: () => fetchAPI('/api/tournaments'),
  });

  const { data: rounds = [] } = useQuery<any[]>({
    queryKey: ['rounds', slug],
    queryFn: () => fetchAPI(`/api/t/${slug}/rounds`),
  });

  const { data: draw = [], isLoading } = useQuery<any[]>({
    queryKey: ['draw', slug, roundId],
    queryFn: () => fetchAPI(`/api/t/${slug}/rounds/${roundId}/draw`),
  });

  const currentTourney = tournaments.find((t: any) => t.slug === slug);
  const currentRound = rounds.find((r: any) => r.id === roundId);

  useEffect(() => {
    document.title = `Print Ballots - ${currentRound?.name || 'Round'} - ${currentTourney?.name || slug}`;
  }, [currentRound, currentTourney, slug]);

  return (
    <div style={{ minHeight: '100vh', background: '#fff', color: '#000', padding: '1.5rem', fontFamily: 'system-ui, -apple-system, sans-serif' }}>
      {/* Print Controls (Hidden in Print) */}
      <div className="no-print" style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        background: '#f4f4f5',
        padding: '0.75rem 1.25rem',
        borderRadius: '8px',
        marginBottom: '1.5rem',
        border: '1px solid #e4e4e7'
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          <button
            onClick={() => navigate(-1)}
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '0.4rem',
              background: '#fff',
              border: '1px solid #d4d4d8',
              padding: '0.4rem 0.8rem',
              borderRadius: '6px',
              cursor: 'pointer',
              fontSize: '0.85rem'
            }}
          >
            <ArrowLeft size={16} /> Back
          </button>
          <span style={{ fontWeight: 600, fontSize: '0.9rem' }}>Debate Official Score Sheet / Ballot Printer</span>
        </div>
        <button
          onClick={() => window.print()}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: '0.5rem',
            background: '#18181b',
            color: '#fff',
            border: 'none',
            padding: '0.5rem 1rem',
            borderRadius: '6px',
            cursor: 'pointer',
            fontWeight: 600,
            fontSize: '0.85rem'
          }}
        >
          <Printer size={16} /> Print All Ballots ({draw.length} Debates)
        </button>
      </div>

      {isLoading ? (
        <div style={{ textAlign: 'center', padding: '3rem 0', color: '#71717a' }}>Loading ballot sheets...</div>
      ) : draw.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '3rem 0', color: '#71717a' }}>No debates found for this round.</div>
      ) : (
        draw.map((d: any, dIdx: number) => {
          const chairs = d.adjudicators?.filter((a: any) => a.role === 'chair') || [];
          const panelists = d.adjudicators?.filter((a: any) => a.role === 'panel') || [];
          const trainees = d.adjudicators?.filter((a: any) => a.role === 'trainee') || [];
          const isLast = dIdx === draw.length - 1;

          return (
            <div
              key={d.id}
              className="ballot-page"
              style={{
                pageBreakAfter: isLast ? 'auto' : 'always',
                breakAfter: isLast ? 'auto' : 'page',
                marginBottom: isLast ? 0 : '3rem',
                paddingBottom: isLast ? 0 : '2rem',
                borderBottom: isLast ? 'none' : '1px dashed #cbd5e1'
              }}
            >
              {/* Header */}
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', borderBottom: '3px solid #000', paddingBottom: '0.75rem', marginBottom: '1rem' }}>
                <div>
                  <div style={{ fontSize: '0.8rem', textTransform: 'uppercase', letterSpacing: '0.05em', color: '#52525b', fontWeight: 700 }}>
                    Official Debate Ballot
                  </div>
                  <h1 style={{ fontSize: '1.5rem', fontWeight: 800, margin: '2px 0 0 0' }}>
                    {currentTourney?.name || slug}
                  </h1>
                </div>
                <div style={{ textAlign: 'right' }}>
                  <div style={{ fontSize: '1.2rem', fontWeight: 800, color: '#18181b' }}>
                    {currentRound?.name || 'Round'}
                  </div>
                  <div style={{ fontSize: '1.1rem', fontWeight: 700, color: '#2563eb' }}>
                    Venue: {d.venue}
                  </div>
                </div>
              </div>

              {/* Panel & Adjudicator Selector */}
              <div style={{
                background: '#f8fafc',
                border: '1px solid #cbd5e1',
                borderRadius: '6px',
                padding: '0.75rem 1rem',
                marginBottom: '1.25rem',
                fontSize: '0.85rem'
              }}>
                <div style={{ fontWeight: 700, marginBottom: '0.35rem' }}>Adjudication Panel (Please check your name):</div>
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: '1.5rem' }}>
                  {chairs.map((c: any) => (
                    <div key={c.adjudicator_id} style={{ display: 'inline-flex', alignItems: 'center', gap: '0.35rem' }}>
                      <div style={{ width: '14px', height: '14px', border: '1.5px solid #000', borderRadius: '3px' }} />
                      <span style={{ fontWeight: 700 }}>[Chair] {c.adjudicator_name}</span>
                    </div>
                  ))}
                  {panelists.map((p: any) => (
                    <div key={p.adjudicator_id} style={{ display: 'inline-flex', alignItems: 'center', gap: '0.35rem' }}>
                      <div style={{ width: '14px', height: '14px', border: '1.5px solid #000', borderRadius: '3px' }} />
                      <span>[Panel] {p.adjudicator_name}</span>
                    </div>
                  ))}
                  {trainees.map((t: any) => (
                    <div key={t.adjudicator_id} style={{ display: 'inline-flex', alignItems: 'center', gap: '0.35rem', color: '#64748b' }}>
                      <div style={{ width: '14px', height: '14px', border: '1.5px solid #64748b', borderRadius: '3px' }} />
                      <span>[Trainee] {t.adjudicator_name}</span>
                    </div>
                  ))}
                </div>
              </div>

              {/* Scoring Table */}
              <table style={{
                width: '100%',
                borderCollapse: 'collapse',
                fontSize: '0.85rem',
                marginBottom: '1.25rem'
              }}>
                <thead>
                  <tr style={{ background: '#f1f5f9', borderTop: '2px solid #000', borderBottom: '2px solid #000' }}>
                    <th style={{ padding: '8px 10px', border: '1px solid #94a3b8', width: '12%', textAlign: 'left' }}>Side</th>
                    <th style={{ padding: '8px 10px', border: '1px solid #94a3b8', width: '34%', textAlign: 'left' }}>Team &amp; Speakers</th>
                    <th style={{ padding: '8px 10px', border: '1px solid #94a3b8', width: '18%', textAlign: 'center' }}>Speaker Scores</th>
                    <th style={{ padding: '8px 10px', border: '1px solid #94a3b8', width: '18%', textAlign: 'center' }}>Total Spkr Pts</th>
                    <th style={{ padding: '8px 10px', border: '1px solid #94a3b8', width: '18%', textAlign: 'center' }}>Rank / Points</th>
                  </tr>
                </thead>
                <tbody>
                  {d.teams?.map((t: any) => {
                    const roleSlots = getRoleSlotsForSide(t.side, d.teams?.length || 4);

                    return (
                      <tr key={t.team_id} style={{ borderBottom: '1.5px solid #94a3b8', pageBreakInside: 'avoid' }}>
                        {/* Side */}
                        <td style={{ padding: '12px 10px', border: '1px solid #94a3b8', verticalAlign: 'top', fontWeight: 800, fontSize: '1rem' }}>
                          {t.side}
                        </td>

                        {/* Team & Speaker Lines */}
                        <td style={{ padding: '12px 10px', border: '1px solid #94a3b8', verticalAlign: 'top' }}>
                          <div style={{ fontWeight: 700, fontSize: '0.95rem', marginBottom: '8px' }}>
                            {t.team_name}
                          </div>
                          <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                            {roleSlots.map((slot, sIdx) => (
                              <div key={sIdx} style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                                <span style={{ fontSize: '0.75rem', color: '#475569', width: '45px', fontWeight: 600 }}>
                                  {slot.role}:
                                </span>
                                <div style={{ borderBottom: '1px dotted #000', flex: 1, height: '16px' }} />
                              </div>
                            ))}
                          </div>
                        </td>

                        {/* Speaker Scores boxes */}
                        <td style={{ padding: '12px 10px', border: '1px solid #94a3b8', verticalAlign: 'middle', textAlign: 'center' }}>
                          <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', alignItems: 'center' }}>
                            {roleSlots.map((_, sIdx) => (
                              <div key={sIdx} style={{ width: '60px', height: '22px', border: '1px solid #475569', borderRadius: '4px', background: '#fff' }} />
                            ))}
                          </div>
                        </td>

                        {/* Total Speaker Points Box */}
                        <td style={{ padding: '12px 10px', border: '1px solid #94a3b8', verticalAlign: 'middle', textAlign: 'center' }}>
                          <div style={{
                            width: '80px',
                            height: '50px',
                            border: '2px solid #000',
                            borderRadius: '6px',
                            margin: '0 auto',
                            background: '#fff'
                          }} />
                        </td>

                        {/* Team Rank / Win-Loss Box */}
                        <td style={{ padding: '12px 10px', border: '1px solid #94a3b8', verticalAlign: 'middle', textAlign: 'center' }}>
                          <div style={{
                            width: '80px',
                            height: '50px',
                            border: '2px solid #000',
                            borderRadius: '6px',
                            margin: '0 auto',
                            background: '#fff'
                          }} />
                          <div style={{ fontSize: '0.7rem', color: '#64748b', marginTop: '4px' }}>(1st, 2nd, 3rd, 4th / Win)</div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>

              {/* Feedback Reminder & Adjudicator Sign-off */}
              <div style={{
                display: 'grid',
                gridTemplateColumns: '1.2fr 1fr',
                gap: '1.5rem',
                border: '1.5px solid #000',
                borderRadius: '6px',
                padding: '1rem',
                background: '#fafafa',
                fontSize: '0.85rem'
              }}>
                <div>
                  <div style={{ fontWeight: 800, textTransform: 'uppercase', fontSize: '0.75rem', marginBottom: '0.25rem', color: '#dc2626' }}>
                    Important Instructions:
                  </div>
                  <ul style={{ margin: 0, paddingLeft: '1.2rem', lineHeight: '1.4', fontSize: '0.8rem' }}>
                    <li>Ensure all individual speaker scores and team ranks are filled clearly.</li>
                    <li>Verify math totals before signing this sheet.</li>
                    <li><strong>Adjudicator Feedback:</strong> Please remember to submit constructive peer and chair feedback promptly after this round.</li>
                  </ul>
                </div>

                <div style={{ display: 'flex', flexDirection: 'column', justifyContent: 'flex-end', gap: '0.75rem' }}>
                  <div style={{ display: 'flex', alignItems: 'flex-end', gap: '0.5rem' }}>
                    <span style={{ fontWeight: 700, fontSize: '0.8rem' }}>Signature:</span>
                    <div style={{ borderBottom: '1.5px solid #000', flex: 1, height: '18px' }} />
                  </div>
                  <div style={{ display: 'flex', alignItems: 'flex-end', gap: '0.5rem' }}>
                    <span style={{ fontWeight: 700, fontSize: '0.8rem' }}>Date/Time:</span>
                    <div style={{ borderBottom: '1.5px solid #000', flex: 1, height: '18px' }} />
                  </div>
                </div>
              </div>
            </div>
          );
        })
      )}

      {/* Print Stylesheet */}
      <style>{`
        @media print {
          .no-print {
            display: none !important;
          }
          body {
            background: #fff !important;
            color: #000 !important;
            margin: 0 !important;
            padding: 0 !important;
          }
          .ballot-page {
            page-break-after: always !important;
            break-after: page !important;
            border-bottom: none !important;
            padding-bottom: 0 !important;
            margin-bottom: 0 !important;
            min-height: 98vh;
            display: flex;
            flex-direction: column;
            justifyContent: space-between;
          }
          @page {
            size: portrait;
            margin: 1.2cm;
          }
        }
      `}</style>
    </div>
  );
}
