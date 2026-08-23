import { Link, useOutletContext } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Users, Gavel, Building2, CalendarDays, ArrowRight } from 'lucide-react';
import { fetchAPI, type AdminContext } from '../../lib/api';

function roundStatus(round: any): 'green' | 'orange' | 'gray' {
  if (round.draw_released && round.results_released) return 'green';
  if (round.draw_released) return 'orange';
  return 'gray';
}

export default function AdminOverview() {
  const { slug } = useOutletContext<AdminContext>();

  const { data: rounds = [] } = useQuery<any[]>({
    queryKey: ['rounds', slug],
    queryFn: () => fetchAPI(`/api/t/${slug}/rounds`),
  });
  const { data: teams = [] } = useQuery<any[]>({
    queryKey: ['teams', slug],
    queryFn: () => fetchAPI(`/api/t/${slug}/teams`),
  });
  const { data: adjudicators = [] } = useQuery<any[]>({
    queryKey: ['adjudicators', slug],
    queryFn: () => fetchAPI(`/api/t/${slug}/adjudicators`),
  });
  const { data: institutions = [] } = useQuery<any[]>({
    queryKey: ['institutions', slug],
    queryFn: () => fetchAPI(`/api/t/${slug}/institutions`),
  });

  const sortedRounds = [...rounds].sort((a, b) => a.seq - b.seq);
  const stats = [
    { label: 'Teams', value: teams.length, icon: Users },
    { label: 'Adjudicators', value: adjudicators.length, icon: Gavel },
    { label: 'Institutions', value: institutions.length, icon: Building2 },
    { label: 'Rounds', value: rounds.length, icon: CalendarDays },
  ];

  return (
    <div>
      <h2>Overview</h2>

      <div className="grid grid-cols-3" style={{ marginBottom: '2rem' }}>
        {stats.map(s => (
          <div key={s.label} className="card stat-card">
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.75rem', textTransform: 'uppercase', letterSpacing: '0.05em', color: 'var(--text-mute)', marginBottom: '0.5rem' }}>
              <s.icon size={14} /> {s.label}
            </div>
            <div style={{ fontSize: '2rem', fontWeight: '700', color: 'var(--accent)' }}>{s.value}</div>
          </div>
        ))}
      </div>

      <div className="card">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
          <h3 style={{ margin: 0 }}>Rounds</h3>
          <Link to={`/t/${slug}/admin/setup`} className="btn btn-secondary" style={{ padding: '0.35rem 0.7rem', fontSize: '0.8rem' }}>
            Manage Rounds
          </Link>
        </div>
        {sortedRounds.length === 0 ? (
          <div style={{ padding: '2rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>
            No rounds yet. Create them in Setup.
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', marginTop: '1rem' }}>
            {sortedRounds.map(r => (
              <Link
                key={r.id}
                to={`/t/${slug}/admin/rounds/${r.id}/draw`}
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  padding: '0.85rem 0.5rem',
                  borderBottom: '1px solid var(--border)',
                  color: 'var(--text-h)',
                }}
              >
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.6rem' }}>
                  <span className={`status-dot ${roundStatus(r)}`} title={
                    roundStatus(r) === 'green' ? 'Draw and results released'
                      : roundStatus(r) === 'orange' ? 'Draw released, results pending'
                        : 'Not started'
                  } />
                  <span style={{ fontWeight: 500 }}>{r.name}</span>
                  <span className={`badge ${r.stage === 'elimination' ? 'badge-warning' : 'badge-info'}`} style={{ textTransform: 'capitalize', fontSize: '0.65rem' }}>
                    {r.stage}
                  </span>
                </span>
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.4rem', fontSize: '0.8rem', color: 'var(--text-mute)' }}>
                  Seq {r.seq} <ArrowRight size={13} />
                </span>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
