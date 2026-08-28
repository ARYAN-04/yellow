import { useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Printer, ArrowLeft, Accessibility } from 'lucide-react';
import { fetchAPI } from '../../../lib/api';

export default function PrintMasterDraw() {
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
  const printTimestamp = new Date().toLocaleString();

  useEffect(() => {
    document.title = `Master Draw - ${currentRound?.name || 'Round'} - ${currentTourney?.name || slug}`;
  }, [currentRound, currentTourney, slug]);

  return (
    <div style={{ minHeight: '100vh', background: '#fff', color: '#000', padding: '1.5rem', fontFamily: 'system-ui, -apple-system, sans-serif' }}>
      {/* Print Controls (Hidden when printing) */}
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
          <span style={{ fontWeight: 600, fontSize: '0.9rem' }}>Master Draw Print View</span>
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
          <Printer size={16} /> Print Master Draw
        </button>
      </div>

      {/* Printable Sheet Header */}
      <div style={{ borderBottom: '2px solid #000', paddingBottom: '0.75rem', marginBottom: '1.25rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div>
            <h1 style={{ fontSize: '1.6rem', fontWeight: 800, margin: '0 0 0.25rem 0', textTransform: 'uppercase', letterSpacing: '-0.02em' }}>
              {currentTourney?.name || slug}
            </h1>
            <h2 style={{ fontSize: '1.2rem', fontWeight: 600, margin: 0, color: '#3f3f46' }}>
              {currentRound ? `${currentRound.name} • Master Draw` : 'Round Master Draw'}
            </h2>
          </div>
          <div style={{ textAlign: 'right', fontSize: '0.8rem', color: '#71717a' }}>
            <div>Printed: {printTimestamp}</div>
            <div>Total Debates: {draw.length}</div>
          </div>
        </div>
      </div>

      {isLoading ? (
        <div style={{ textAlign: 'center', padding: '3rem 0', color: '#71717a' }}>Loading draw data...</div>
      ) : draw.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '3rem 0', color: '#71717a' }}>No draw pairings generated for this round.</div>
      ) : (
        <table style={{
          width: '100%',
          borderCollapse: 'collapse',
          fontSize: '0.85rem',
          textAlign: 'left'
        }}>
          <thead>
            <tr style={{ background: '#f4f4f5', borderBottom: '2px solid #000' }}>
              <th style={{ padding: '8px 10px', border: '1px solid #d4d4d8', width: '12%', fontWeight: 700 }}>Venue</th>
              <th style={{ padding: '8px 10px', border: '1px solid #d4d4d8', width: '53%', fontWeight: 700 }}>Teams & Sides</th>
              <th style={{ padding: '8px 10px', border: '1px solid #d4d4d8', width: '35%', fontWeight: 700 }}>Adjudicators</th>
            </tr>
          </thead>
          <tbody>
            {draw.map((d: any, idx: number) => {
              const chairs = d.adjudicators?.filter((a: any) => a.role === 'chair') || [];
              const panelists = d.adjudicators?.filter((a: any) => a.role === 'panel') || [];
              const trainees = d.adjudicators?.filter((a: any) => a.role === 'trainee') || [];

              return (
                <tr key={d.id} style={{
                  background: idx % 2 === 0 ? '#fff' : '#fafafa',
                  pageBreakInside: 'avoid',
                  borderBottom: '1px solid #d4d4d8'
                }}>
                  {/* Venue */}
                  <td style={{ padding: '8px 10px', border: '1px solid #d4d4d8', verticalAlign: 'top', fontWeight: 600 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.3rem' }}>
                      <span>{d.venue}</span>
                      {d.venue_accessible && (
                        <span title="Wheelchair Accessible" style={{ display: 'inline-flex' }}>
                          <Accessibility size={13} style={{ color: '#2563eb' }} />
                        </span>
                      )}
                    </div>
                  </td>

                  {/* Teams */}
                  <td style={{ padding: '8px 10px', border: '1px solid #d4d4d8', verticalAlign: 'top' }}>
                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '6px' }}>
                      {d.teams?.map((t: any) => (
                        <div key={t.team_id} style={{
                          display: 'flex',
                          justifyContent: 'space-between',
                          background: '#f4f4f5',
                          padding: '4px 6px',
                          borderRadius: '4px',
                          border: '1px solid #e4e4e7'
                        }}>
                          <span style={{ fontWeight: 600 }}>{t.team_name}</span>
                          <span style={{
                            fontWeight: 700,
                            textTransform: 'uppercase',
                            fontSize: '0.75rem',
                            color: '#52525b',
                            marginLeft: '6px'
                          }}>
                            {t.side} {t.pull_up ? '(PU)' : ''}
                          </span>
                        </div>
                      ))}
                    </div>
                  </td>

                  {/* Adjudicators */}
                  <td style={{ padding: '8px 10px', border: '1px solid #d4d4d8', verticalAlign: 'top' }}>
                    <div>
                      {chairs.length > 0 && (
                        <div style={{ marginBottom: '2px' }}>
                          <span style={{ fontWeight: 700, fontSize: '0.8rem', color: '#18181b' }}>Ⓒ </span>
                          <span style={{ fontWeight: 600 }}>{chairs.map((c: any) => c.adjudicator_name).join(', ')}</span>
                        </div>
                      )}
                      {panelists.length > 0 && (
                        <div style={{ fontSize: '0.8rem', color: '#3f3f46', marginBottom: '2px' }}>
                          <span style={{ fontWeight: 600 }}>Panel: </span>
                          {panelists.map((p: any) => p.adjudicator_name).join(', ')}
                        </div>
                      )}
                      {trainees.length > 0 && (
                        <div style={{ fontSize: '0.75rem', color: '#71717a' }}>
                          <span style={{ fontStyle: 'italic' }}>Trainee: </span>
                          {trainees.map((tr: any) => tr.adjudicator_name).join(', ')}
                        </div>
                      )}
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
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
          table {
            border-collapse: collapse !important;
          }
          tr {
            page-break-inside: avoid !important;
          }
          @page {
            size: landscape;
            margin: 1cm;
          }
        }
      `}</style>
    </div>
  );
}
