import { useState, useEffect, useRef } from 'react';
import { useParams, useSearchParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Maximize, Minimize, RefreshCw, Moon, Sun, ArrowLeft, Info, AlertCircle } from 'lucide-react';
import { fetchAPI } from '../../lib/api';

export default function ProjectorMotion() {
  const { slug } = useParams<{ slug: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();

  const roundIdParam = searchParams.get('round_id');
  const [theme, setTheme] = useState<'dark' | 'light'>('dark');
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [controlsVisible, setControlsVisible] = useState(true);
  const timeoutRef = useRef<any>(null);

  // Load rounds
  const { data: rounds = [] } = useQuery<any[]>({
    queryKey: ['rounds', slug],
    queryFn: () => fetchAPI(`/api/t/${slug}/rounds`),
  });

  const sortedRounds = [...rounds].sort((a, b) => b.seq - a.seq);
  const activeRoundId = roundIdParam || sortedRounds[0]?.id;
  const currentRound = rounds.find(r => r.id === activeRoundId);

  // Load motions with auto-refresh interval
  const { data: motions = [], isLoading, isFetching, refetch } = useQuery<any[]>({
    queryKey: ['round-motions', slug, activeRoundId],
    queryFn: () => fetchAPI(`/api/t/${slug}/rounds/${activeRoundId}/motions`),
    enabled: !!activeRoundId,
    refetchInterval: 10000, // Poll every 10 seconds for live motion release
  });

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
        display: 'flex',
        flexDirection: 'column',
        position: 'relative',
        overflowX: 'hidden',
        transition: 'background-color 0.2s, color 0.2s'
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
          title="Refresh Motion"
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

      {/* Main Presentation Body */}
      <div style={{
        flex: 1,
        maxWidth: '1200px',
        width: '100%',
        margin: '0 auto',
        padding: '5rem 2rem 3rem 2rem',
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'center',
        boxSizing: 'border-box'
      }}>
        {/* Tournament & Round Tag */}
        <div style={{ textAlign: 'center', marginBottom: '2.5rem' }}>
          <div style={{
            display: 'inline-block',
            textTransform: 'uppercase',
            letterSpacing: '0.15em',
            fontSize: '0.9rem',
            fontWeight: 800,
            color: isDark ? '#38bdf8' : '#0284c7',
            marginBottom: '0.5rem'
          }}>
            {slug} • {currentRound?.name || 'Round'}
          </div>
          <h2 style={{
            fontSize: '1.2rem',
            fontWeight: 500,
            color: textMute,
            margin: 0
          }}>
            Auditorium Motion Presentation
          </h2>
        </div>

        {isLoading ? (
          <div style={{ textAlign: 'center', fontSize: '1.5rem', color: textMute, padding: '4rem 0' }}>
            Loading motion release...
          </div>
        ) : motions.length === 0 ? (
          <div style={{
            textAlign: 'center',
            padding: '4rem 2rem',
            background: cardBg,
            border: `1px solid ${cardBorder}`,
            borderRadius: '16px',
            maxWidth: '640px',
            margin: '0 auto'
          }}>
            <AlertCircle size={48} style={{ color: textMute, marginBottom: '1rem' }} />
            <h3 style={{ fontSize: '1.4rem', fontWeight: 700, margin: '0 0 0.5rem 0' }}>Motion Not Released Yet</h3>
            <p style={{ color: textMute, fontSize: '1rem', margin: 0 }}>
              The adjudication core has not released the motion for this round. This display will automatically update once released.
            </p>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '2.5rem' }}>
            {motions.map((m: any, idx: number) => (
              <div key={m.id || idx} style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
                {/* Motion Reference Badge if present */}
                {m.reference && (
                  <div style={{ textAlign: 'center' }}>
                    <span style={{
                      display: 'inline-block',
                      background: isDark ? 'rgba(56, 189, 248, 0.15)' : 'rgba(2, 132, 199, 0.1)',
                      color: isDark ? '#38bdf8' : '#0284c7',
                      border: `1px solid ${isDark ? 'rgba(56, 189, 248, 0.3)' : 'rgba(2, 132, 199, 0.25)'}`,
                      padding: '0.35rem 1.25rem',
                      borderRadius: '9999px',
                      fontWeight: 700,
                      fontSize: '1rem',
                      textTransform: 'uppercase',
                      letterSpacing: '0.05em'
                    }}>
                      {m.reference}
                    </span>
                  </div>
                )}

                {/* Info Slide Card */}
                {m.info_slide && (
                  <div style={{
                    background: cardBg,
                    border: `2px solid ${isDark ? '#0284c7' : '#38bdf8'}`,
                    borderRadius: '16px',
                    padding: '2rem',
                    boxShadow: '0 10px 30px -10px rgba(0, 0, 0, 0.3)'
                  }}>
                    <div style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.5rem',
                      fontSize: '1rem',
                      fontWeight: 800,
                      textTransform: 'uppercase',
                      letterSpacing: '0.08em',
                      color: isDark ? '#38bdf8' : '#0284c7',
                      marginBottom: '1rem'
                    }}>
                      <Info size={20} /> Info Slide
                    </div>
                    <div style={{
                      fontSize: '1.45rem',
                      lineHeight: '1.6',
                      whiteSpace: 'pre-wrap',
                      color: isDark ? '#e2e8f0' : '#1e293b'
                    }}>
                      {m.info_slide}
                    </div>
                  </div>
                )}

                {/* Motion Text */}
                <div style={{
                  background: cardBg,
                  border: `1px solid ${cardBorder}`,
                  borderRadius: '20px',
                  padding: '3rem 2.5rem',
                  textAlign: 'center',
                  boxShadow: '0 20px 40px -15px rgba(0, 0, 0, 0.4)'
                }}>
                  <div style={{
                    fontSize: '0.9rem',
                    textTransform: 'uppercase',
                    letterSpacing: '0.15em',
                    fontWeight: 700,
                    color: textMute,
                    marginBottom: '1.25rem'
                  }}>
                    Motion #{m.seq}
                  </div>
                  <h1 style={{
                    fontSize: '2.5rem',
                    fontWeight: 800,
                    lineHeight: '1.35',
                    margin: 0,
                    color: textPrimary,
                    letterSpacing: '-0.01em'
                  }}>
                    {m.text}
                  </h1>
                </div>
              </div>
            ))}
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
