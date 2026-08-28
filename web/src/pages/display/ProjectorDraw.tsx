import { useState, useEffect, useRef } from 'react';
import { useParams, useSearchParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Maximize, Minimize, RefreshCw, Moon, Sun, ArrowLeft, Accessibility, Play, Pause } from 'lucide-react';
import { fetchAPI } from '../../lib/api';

export default function ProjectorDraw() {
  const { slug } = useParams<{ slug: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();

  const roundIdParam = searchParams.get('round_id');
  const [theme, setTheme] = useState<'dark' | 'light'>('dark');
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [controlsVisible, setControlsVisible] = useState(true);
  const [autoScroll, setAutoScroll] = useState(false);
  const timeoutRef = useRef<any>(null);
  const scrollIntervalRef = useRef<any>(null);

  // Load rounds
  const { data: rounds = [] } = useQuery<any[]>({
    queryKey: ['rounds', slug],
    queryFn: () => fetchAPI(`/api/t/${slug}/rounds`),
  });

  const sortedRounds = [...rounds].sort((a, b) => b.seq - a.seq);
  const activeRoundId = roundIdParam || sortedRounds[0]?.id;
  const currentRound = rounds.find(r => r.id === activeRoundId);

  // Load draw with polling
  const { data: draw = [], isLoading, isFetching, refetch } = useQuery<any[]>({
    queryKey: ['draw', slug, activeRoundId],
    queryFn: () => fetchAPI(`/api/t/${slug}/rounds/${activeRoundId}/draw`),
    enabled: !!activeRoundId,
    refetchInterval: 15000,
  });

  // Auto-scroll loop
  useEffect(() => {
    if (!autoScroll) {
      if (scrollIntervalRef.current) clearInterval(scrollIntervalRef.current);
      return;
    }

    let direction = 1;
    scrollIntervalRef.current = setInterval(() => {
      const scrollY = window.scrollY;
      const maxScroll = document.documentElement.scrollHeight - window.innerHeight;

      if (scrollY >= maxScroll - 5) {
        // Pause at bottom then jump to top
        setTimeout(() => {
          window.scrollTo({ top: 0, behavior: 'smooth' });
        }, 1500);
      } else {
        window.scrollBy({ top: 2 * direction, behavior: 'auto' });
      }
    }, 50);

    return () => {
      if (scrollIntervalRef.current) clearInterval(scrollIntervalRef.current);
    };
  }, [autoScroll]);

  // Handle Fullscreen
  const toggleFullscreen = () => {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen().then(() => setIsFullscreen(true)).catch(() => {});
    } else {
      document.exitFullscreen().then(() => setIsFullscreen(false)).catch(() => {});
    }
  };

  useEffect(() => {
    const handleFSChange = () => {
      setIsFullscreen(!!document.fullscreenElement);
    };
    document.addEventListener('fullscreenchange', handleFSChange);
    return () => document.removeEventListener('fullscreenchange', handleFSChange);
  }, []);

  // Idle controls hide
  const handleMouseMove = () => {
    setControlsVisible(true);
    if (timeoutRef.current) clearTimeout(timeoutRef.current);
    timeoutRef.current = setTimeout(() => {
      setControlsVisible(false);
    }, 3500);
  };

  useEffect(() => {
    window.addEventListener('mousemove', handleMouseMove);
    return () => {
      window.removeEventListener('mousemove', handleMouseMove);
      if (timeoutRef.current) clearTimeout(timeoutRef.current);
    };
  }, []);

  const isDark = theme === 'dark';
  const bg = isDark ? '#09090b' : '#ffffff';
  const textPrimary = isDark ? '#f8fafc' : '#0f172a';
  const textMute = isDark ? '#94a3b8' : '#64748b';
  const cardBg = isDark ? '#18181b' : '#f8fafc';
  const cardBorder = isDark ? '#27272a' : '#e2e8f0';

  return (
    <div
      onMouseMove={handleMouseMove}
      style={{
        minHeight: '100vh',
        background: bg,
        color: textPrimary,
        fontFamily: 'system-ui, -apple-system, sans-serif',
        position: 'relative',
        transition: 'background-color 0.2s, color 0.2s',
        paddingBottom: '4rem'
      }}
    >
      {/* Floating Control Toolbar */}
      <div
        style={{
          position: 'fixed',
          top: '1rem',
          left: '50%',
          transform: 'translateX(-50%)',
          zIndex: 100,
          background: isDark ? 'rgba(24, 24, 27, 0.85)' : 'rgba(255, 255, 255, 0.85)',
          backdropFilter: 'blur(10px)',
          border: `1px solid ${cardBorder}`,
          borderRadius: '9999px',
          padding: '0.4rem 1rem',
          display: 'flex',
          alignItems: 'center',
          gap: '0.75rem',
          opacity: controlsVisible ? 1 : 0,
          pointerEvents: controlsVisible ? 'auto' : 'none',
          transition: 'opacity 0.3s ease',
          boxShadow: '0 10px 25px -5px rgba(0, 0, 0, 0.2)'
        }}
      >
        <button
          onClick={() => navigate(-1)}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: '0.3rem',
            background: 'transparent',
            border: 'none',
            color: textMute,
            cursor: 'pointer',
            fontSize: '0.85rem'
          }}
          title="Exit Projector Mode"
        >
          <ArrowLeft size={16} /> Exit
        </button>

        <div style={{ height: '16px', width: '1px', background: cardBorder }} />

        {/* Round Picker */}
        <select
          value={activeRoundId || ''}
          onChange={e => setSearchParams({ round_id: e.target.value })}
          style={{
            background: isDark ? '#27272a' : '#f1f5f9',
            color: textPrimary,
            border: `1px solid ${cardBorder}`,
            borderRadius: '6px',
            padding: '0.3rem 0.6rem',
            fontSize: '0.85rem',
            cursor: 'pointer'
          }}
        >
          {sortedRounds.map(r => (
            <option key={r.id} value={r.id}>{r.name}</option>
          ))}
        </select>

        <div style={{ height: '16px', width: '1px', background: cardBorder }} />

        {/* Auto Scroll Toggle */}
        <button
          onClick={() => setAutoScroll(s => !s)}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: '0.3rem',
            background: autoScroll ? (isDark ? '#0284c7' : '#0284c7') : 'transparent',
            color: autoScroll ? '#fff' : textMute,
            border: 'none',
            padding: '0.2rem 0.6rem',
            borderRadius: '4px',
            cursor: 'pointer',
            fontSize: '0.8rem',
            fontWeight: 600
          }}
          title="Toggle Auto-Scroll"
        >
          {autoScroll ? <Pause size={14} /> : <Play size={14} />} Auto-Scroll
        </button>

        {/* Refresh button */}
        <button
          onClick={() => refetch()}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: '0.3rem',
            background: 'transparent',
            border: 'none',
            color: isFetching ? '#38bdf8' : textMute,
            cursor: 'pointer',
            fontSize: '0.85rem'
          }}
          title="Refresh Draw"
        >
          <RefreshCw size={15} className={isFetching ? 'spin' : ''} />
        </button>

        {/* Theme Toggle */}
        <button
          onClick={() => setTheme(t => t === 'dark' ? 'light' : 'dark')}
          style={{
            background: 'transparent',
            border: 'none',
            color: textMute,
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center'
          }}
          title="Toggle Dark/Light Mode"
        >
          {isDark ? <Sun size={16} /> : <Moon size={16} />}
        </button>

        {/* Fullscreen Toggle */}
        <button
          onClick={toggleFullscreen}
          style={{
            background: 'transparent',
            border: 'none',
            color: textMute,
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center'
          }}
          title="Toggle Fullscreen"
        >
          {isFullscreen ? <Minimize size={16} /> : <Maximize size={16} />}
        </button>
      </div>

      {/* Main Content */}
      <div style={{
        maxWidth: '1400px',
        margin: '0 auto',
        padding: '5rem 2rem 2rem 2rem'
      }}>
        {/* Header Title */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end', borderBottom: `2px solid ${cardBorder}`, paddingBottom: '1rem', marginBottom: '2rem' }}>
          <div>
            <div style={{ fontSize: '0.85rem', fontWeight: 800, textTransform: 'uppercase', letterSpacing: '0.1em', color: isDark ? '#38bdf8' : '#0284c7', marginBottom: '0.25rem' }}>
              {slug}
            </div>
            <h1 style={{ fontSize: '2.2rem', fontWeight: 800, margin: 0, letterSpacing: '-0.02em' }}>
              {currentRound?.name || 'Round'} Pairings & Allocations
            </h1>
          </div>
          <div style={{ textAlign: 'right', color: textMute, fontSize: '0.9rem' }}>
            <div>Total Debates: <strong>{draw.length}</strong></div>
          </div>
        </div>

        {isLoading ? (
          <div style={{ textAlign: 'center', padding: '5rem 0', color: textMute, fontSize: '1.4rem' }}>
            Loading draw pairings...
          </div>
        ) : draw.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '5rem 0', color: textMute, fontSize: '1.4rem' }}>
            No draw pairings released for this round yet.
          </div>
        ) : (
          <div style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fill, minmax(420px, 1fr))',
            gap: '1.5rem'
          }}>
            {draw.map((d: any) => {
              const chairs = d.adjudicators?.filter((a: any) => a.role === 'chair') || [];
              const panelists = d.adjudicators?.filter((a: any) => a.role === 'panel') || [];
              const trainees = d.adjudicators?.filter((a: any) => a.role === 'trainee') || [];

              return (
                <div
                  key={d.id}
                  style={{
                    background: cardBg,
                    border: `1px solid ${cardBorder}`,
                    borderRadius: '12px',
                    padding: '1.25rem',
                    display: 'flex',
                    flexDirection: 'column',
                    justifyContent: 'space-between',
                    boxShadow: '0 4px 12px rgba(0, 0, 0, 0.08)'
                  }}
                >
                  {/* Venue Bar */}
                  <div style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    borderBottom: `1px solid ${cardBorder}`,
                    paddingBottom: '0.75rem',
                    marginBottom: '1rem'
                  }}>
                    <span style={{ fontSize: '1.2rem', fontWeight: 800, color: textPrimary, display: 'inline-flex', alignItems: 'center', gap: '0.4rem' }}>
                      {d.venue}
                      {d.venue_accessible && (
                        <span title="Wheelchair Accessible" style={{ display: 'inline-flex' }}>
                          <Accessibility size={16} style={{ color: '#38bdf8' }} />
                        </span>
                      )}
                    </span>
                  </div>

                  {/* Teams Grid */}
                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.75rem', marginBottom: '1rem' }}>
                    {d.teams?.map((t: any) => (
                      <div
                        key={t.team_id}
                        style={{
                          background: isDark ? '#27272a' : '#f1f5f9',
                          padding: '0.75rem',
                          borderRadius: '8px',
                          display: 'flex',
                          flexDirection: 'column',
                          gap: '0.25rem'
                        }}
                      >
                        <span style={{
                          fontSize: '0.75rem',
                          fontWeight: 800,
                          textTransform: 'uppercase',
                          letterSpacing: '0.05em',
                          color: isDark ? '#38bdf8' : '#0284c7'
                        }}>
                          {t.side} {t.pull_up ? '(PU)' : ''}
                        </span>
                        <span style={{ fontSize: '1rem', fontWeight: 700, color: textPrimary }}>
                          {t.team_name}
                        </span>
                      </div>
                    ))}
                  </div>

                  {/* Adjudicators */}
                  <div style={{
                    borderTop: `1px solid ${cardBorder}`,
                    paddingTop: '0.75rem',
                    fontSize: '0.9rem',
                    color: textMute
                  }}>
                    {chairs.length > 0 && (
                      <div style={{ color: textPrimary, fontWeight: 600, marginBottom: '0.2rem' }}>
                        <span style={{ color: '#eab308', fontWeight: 800 }}>Ⓒ Chair: </span>
                        {chairs.map((c: any) => c.adjudicator_name).join(', ')}
                      </div>
                    )}
                    {panelists.length > 0 && (
                      <div style={{ fontSize: '0.85rem', marginBottom: '0.2rem' }}>
                        <strong>Panel: </strong>
                        {panelists.map((p: any) => p.adjudicator_name).join(', ')}
                      </div>
                    )}
                    {trainees.length > 0 && (
                      <div style={{ fontSize: '0.8rem', color: textMute, fontStyle: 'italic' }}>
                        <span>Trainee: </span>
                        {trainees.map((tr: any) => tr.adjudicator_name).join(', ')}
                      </div>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      <style>{`
        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
        .spin {
          animation: spin 1s linear infinite;
        }
      `}</style>
    </div>
  );
}
