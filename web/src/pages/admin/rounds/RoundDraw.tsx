import { useState } from 'react';
import { useOutletContext } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { MapPin } from 'lucide-react';
import { fetchAPI, type RoundContext } from '../../../lib/api';
import AllocationsBoard from './AllocationsBoard';

export default function RoundDraw() {
  const { slug, roundId, round, isReadOnly } = useOutletContext<RoundContext>();
  const queryClient = useQueryClient();
  const [view, setView] = useState<'cards' | 'board'>('cards');

  const { data: draw = [], isLoading } = useQuery<any[]>({
    queryKey: ['draw', slug, roundId],
    queryFn: () => fetchAPI(`/api/t/${slug}/rounds/${roundId}/draw`),
  });

  const generateDraw = useMutation({
    mutationFn: () => fetchAPI(`/api/t/${slug}/rounds/${roundId}/draw`, 'POST'),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['draw', slug, roundId] }),
    onError: (err: any) => alert("Failed to generate draw: " + err.message),
  });

  const updateRound = async (field: string, value: any) => {
    try {
      await fetchAPI(`/api/t/${slug}/rounds/${roundId}`, 'PUT', { [field]: value });
      queryClient.invalidateQueries({ queryKey: ['rounds', slug] });
    } catch (err: any) { alert(err.message); }
  };

  return (
    <div>
      <div className="card" style={{ marginBottom: '1.5rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
          <h3 style={{ margin: 0 }}>Round Draw</h3>
          <div style={{ display: 'inline-flex', border: '1px solid var(--border)', borderRadius: '6px', overflow: 'hidden' }}>
            {(['cards', 'board'] as const).map(v => (
              <button
                key={v}
                onClick={() => setView(v)}
                style={{
                  padding: '0.35rem 0.8rem',
                  fontSize: '0.8rem',
                  border: 'none',
                  cursor: 'pointer',
                  background: view === v ? 'var(--accent)' : '#fff',
                  color: view === v ? '#fff' : 'var(--text-mute)',
                }}
              >
                {v === 'cards' ? 'Cards' : 'Allocations Board'}
              </button>
            ))}
          </div>
          {!isReadOnly && (
            <button
              className="btn btn-primary"
              onClick={() => generateDraw.mutate()}
              disabled={generateDraw.isPending}
            >
              Generate/Reset Draw
            </button>
          )}
        </div>

        {round && (
          <div style={{ display: 'flex', gap: '1.5rem', marginTop: '1.25rem', paddingTop: '1rem', borderTop: '1px solid var(--border)', flexWrap: 'wrap' }}>
            <label style={{ display: 'inline-flex', alignItems: 'center', gap: '0.4rem', cursor: isReadOnly ? 'not-allowed' : 'pointer', fontSize: '0.85rem', color: 'var(--text-h)' }}>
              <input
                type="checkbox"
                disabled={isReadOnly}
                checked={!!round.draw_released}
                onChange={e => updateRound('draw_released', e.target.checked)}
              />
              Draw Released
            </label>
            <label style={{ display: 'inline-flex', alignItems: 'center', gap: '0.4rem', cursor: isReadOnly ? 'not-allowed' : 'pointer', fontSize: '0.85rem', color: 'var(--text-h)' }}>
              <input
                type="checkbox"
                disabled={isReadOnly}
                checked={!!round.silent}
                onChange={e => updateRound('silent', e.target.checked)}
              />
              Silent Round
            </label>
            <label style={{ display: 'inline-flex', alignItems: 'center', gap: '0.4rem', cursor: isReadOnly ? 'not-allowed' : 'pointer', fontSize: '0.85rem', color: 'var(--text-h)' }}>
              <input
                type="checkbox"
                disabled={isReadOnly}
                checked={!!round.results_released}
                onChange={e => updateRound('results_released', e.target.checked)}
              />
              Results Released
            </label>
          </div>
        )}
      </div>

      {isLoading ? (
        <div style={{ padding: '3rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>Loading draw...</div>
      ) : draw.length === 0 ? (
        <div style={{ padding: '3rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>
          No matches generated for this round yet.
        </div>
      ) : view === 'board' ? (
        <AllocationsBoard slug={slug} roundId={roundId} draw={draw} />
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
          {draw.map(d => (
            <div key={d.id} className="card" style={{ padding: '1.25rem', background: 'rgba(0,0,0,0.01)', border: '1px solid var(--border)' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '1px solid var(--border)', paddingBottom: '0.5rem', marginBottom: '0.75rem' }}>
                <span style={{ fontWeight: '600', color: 'var(--text-h)', display: 'inline-flex', alignItems: 'center', gap: '0.4rem' }}>
                  <MapPin size={16} /> {d.venue}
                </span>
              </div>

              {/* BP Grid representation */}
              <div className="grid grid-cols-2" style={{ gap: '0.75rem' }}>
                {d.teams.map((t: any) => (
                  <div key={t.team_id} style={{ display: 'flex', justifyContent: 'space-between', padding: '0.5rem', background: 'rgba(0,0,0,0.02)', borderRadius: '6px' }}>
                    <span style={{ fontSize: '0.85rem' }}>{t.team_name}</span>
                    <span className="badge badge-info" style={{ textTransform: 'uppercase' }}>{t.side}</span>
                  </div>
                ))}
              </div>

              <div style={{ marginTop: '0.75rem', fontSize: '0.8rem', color: 'var(--text-mute)' }}>
                Chair Panel: {d.adjudicators.filter((a: any) => a.role === 'chair').map((a: any) => a.adjudicator_name).join(', ') || 'No chair assigned'}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
