import { NavLink, useParams } from 'react-router-dom';
import { TrendingUp, FileText, Shuffle, Megaphone, Home } from 'lucide-react';

export default function PublicNav({ title }: { title?: string }) {
  const { slug } = useParams<{ slug: string }>();

  const links = [
    { to: `/t/${slug}/standings`, label: 'Standings', icon: TrendingUp },
    { to: `/t/${slug}/results`, label: 'Results', icon: FileText },
    { to: `/t/${slug}/draw`, label: 'Draw', icon: Shuffle },
    { to: `/t/${slug}/motions`, label: 'Motions', icon: Megaphone },
  ];

  return (
    <header
      style={{
        background: '#fff',
        borderBottom: '1px solid var(--border)',
        padding: '0.75rem 1.5rem',
        marginBottom: '1.5rem',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        flexWrap: 'wrap',
        gap: '1rem',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
        <NavLink
          to="/"
          style={{
            fontWeight: 800,
            fontSize: '1.1rem',
            color: 'var(--text-h)',
            textDecoration: 'none',
            display: 'inline-flex',
            alignItems: 'center',
            gap: '0.35rem',
          }}
        >
          Yellow
        </NavLink>
        <span style={{ color: 'var(--border)' }}>/</span>
        <span style={{ fontSize: '0.85rem', fontWeight: 600, color: 'var(--accent)' }}>
          {slug}
        </span>
        {title && (
          <>
            <span style={{ color: 'var(--border)' }}>•</span>
            <span style={{ fontSize: '0.85rem', color: 'var(--text-mute)' }}>{title}</span>
          </>
        )}
      </div>

      <nav style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap', alignItems: 'center' }}>
        {links.map(l => (
          <NavLink
            key={l.to}
            to={l.to}
            className={({ isActive }) => `btn ${isActive ? 'btn-primary' : 'btn-secondary'}`}
            style={{
              fontSize: '0.8rem',
              padding: '0.35rem 0.75rem',
              display: 'inline-flex',
              alignItems: 'center',
              gap: '0.35rem',
            }}
          >
            <l.icon size={14} /> {l.label}
          </NavLink>
        ))}
        <NavLink
          to={`/t/${slug}/admin`}
          className="btn btn-secondary"
          style={{
            fontSize: '0.8rem',
            padding: '0.35rem 0.75rem',
            display: 'inline-flex',
            alignItems: 'center',
            gap: '0.35rem',
          }}
        >
          <Home size={14} /> Admin Portal
        </NavLink>
      </nav>
    </header>
  );
}
