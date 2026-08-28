import { NavLink, useParams, Outlet, useOutletContext } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { CheckCircle2, Shuffle, FileText, Megaphone } from 'lucide-react';
import { fetchAPI, type AdminContext, type RoundContext } from '../../../lib/api';

const subTabs = [
  { to: 'availability', label: 'Availability', icon: CheckCircle2 },
  { to: 'draw', label: 'Draw', icon: Shuffle },
  { to: 'results', label: 'Results', icon: FileText },
  { to: 'motions', label: 'Motions', icon: Megaphone },
];

export default function RoundLayout() {
  const { slug, roundId } = useParams<{ slug: string; roundId: string }>();
  const { isReadOnly, role, isAssistant } = useOutletContext<AdminContext>();

  const { data: rounds = [] } = useQuery<any[]>({
    queryKey: ['rounds', slug],
    queryFn: () => fetchAPI(`/api/t/${slug}/rounds`),
  });

  const round = rounds.find(r => r.id === roundId) || null;

  return (
    <div>
      <div style={{ display: 'flex', alignItems: 'baseline', gap: '0.75rem', marginBottom: '1rem' }}>
        <h2 style={{ margin: 0 }}>{round ? `Round ${round.seq}: ${round.name}` : 'Round'}</h2>
        {round && (
          <span className={`badge ${round.stage === 'elimination' ? 'badge-warning' : 'badge-info'}`} style={{ textTransform: 'capitalize' }}>
            {round.stage}
          </span>
        )}
      </div>

      <div className="tabs">
        {subTabs.map(tab => (
          <NavLink
            key={tab.to}
            to={`/t/${slug}/admin/rounds/${roundId}/${tab.to}`}
            className={({ isActive }) => `tab-btn ${isActive ? 'active' : ''}`}
          >
            <tab.icon size={15} /> {tab.label}
          </NavLink>
        ))}
      </div>

      <Outlet context={{ slug: slug ?? '', isReadOnly, role, isAssistant, roundId: roundId ?? '', round } satisfies RoundContext} />
    </div>
  );
}
