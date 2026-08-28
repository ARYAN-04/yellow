import { useState } from 'react';
import { useParams, useOutletContext } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Trophy, Upload } from 'lucide-react';
import { fetchAPI, type AdminContext } from '../../lib/api';

const baseTabs = [
  { key: 'open', label: 'Open' },
  { key: 'novice', label: 'Novice' },
  { key: 'esl', label: 'ESL' },
  { key: 'efl', label: 'EFL' },
];

export default function Breaks() {
  const { slug } = useParams<{ slug: string }>();
  const { isReadOnly, isAssistant } = useOutletContext<AdminContext>();
  const isRestricted = isReadOnly || isAssistant;
  const queryClient = useQueryClient();
  const [category, setCategory] = useState('open');

  const { data: categories = [] } = useQuery<any[]>({
    queryKey: ['break-categories', slug],
    queryFn: () => fetchAPI(`/api/t/${slug}/break-categories`),
  });

  const { data: breakResult, isLoading } = useQuery<any>({
    queryKey: ['break', slug, category],
    queryFn: () => fetchAPI(`/api/t/${slug}/breaks/${category}`),
  });

  const publish = useMutation({
    mutationFn: () => fetchAPI(`/api/t/${slug}/breaks/${category}`, 'POST'),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['break', slug, category] }),
    onError: (err: any) => alert('Failed to publish break: ' + err.message),
  });

  const tabs = [...baseTabs, ...(categories || []).map((c: any) => ({ key: c.id, label: c.name }))];
  const qualifiers = breakResult?.qualifiers || [];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
        <h2 style={{ display: 'inline-flex', alignItems: 'center', gap: '0.5rem', margin: 0 }}>
          <Trophy size={20} /> Break
        </h2>
        {!isRestricted && (
          <button className="btn btn-primary" onClick={() => publish.mutate()} disabled={publish.isPending}>
            <Upload size={15} /> Publish Break Snapshot
          </button>
        )}
      </div>

      <div className="card" style={{ marginTop: '1rem' }}>
        <div className="tabs" style={{ marginBottom: '1rem' }}>
          {tabs.map(t => (
            <button key={t.key} className={`tab-btn ${category === t.key ? 'active' : ''}`} onClick={() => setCategory(t.key)}>
              {t.label}
            </button>
          ))}
        </div>

        {breakResult && (
          <p style={{ color: 'var(--text-mute)', fontSize: '0.85rem', margin: '0 0 1rem' }}>
            {breakResult.category_name} &middot; cutoff at {breakResult.cutoff ?? 0} points
            {breakResult.size ? ` (break size ${breakResult.size})` : ''}
          </p>
        )}

        {isLoading ? (
          <div style={{ padding: '3rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>Computing break...</div>
        ) : !breakResult || qualifiers.length === 0 ? (
          <div style={{ padding: '3rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>
            No qualifiers computed yet — confirm some round results first.
          </div>
        ) : (
          <div className="table-wrapper">
            <table className="table">
              <thead>
                <tr>
                  <th>Rank</th>
                  <th>Team</th>
                  <th>Points</th>
                  <th>Speaker Points</th>
                  <th>Margin</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {qualifiers.map((q: any) => (
                  <tr key={q.team_id} style={q.rank > (breakResult?.size ?? Infinity) ? { opacity: 0.55 } : undefined}>
                    <td style={{ fontWeight: '600', color: 'var(--accent)' }}>{q.rank}</td>
                    <td style={{ fontWeight: '500', color: 'var(--text-h)' }}>
                      {q.team_name}
                      {q.is_novice || q.is_esl || q.is_efl ? (
                        <span style={{ marginLeft: '0.4rem', fontSize: '0.7rem', color: 'var(--text-mute)' }}>
                          {[q.is_novice && 'N', q.is_esl && 'ESL', q.is_efl && 'EFL'].filter(Boolean).join(' / ')}
                        </span>
                      ) : null}
                    </td>
                    <td>{q.points}</td>
                    <td>{q.speaker_points?.toFixed(1) || '0.0'}</td>
                    <td>{(q.margin || 0).toFixed(1)}</td>
                    <td>{q.bubble && <span className="badge badge-warning">Bubble</span>}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
