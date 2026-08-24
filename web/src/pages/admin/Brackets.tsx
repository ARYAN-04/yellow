import { useState } from 'react';
import { useParams, useOutletContext } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Network, Play, SkipForward } from 'lucide-react';
import { fetchAPI, type AdminContext } from '../../lib/api';

const baseTabs = [
  { key: 'open', label: 'Open' },
  { key: 'novice', label: 'Novice' },
  { key: 'esl', label: 'ESL' },
  { key: 'efl', label: 'EFL' },
];

function nextPow2(n: number) {
  let p = 1;
  while (p < n) p *= 2;
  return p;
}

function roundName(teamCount: number) {
  return ({ 2: 'Final', 4: 'Semifinals', 8: 'Quarterfinals', 16: 'Octofinals' } as Record<number, string>)[teamCount] ?? `Elimination ${teamCount}`;
}

function DebateCard({ debate }: { debate: any }) {
  return (
    <div
      style={{
        border: '1px solid var(--border)',
        borderRadius: '8px',
        background: '#fff',
        padding: '0.6rem 0.75rem',
        minWidth: '180px',
        boxShadow: '0 1px 2px rgba(0,0,0,0.04)',
      }}
    >
      <div style={{ fontSize: '0.7rem', color: 'var(--text-mute)', marginBottom: '0.4rem' }}>
        {debate.venue || (debate.bye ? 'Bye' : `Debate ${debate.bracket_position ?? ''}`)}
      </div>
      {(debate.teams.length ? debate.teams : [{ team_id: 'none', team_name: '(no team)', side: '' }]).map((t: any) => (
        <div
          key={t.team_id}
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            gap: '0.5rem',
            padding: '0.25rem 0.4rem',
            borderRadius: '6px',
            fontSize: '0.82rem',
            background: debate.winner_team_id && debate.winner_team_id === t.team_id ? 'rgba(22,163,74,0.12)' : 'transparent',
            fontWeight: debate.winner_team_id === t.team_id ? 600 : 400,
            color: debate.winner_team_id === t.team_id ? 'var(--accent)' : 'var(--text-h)',
          }}
        >
          <span>{t.team_name}</span>
          {t.side && <span className="badge badge-info" style={{ fontSize: '0.6rem' }}>{t.side}</span>}
        </div>
      ))}
      {!debate.winner_team_id && !debate.bye && debate.teams.length > 0 && (
        <div style={{ fontSize: '0.68rem', color: 'var(--text-mute)', marginTop: '0.35rem' }}>Result pending</div>
      )}
    </div>
  );
}

export default function Brackets() {
  const { slug } = useParams<{ slug: string }>();
  const { isReadOnly } = useOutletContext<AdminContext>();
  const queryClient = useQueryClient();
  const [category, setCategory] = useState('open');

  const { data: categories = [] } = useQuery<any[]>({
    queryKey: ['break-categories', slug],
    queryFn: () => fetchAPI(`/api/t/${slug}/break-categories`),
  });

  const { data: rounds = [], isLoading } = useQuery<any[]>({
    queryKey: ['bracket', slug],
    queryFn: () => fetchAPI(`/api/t/${slug}/bracket`),
  });

  const generate = useMutation({
    mutationFn: async () => {
      const breakResult = await fetchAPI(`/api/t/${slug}/breaks/${category}`);
      const teamCount = Math.max(2, nextPow2(breakResult.qualifiers?.length ?? 0));
      const existing = await fetchAPI(`/api/t/${slug}/rounds`);
      const maxSeq = Math.max(0, ...existing.map((r: any) => r.seq));
      const created = await fetchAPI(`/api/t/${slug}/rounds`, 'POST', {
        name: roundName(teamCount),
        seq: maxSeq + 1,
        stage: 'elimination',
      });
      await fetchAPI(`/api/t/${slug}/rounds/${created.id}/generate-bracket`, 'POST', { category_id: category });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['bracket', slug] });
      queryClient.invalidateQueries({ queryKey: ['rounds', slug] });
    },
    onError: (err: any) => alert('Failed to generate bracket: ' + err.message),
  });

  const advance = useMutation({
    mutationFn: (roundId: string) => fetchAPI(`/api/t/${slug}/rounds/${roundId}/advance`, 'POST'),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['bracket', slug] });
      queryClient.invalidateQueries({ queryKey: ['rounds', slug] });
    },
    onError: (err: any) => alert('Failed to advance: ' + err.message),
  });

  const lastRound = rounds.length ? rounds[rounds.length - 1] : null;
  const allDecided = (round: any) =>
    round.debates.length > 0 && round.debates.every((d: any) => d.winner_team_id);

  const tabs = [...baseTabs, ...categories.map((c: any) => ({ key: c.id, label: c.name }))];
  const columns = [...rounds].sort((a: any, b: any) => a.seq - b.seq);

  return (
    <div>
      <h2 style={{ display: 'inline-flex', alignItems: 'center', gap: '0.5rem', margin: 0 }}>
        <Network size={20} /> Brackets
      </h2>

      {!isReadOnly && (
        <div className="card" style={{ maxWidth: '640px', margin: '1rem 0' }}>
          <h3>Generate Bracket</h3>
          <p style={{ color: 'var(--text-mute)', fontSize: '0.85rem' }}>
            Seeds an elimination round from the current break standings (published snapshot takes precedence).
          </p>
          <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center', flexWrap: 'wrap' }}>
            <select className="input select" style={{ width: '180px' }} value={category} onChange={e => setCategory(e.target.value)} aria-label="Break category">
              {tabs.map(t => <option key={t.key} value={t.key}>{t.label}</option>)}
            </select>
            <button className="btn btn-primary" onClick={() => generate.mutate()} disabled={generate.isPending}>
              <Play size={15} /> Generate Bracket
            </button>
          </div>
        </div>
      )}

      {isLoading ? (
        <div style={{ padding: '3rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>Loading bracket...</div>
      ) : rounds.length === 0 ? (
        <div className="card" style={{ padding: '3rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>
          No elimination rounds yet — generate a bracket to start the knockouts.
        </div>
      ) : (
        <div style={{ display: 'flex', gap: '1.75rem', overflowX: 'auto', alignItems: 'stretch', paddingBottom: '1rem' }}>
          {columns.map(round => (
            <div key={round.id} style={{ display: 'flex', flexDirection: 'column', minWidth: '210px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
                <span style={{ fontWeight: 600, fontSize: '0.9rem', color: 'var(--text-h)' }}>{round.name}</span>
                {!isReadOnly && round.id === lastRound.id && allDecided(round) && (
                  <button
                    className="btn btn-secondary"
                    style={{ padding: '0.25rem 0.6rem', fontSize: '0.72rem' }}
                    onClick={() => advance.mutate(round.id)}
                    disabled={advance.isPending}
                  >
                    <SkipForward size={12} /> Advance
                  </button>
                )}
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', justifyContent: 'space-around', flex: 1 }}>
                {round.debates.map((d: any) => <DebateCard key={d.id} debate={d} />)}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
