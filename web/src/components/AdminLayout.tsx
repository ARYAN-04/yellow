import { useState, useEffect } from 'react';
import { NavLink, useParams, Outlet, useLocation } from 'react-router-dom';
import {
  LayoutDashboard,
  MessageSquare,
  TrendingUp,
  Trophy,
  Settings,
  Ban,
  QrCode,
  ChevronDown,
  CalendarDays,
  CheckCircle2,
  Shuffle,
  FileText,
  Megaphone,
  Network,
} from 'lucide-react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import LoginPage from '../pages/LoginPage';
import { fetchAPI, type AdminContext } from '../lib/api';

const tournamentLinks = [
  { to: '', label: 'Overview', icon: LayoutDashboard, end: true },
  { to: 'feedback', label: 'Feedback', icon: MessageSquare },
  { to: 'standings', label: 'Standings', icon: TrendingUp },
  { to: 'breaks', label: 'Breaks', icon: Trophy },
  { to: 'brackets', label: 'Brackets', icon: Network },
  { to: 'setup', label: 'Setup', icon: Settings },
  { to: 'conflicts', label: 'Conflicts', icon: Ban },
  { to: 'checkins', label: 'Check-ins', icon: QrCode },
];

const roundSubLinks = [
  { to: 'availability', label: 'Availability', icon: CheckCircle2 },
  { to: 'draw', label: 'Draw', icon: Shuffle },
  { to: 'results', label: 'Results', icon: FileText },
  { to: 'motions', label: 'Motions', icon: Megaphone },
];

function roundStatus(round: any): 'green' | 'orange' | 'gray' {
  if (round.draw_released && round.results_released) return 'green';
  if (round.draw_released) return 'orange';
  return 'gray';
}

function Sidebar() {
  const { slug } = useParams<{ slug: string }>();
  const location = useLocation();
  const [expanded, setExpanded] = useState<Record<string, boolean>>({});

  const { data: rounds = [] } = useQuery<any[]>({
    queryKey: ['rounds', slug],
    queryFn: () => fetchAPI(`/api/t/${slug}/rounds`),
  });

  const activeRoundId = location.pathname.match(/^\/t\/[^/]+\/admin\/rounds\/([^/]+)/)?.[1];

  useEffect(() => {
    if (activeRoundId && !expanded[activeRoundId]) {
      setExpanded(prev => ({ ...prev, [activeRoundId]: true }));
    }
  }, [activeRoundId]);

  const toggle = (id: string) => setExpanded(prev => ({ ...prev, [id]: !prev[id] }));

  const sortedRounds = [...rounds].sort((a, b) => a.seq - b.seq);

  return (
    <nav className="admin-sidebar">
      <NavLink to="/" className="sidebar-brand">
        Yellow
      </NavLink>

      <div className="sidebar-section-label">Tournament</div>
      <div className="sidebar-group">
        {tournamentLinks.map(link => (
          <NavLink
            key={link.to}
            to={`/t/${slug}/admin${link.to ? `/${link.to}` : ''}`}
            end={link.end}
            className={({ isActive }) => `sidebar-item ${isActive ? 'active' : ''}`}
          >
            <link.icon size={15} />
            <span>{link.label}</span>
          </NavLink>
        ))}
      </div>

      <div className="sidebar-section-label">Rounds</div>
      <div className="sidebar-group">
        {sortedRounds.length === 0 && (
          <div className="sidebar-empty">No rounds created yet</div>
        )}
        {sortedRounds.map(r => {
          const isOpen = !!expanded[r.id];
          const isActiveRound = activeRoundId === r.id;
          return (
            <div key={r.id} className={`sidebar-round-group ${isActiveRound ? 'active-round' : ''}`}>
              <button
                type="button"
                className={`sidebar-item sidebar-round-header ${isActiveRound ? 'active' : ''}`}
                onClick={() => toggle(r.id)}
              >
                <CalendarDays size={15} />
                <span>{r.name}</span>
                <span className={`status-dot ${roundStatus(r)}`} title={
                  roundStatus(r) === 'green' ? 'Draw and results released'
                    : roundStatus(r) === 'orange' ? 'Draw released, results pending'
                      : 'Not started'
                } />
                <ChevronDown size={13} className={`sidebar-chevron ${isOpen ? 'open' : ''}`} />
              </button>
              {isOpen && (
                <div className="sidebar-subgroup">
                  {roundSubLinks.map(sub => (
                    <NavLink
                      key={sub.to}
                      to={`/t/${slug}/admin/rounds/${r.id}/${sub.to}`}
                      className={({ isActive }) => `sidebar-item sidebar-subitem ${isActive ? 'active' : ''}`}
                    >
                      <sub.icon size={14} />
                      <span>{sub.label}</span>
                    </NavLink>
                  ))}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </nav>
  );
}

export default function AdminLayout() {
  const { slug } = useParams<{ slug: string }>();
  const queryClient = useQueryClient();
  const [authed, setAuthed] = useState(false);
  const [tournamentInfo, setTournamentInfo] = useState<any>(null);

  // Check login and archive status
  useEffect(() => {
    fetchAPI('/api/tournaments')
      .then(list => {
        const current = list.find((t: any) => t.slug === slug);
        setTournamentInfo(current);
        if (current && current.is_archived) {
          setAuthed(true); // Bypass login screen for archived public views
        } else {
          fetchAPI(`/api/t/${slug}/institutions`)
            .then(() => setAuthed(true))
            .catch(() => setAuthed(false));
        }
      })
      .catch(() => {
        fetchAPI(`/api/t/${slug}/institutions`)
          .then(() => setAuthed(true))
          .catch(() => setAuthed(false));
      });
  }, [slug]);

  const isReadOnly = tournamentInfo?.is_archived === 1 || tournamentInfo?.is_archived === true;

  if (!authed) {
    return (
      <LoginPage
        onSuccess={() => {
          setAuthed(true);
          queryClient.invalidateQueries({ queryKey: ['tournaments'] });
        }}
      />
    );
  }

  return (
    <div className="admin-shell">
      <Sidebar />
      <div className="admin-main">
        <header className="admin-topbar">
          <div className="logo-text" style={{ fontSize: '1rem' }}>
            Yellow {isReadOnly ? 'Record' : 'Admin'} <span style={{ color: 'var(--accent)', fontSize: '0.8rem', fontWeight: '500' }}>/{slug}</span>
          </div>
          <NavLink to="/" className="btn btn-secondary" style={{ padding: '0.4rem 0.8rem', fontSize: '0.8rem' }}>
            Exit {isReadOnly ? 'View' : ''}
          </NavLink>
        </header>
        <main className="admin-content">
          {isReadOnly && (
            <div className="badge badge-success" style={{ width: '100%', padding: '0.75rem', borderRadius: '8px', marginBottom: '1.5rem', fontSize: '0.9rem', justifyContent: 'center' }}>
              This tournament record is archived and read-only.
            </div>
          )}
          <Outlet context={{ slug: slug ?? '', isReadOnly } satisfies AdminContext} />
        </main>
      </div>
    </div>
  );
}
