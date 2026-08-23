import { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { MapPin } from 'lucide-react';
import { fetchAPI } from '../../../lib/api';

interface DragPayload {
  kind: 'team' | 'adj';
  debateId: string;
  assignmentId: string;
  role?: string;
}

const roleStyles: Record<string, React.CSSProperties> = {
  chair: { fontWeight: 600, color: 'var(--text-h)' },
  panel: {},
  trainee: { color: 'var(--text-mute)', fontStyle: 'italic' },
};

function ConflictBadges({ slug, debateId }: { slug: string; debateId: string }) {
  const { data: conflicts } = useQuery({
    queryKey: ['debate-conflicts', slug, debateId],
    queryFn: () => fetchAPI(`/api/t/${slug}/debates/${debateId}/conflicts`),
  });
  if (!conflicts || (!conflicts.hard.length && !conflicts.soft.length)) return null;
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem', marginTop: '0.5rem' }}>
      {conflicts.hard.map((c: string) => (
        <span key={c} title={c} style={{ fontSize: '0.65rem', padding: '0.15rem 0.4rem', borderRadius: '6px', background: '#fee2e2', color: '#b91c1c' }}>
          {c}
        </span>
      ))}
      {conflicts.soft.map((c: string) => (
        <span key={c} title={c} style={{ fontSize: '0.65rem', padding: '0.15rem 0.4rem', borderRadius: '6px', background: '#ffedd5', color: '#c2410c' }}>
          {c}
        </span>
      ))}
    </div>
  );
}

export default function AllocationsBoard({ slug, roundId, draw }: { slug: string; roundId: string; draw: any[] }) {
  const queryClient = useQueryClient();
  const [dragOverId, setDragOverId] = useState<string | null>(null);

  const move = async (payload: DragPayload, targetDebateId: string) => {
    if (payload.debateId === targetDebateId && payload.kind !== 'adj') return;
    try {
      await fetchAPI(
        `/api/t/${slug}/debates/${payload.debateId}/${payload.kind === 'team' ? 'teams' : 'adjudicators'}/${payload.assignmentId}`,
        'PUT',
        { target_debate_id: targetDebateId, ...(payload.role ? { role: payload.role } : {}) }
      );
      queryClient.invalidateQueries({ queryKey: ['draw', slug, roundId] });
      queryClient.invalidateQueries({ queryKey: ['debate-conflicts', slug] });
    } catch (err: any) {
      alert(err.message);
    }
  };

  const onDrop = (e: React.DragEvent, targetDebateId: string) => {
    e.preventDefault();
    setDragOverId(null);
    const raw = e.dataTransfer.getData('application/json');
    if (!raw) return;
    move(JSON.parse(raw) as DragPayload, targetDebateId);
  };

  const chipProps = (payload: DragPayload) => ({
    draggable: true,
    onDragStart: (e: React.DragEvent) => {
      e.dataTransfer.setData('application/json', JSON.stringify(payload));
      e.dataTransfer.effectAllowed = 'move';
    },
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: '0.35rem',
      padding: '0.3rem 0.5rem',
      background: '#fff',
      border: '1px solid var(--border)',
      borderRadius: '6px',
      cursor: 'grab',
      fontSize: '0.8rem' as const,
    },
  });

  return (
    <div style={{ display: 'flex', gap: '1rem', overflowX: 'auto', paddingBottom: '1rem' }}>
      {draw.map(d => (
        <div
          key={d.id}
          onDragOver={e => {
            e.preventDefault();
            setDragOverId(d.id);
          }}
          onDragLeave={() => setDragOverId(prev => (prev === d.id ? null : prev))}
          onDrop={e => onDrop(e, d.id)}
          className="card"
          style={{
            minWidth: '230px',
            flex: '0 0 auto',
            padding: '1rem',
            border: `1px solid ${dragOverId === d.id ? 'var(--border-focus)' : 'var(--border)'}`,
            background: dragOverId === d.id ? 'rgba(0,0,0,0.03)' : 'rgba(0,0,0,0.01)',
          }}
        >
          <div style={{ fontWeight: 600, color: 'var(--text-h)', display: 'inline-flex', alignItems: 'center', gap: '0.4rem', marginBottom: '0.75rem' }}>
            <MapPin size={14} /> {d.venue}
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
            {d.teams.map((t: any) => (
              <div key={t.id ?? t.team_id} {...chipProps({ kind: 'team', debateId: d.id, assignmentId: t.id })}>
                <span className="badge badge-info" style={{ textTransform: 'uppercase', fontSize: '0.65rem' }}>{t.side}</span>
                <span>{t.team_name}</span>
                {t.pull_up && (
                  <span title="Pulled up from a lower points bracket" style={{ fontSize: '0.6rem', fontWeight: 700, padding: '0.05rem 0.3rem', borderRadius: '4px', background: 'var(--accent)', color: '#fff' }}>
                    PU
                  </span>
                )}
              </div>
            ))}
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.3rem', marginTop: '0.75rem', paddingTop: '0.75rem', borderTop: '1px solid var(--border)' }}>
            {['chair', 'panel', 'trainee'].map(role =>
              d.adjudicators
                .filter((a: any) => a.role === role)
                .map((a: any) => (
                  <div key={a.id ?? a.adjudicator_id} {...chipProps({ kind: 'adj', debateId: d.id, assignmentId: a.id, role })} style={{ ...chipProps({ kind: 'adj', debateId: d.id, assignmentId: a.id, role }).style, ...roleStyles[role] }}>
                    <span style={{ fontSize: '0.6rem', textTransform: 'uppercase', letterSpacing: '0.04em', color: 'var(--text-mute)' }}>{role}</span>
                    <span>{a.adjudicator_name}</span>
                  </div>
                ))
            )}
            {d.adjudicators.length === 0 && (
              <span style={{ fontSize: '0.7rem', color: 'var(--text-mute)' }}>No adjudicators</span>
            )}
          </div>

          <ConflictBadges slug={slug} debateId={d.id} />
        </div>
      ))}
    </div>
  );
}
