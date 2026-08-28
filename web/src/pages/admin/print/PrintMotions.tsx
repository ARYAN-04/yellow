import { useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Printer, ArrowLeft, Info } from 'lucide-react';
import { fetchAPI } from '../../../lib/api';

export default function PrintMotions() {
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

  const { data: motions = [], isLoading } = useQuery<any[]>({
    queryKey: ['round-motions', slug, roundId],
    queryFn: () => fetchAPI(`/api/t/${slug}/rounds/${roundId}/motions`),
  });

  const currentTourney = tournaments.find((t: any) => t.slug === slug);
  const currentRound = rounds.find((r: any) => r.id === roundId);

  useEffect(() => {
    document.title = `Motion Slips - ${currentRound?.name || 'Round'} - ${currentTourney?.name || slug}`;
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
          <span style={{ fontWeight: 600, fontSize: '0.9rem' }}>Motion Handout & Distribution Slips</span>
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
          <Printer size={16} /> Print Motion Slips
        </button>
      </div>

      {isLoading ? (
        <div style={{ textAlign: 'center', padding: '3rem 0', color: '#71717a' }}>Loading motions...</div>
      ) : motions.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '3rem 0', color: '#71717a' }}>No motions found for this round.</div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>
          {motions.map((m: any, idx: number) => (
            <div
              key={m.id || idx}
              className="motion-slip"
              style={{
                border: '2px dashed #94a3b8',
                borderRadius: '8px',
                padding: '1.75rem',
                background: '#fafafa',
                pageBreakInside: 'avoid',
                position: 'relative'
              }}
            >
              {/* Slip Header */}
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '2px solid #000', paddingBottom: '0.75rem', marginBottom: '1.25rem' }}>
                <div>
                  <div style={{ fontSize: '0.8rem', textTransform: 'uppercase', letterSpacing: '0.05em', color: '#52525b', fontWeight: 700 }}>
                    {currentTourney?.name || slug} • Motion Release Slip
                  </div>
                  <h2 style={{ fontSize: '1.3rem', fontWeight: 800, margin: '2px 0 0 0' }}>
                    {currentRound?.name || 'Round'} {m.reference ? `• ${m.reference}` : `• Motion #${m.seq}`}
                  </h2>
                </div>
                <div style={{ textAlign: 'right', fontSize: '0.8rem', color: '#64748b' }}>
                  {m.released_at ? (
                    <div>Released: {new Date(m.released_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</div>
                  ) : (
                    <div>Official Motion Slip</div>
                  )}
                </div>
              </div>

              {/* Info Slide if present */}
              {m.info_slide && (
                <div style={{
                  background: '#f1f5f9',
                  borderLeft: '4px solid #0284c7',
                  padding: '1rem',
                  borderRadius: '0 6px 6px 0',
                  marginBottom: '1.25rem'
                }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', fontWeight: 700, fontSize: '0.85rem', color: '#0369a1', marginBottom: '0.35rem', textTransform: 'uppercase' }}>
                    <Info size={15} /> Info Slide / Context
                  </div>
                  <div style={{ fontSize: '0.95rem', lineHeight: '1.5', whiteSpace: 'pre-wrap', color: '#1e293b' }}>
                    {m.info_slide}
                  </div>
                </div>
              )}

              {/* Main Motion Text */}
              <div style={{
                background: '#fff',
                border: '1.5px solid #e2e8f0',
                borderRadius: '6px',
                padding: '1.25rem',
                marginBottom: '1rem'
              }}>
                <div style={{ fontSize: '0.8rem', textTransform: 'uppercase', fontWeight: 700, color: '#475569', marginBottom: '0.4rem' }}>
                  Motion
                </div>
                <div style={{
                  fontSize: '1.25rem',
                  fontWeight: 700,
                  lineHeight: '1.45',
                  color: '#0f172a'
                }}>
                  {m.text}
                </div>
              </div>

              {/* Cut Line Indicator */}
              <div className="no-print" style={{
                position: 'absolute',
                bottom: '-12px',
                right: '20px',
                background: '#fff',
                padding: '0 8px',
                fontSize: '0.7rem',
                color: '#94a3b8',
                fontWeight: 600
              }}>
                ✂ Cut Along Dotted Line
              </div>
            </div>
          ))}
        </div>
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
          .motion-slip {
            page-break-inside: avoid !important;
            margin-bottom: 2rem !important;
            background: #fff !important;
            border-color: #000 !important;
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
