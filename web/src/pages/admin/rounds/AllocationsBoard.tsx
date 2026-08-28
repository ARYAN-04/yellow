import { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { MapPin, Accessibility, X, Users, GripVertical } from 'lucide-react';
import { fetchAPI } from '../../../lib/api';

interface DragPayload {
  kind: 'team' | 'adj' | 'scratch_adj';
  debateId?: string;
  assignmentId?: string;
  adjId?: string;
  adjName?: string;
  role?: string;
}

const roleStyles: Record<string, React.CSSProperties> = {
  chair: { fontWeight: 600, color: 'var(--text-h)', background: '#f8fafc', borderColor: '#cbd5e1' },
  panel: { background: '#ffffff', borderColor: 'var(--border)' },
  trainee: { color: 'var(--text-mute)', fontStyle: 'italic', background: '#fafafa', borderColor: 'var(--border)' },
};

function ConflictBadges({ slug, debateId }: { slug: string; debateId: string }) {
  const { data: conflicts } = useQuery({
    queryKey: ['debate-conflicts', slug, debateId],
    queryFn: () => fetchAPI(`/api/t/${slug}/debates/${debateId}/conflicts`),
  });
  if (!conflicts || (!conflicts.hard?.length && !conflicts.soft?.length)) return null;
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem', marginTop: '0.5rem' }}>
      {(conflicts.hard || []).map((c: string) => (
        <span key={c} title={c} style={{ fontSize: '0.65rem', padding: '0.15rem 0.4rem', borderRadius: '6px', background: '#fee2e2', color: '#b91c1c' }}>
          {c}
        </span>
      ))}
      {(conflicts.soft || []).map((c: string) => (
        <span key={c} title={c} style={{ fontSize: '0.65rem', padding: '0.15rem 0.4rem', borderRadius: '6px', background: '#ffedd5', color: '#c2410c' }}>
          {c}
        </span>
      ))}
    </div>
  );
}

export default function AllocationsBoard({
  slug,
  roundId,
  draw = [],
  isReadOnly = false,
  onOpenTrajectory
}: {
  slug: string;
  roundId: string;
  draw: any[];
  isReadOnly?: boolean;
  onOpenTrajectory?: (teamId: string) => void;
}) {
  const queryClient = useQueryClient();
  const [dragOverId, setDragOverId] = useState<string | null>(null);
  const [dragOverScratch, setDragOverScratch] = useState(false);

  // Fetch all tournament adjudicators to compute unallocated scratch pool
  const { data: allAdjudicators = [] } = useQuery<any[]>({
    queryKey: ['adjudicators', slug],
    queryFn: () => fetchAPI(`/api/t/${slug}/adjudicators`),
  });

  // Calculate unallocated judges
  const assignedAdjIds = new Set<string>();
  (draw || []).forEach(d => {
    (d.adjudicators || []).forEach((a: any) => {
      assignedAdjIds.add(a.adjudicator_id);
    });
  });
  const unallocatedAdjudicators = (allAdjudicators || []).filter(a => !assignedAdjIds.has(a.id));

  const invalidateDraw = () => {
    queryClient.invalidateQueries({ queryKey: ['draw', slug, roundId] });
    queryClient.invalidateQueries({ queryKey: ['debate-conflicts', slug] });
  };

  const handleMoveOrAssign = async (payload: DragPayload, targetDebateId: string) => {
    if (isReadOnly) return;

    if (payload.kind === 'scratch_adj' && payload.adjId) {
      // Assign unallocated judge to target debate
      try {
        await fetchAPI(`/api/t/${slug}/debates/${targetDebateId}/adjudicators`, 'POST', {
          adjudicator_id: payload.adjId,
          role: 'panel',
        });
        invalidateDraw();
      } catch (err: any) {
        alert('Failed to assign judge: ' + err.message);
      }
      return;
    }

    if (!payload.debateId || !payload.assignmentId) return;
    if (payload.debateId === targetDebateId && payload.kind !== 'adj') return;

    try {
      await fetchAPI(
        `/api/t/${slug}/debates/${payload.debateId}/${payload.kind === 'team' ? 'teams' : 'adjudicators'}/${payload.assignmentId}`,
        'PUT',
        { target_debate_id: targetDebateId, ...(payload.role ? { role: payload.role } : {}) }
      );
      invalidateDraw();
    } catch (err: any) {
      alert(err.message);
    }
  };

  const handleUnassignJudge = async (debateId: string, assignmentId: string) => {
    if (isReadOnly) return;
    try {
      await fetchAPI(`/api/t/${slug}/debates/${debateId}/adjudicators/${assignmentId}`, 'DELETE');
      invalidateDraw();
    } catch (err: any) {
      alert('Failed to unassign judge: ' + err.message);
    }
  };

  const handleRoleChange = async (debateId: string, assignmentId: string, newRole: string) => {
    if (isReadOnly) return;
    try {
      await fetchAPI(
        `/api/t/${slug}/debates/${debateId}/adjudicators/${assignmentId}`,
        'PUT',
        { target_debate_id: debateId, role: newRole }
      );
      invalidateDraw();
    } catch (err: any) {
      alert('Failed to change role: ' + err.message);
    }
  };

  const handleQuickAssign = async (debateId: string, adjId: string, role: string = 'panel') => {
    if (isReadOnly || !debateId || !adjId) return;
    try {
      await fetchAPI(`/api/t/${slug}/debates/${debateId}/adjudicators`, 'POST', {
        adjudicator_id: adjId,
        role: role,
      });
      invalidateDraw();
    } catch (err: any) {
      alert('Failed to assign judge: ' + err.message);
    }
  };

  const onDropDebate = (e: React.DragEvent, targetDebateId: string) => {
    if (isReadOnly) return;
    e.preventDefault();
    setDragOverId(null);
    const raw = e.dataTransfer.getData('application/json');
    if (!raw) return;
    const payload = JSON.parse(raw) as DragPayload;
    handleMoveOrAssign(payload, targetDebateId);
  };

  const onDropScratchPool = (e: React.DragEvent) => {
    if (isReadOnly) return;
    e.preventDefault();
    setDragOverScratch(false);
    const raw = e.dataTransfer.getData('application/json');
    if (!raw) return;
    const payload = JSON.parse(raw) as DragPayload;
    if (payload.kind === 'adj' && payload.debateId && payload.assignmentId) {
      handleUnassignJudge(payload.debateId, payload.assignmentId);
    }
  };

  const chipProps = (payload: DragPayload) => ({
    draggable: !isReadOnly,
    onDragStart: (e: React.DragEvent) => {
      if (isReadOnly) return;
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
      cursor: isReadOnly ? 'default' : 'grab',
      fontSize: '0.8rem' as const,
    },
  });

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
      {/* ================= SCRATCH POOL / UNALLOCATED JUDGES TRAY ================= */}
      <div
        className="card"
        onDragOver={e => {
          if (isReadOnly) return;
          e.preventDefault();
          setDragOverScratch(true);
        }}
        onDragLeave={() => {
          if (isReadOnly) return;
          setDragOverScratch(false);
        }}
        onDrop={onDropScratchPool}
        style={{
          background: dragOverScratch ? '#f0fdf4' : '#fafafa',
          border: `2px dashed ${dragOverScratch ? '#16a34a' : 'var(--border)'}`,
          padding: '1rem',
          transition: 'all 0.15s ease',
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem', flexWrap: 'wrap', gap: '0.5rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <Users size={16} color="var(--accent)" />
            <h4 style={{ margin: 0, fontSize: '0.95rem' }}>
              Unallocated Adjudicators Scratch Pool ({unallocatedAdjudicators.length})
            </h4>
            <span style={{ fontSize: '0.75rem', color: 'var(--text-mute)' }}>
              {dragOverScratch ? 'Drop judge here to unassign' : 'Drag judges to/from debates or assign below'}
            </span>
          </div>
        </div>

        {unallocatedAdjudicators.length === 0 ? (
          <div style={{ fontSize: '0.8rem', color: 'var(--text-mute)', fontStyle: 'italic', padding: '0.5rem 0' }}>
            All available tournament adjudicators are currently allocated to debate panels.
          </div>
        ) : (
          <div style={{ display: 'flex', gap: '0.6rem', flexWrap: 'wrap' }}>
            {unallocatedAdjudicators.map((adj: any) => (
              <div
                key={adj.id}
                draggable={!isReadOnly}
                onDragStart={e => {
                  if (isReadOnly) return;
                  const payload: DragPayload = {
                    kind: 'scratch_adj',
                    adjId: adj.id,
                    adjName: adj.name,
                  };
                  e.dataTransfer.setData('application/json', JSON.stringify(payload));
                  e.dataTransfer.effectAllowed = 'move';
                }}
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: '0.4rem',
                  padding: '0.35rem 0.6rem',
                  background: '#fff',
                  border: '1px solid var(--border)',
                  borderRadius: '6px',
                  boxShadow: '0 1px 2px rgba(0,0,0,0.03)',
                  cursor: isReadOnly ? 'default' : 'grab',
                  fontSize: '0.8rem',
                }}
              >
                {!isReadOnly && <GripVertical size={13} style={{ color: 'var(--text-mute)', flexShrink: 0 }} />}
                <span style={{ fontWeight: 600, color: 'var(--text-h)' }}>{adj.name}</span>
                {adj.institution_code && (
                  <span className="badge badge-secondary" style={{ fontSize: '0.65rem' }}>{adj.institution_code}</span>
                )}
                <span style={{ fontSize: '0.7rem', color: 'var(--text-mute)' }}>
                  ({adj.test_score})
                </span>

                {!isReadOnly && (
                  <select
                    className="input select"
                    defaultValue=""
                    onChange={e => {
                      if (e.target.value) {
                        handleQuickAssign(e.target.value, adj.id, 'panel');
                        e.target.value = '';
                      }
                    }}
                    style={{
                      fontSize: '0.72rem',
                      padding: '2px 6px',
                      marginLeft: '0.3rem',
                      border: '1px solid var(--border)',
                      borderRadius: '4px',
                      background: 'rgba(0,0,0,0.02)',
                    }}
                    title="Quick assign to room"
                  >
                    <option value="" disabled>+ Assign...</option>
                    {(draw || []).map((d: any) => (
                      <option key={d.id} value={d.id}>
                        {d.venue}
                      </option>
                    ))}
                  </select>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* ================= DEBATE ROOMS BOARD ================= */}
      <div style={{ display: 'flex', gap: '1rem', overflowX: 'auto', paddingBottom: '1rem' }}>
        {(draw || []).map((d: any) => (
          <div
            key={d.id}
            onDragOver={e => {
              if (isReadOnly) return;
              e.preventDefault();
              setDragOverId(d.id);
            }}
            onDragLeave={() => {
              if (isReadOnly) return;
              setDragOverId(prev => (prev === d.id ? null : prev));
            }}
            onDrop={e => onDropDebate(e, d.id)}
            className="card"
            style={{
              minWidth: '250px',
              maxWidth: '300px',
              flex: '0 0 auto',
              padding: '1rem',
              border: `1px solid ${!isReadOnly && dragOverId === d.id ? 'var(--accent)' : 'var(--border)'}`,
              background: !isReadOnly && dragOverId === d.id ? 'rgba(0,0,0,0.03)' : '#fff',
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'space-between',
            }}
          >
            <div>
              {/* Room Header */}
              <div style={{ fontWeight: 600, color: 'var(--text-h)', display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.35rem' }}>
                  <MapPin size={14} /> {d.venue}
                </span>
                {d.venue_accessible && (
                  <span title="Wheelchair Accessible Venue" style={{ display: 'inline-flex', alignItems: 'center', gap: '0.2rem', fontSize: '0.65rem', background: '#dbeafe', color: '#1d4ed8', padding: '1px 6px', borderRadius: '999px', fontWeight: 600 }}>
                    <Accessibility size={11} /> Accessible
                  </span>
                )}
              </div>

              {/* Teams List */}
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
                {(d.teams || []).map((t: any) => (
                  <div key={t.id ?? t.team_id} {...chipProps({ kind: 'team', debateId: d.id, assignmentId: t.id })}>
                    <span className="badge badge-info" style={{ textTransform: 'uppercase', fontSize: '0.65rem' }}>{t.side}</span>
                    <span
                      onClick={e => {
                        if (onOpenTrajectory) {
                          e.stopPropagation();
                          onOpenTrajectory(t.team_id);
                        }
                      }}
                      style={{
                        cursor: onOpenTrajectory ? 'pointer' : undefined,
                        textDecoration: onOpenTrajectory ? 'underline' : undefined,
                        textDecorationStyle: onOpenTrajectory ? 'dotted' : undefined,
                        flex: 1,
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap',
                      }}
                    >
                      {t.team_name}
                    </span>
                    {t.pull_up && (
                      <span title="Pulled up from a lower points bracket" style={{ fontSize: '0.6rem', fontWeight: 700, padding: '0.05rem 0.3rem', borderRadius: '4px', background: 'var(--accent)', color: '#fff' }}>
                        PU
                      </span>
                    )}
                  </div>
                ))}
              </div>

              {/* Adjudicators List */}
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem', marginTop: '0.75rem', paddingTop: '0.75rem', borderTop: '1px solid var(--border)' }}>
                <div style={{ fontSize: '0.7rem', fontWeight: 600, color: 'var(--text-mute)', textTransform: 'uppercase', marginBottom: '0.2rem' }}>
                  Panel ({d.adjudicators?.length || 0})
                </div>

                {['chair', 'panel', 'trainee'].map(role =>
                  (d.adjudicators || [])
                    .filter((a: any) => a.role === role)
                    .map((a: any) => (
                      <div
                        key={a.id ?? a.adjudicator_id}
                        {...chipProps({
                          kind: 'adj',
                          debateId: d.id,
                          assignmentId: a.id,
                          role,
                          adjId: a.adjudicator_id,
                          adjName: a.adjudicator_name,
                        })}
                        style={{
                          ...chipProps({
                            kind: 'adj',
                            debateId: d.id,
                            assignmentId: a.id,
                            role,
                            adjId: a.adjudicator_id,
                            adjName: a.adjudicator_name,
                          }).style,
                          ...roleStyles[role],
                          justifyContent: 'space-between',
                        }}
                      >
                        <div style={{ display: 'inline-flex', alignItems: 'center', gap: '0.3rem', overflow: 'hidden' }}>
                          {!isReadOnly ? (
                            <select
                              value={a.role}
                              onChange={e => handleRoleChange(d.id, a.id, e.target.value)}
                              onClick={e => e.stopPropagation()}
                              style={{
                                fontSize: '0.65rem',
                                padding: '1px 3px',
                                border: '1px solid var(--border)',
                                borderRadius: '4px',
                                background: '#fff',
                                textTransform: 'uppercase',
                                cursor: 'pointer',
                              }}
                            >
                              <option value="chair">Chair</option>
                              <option value="panel">Panel</option>
                              <option value="trainee">Trainee</option>
                            </select>
                          ) : (
                            <span style={{ fontSize: '0.6rem', textTransform: 'uppercase', letterSpacing: '0.04em', color: 'var(--text-mute)' }}>
                              {role}
                            </span>
                          )}
                          <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                            {a.adjudicator_name}
                          </span>
                        </div>

                        {!isReadOnly && (
                          <button
                            type="button"
                            onClick={e => {
                              e.stopPropagation();
                              handleUnassignJudge(d.id, a.id);
                            }}
                            title="Unassign to scratch pool"
                            style={{
                              background: 'transparent',
                              border: 'none',
                              padding: '1px',
                              cursor: 'pointer',
                              color: 'var(--text-mute)',
                              display: 'flex',
                              alignItems: 'center',
                            }}
                          >
                            <X size={13} />
                          </button>
                        )}
                      </div>
                    ))
                )}

                {(d.adjudicators || []).length === 0 && (
                  <span style={{ fontSize: '0.75rem', color: 'var(--text-mute)', fontStyle: 'italic' }}>No judges allocated</span>
                )}
              </div>
            </div>

            <div>
              {/* Quick Add from Scratch Pool */}
              {!isReadOnly && unallocatedAdjudicators.length > 0 && (
                <div style={{ marginTop: '0.6rem', paddingTop: '0.5rem', borderTop: '1px solid rgba(0,0,0,0.06)' }}>
                  <select
                    className="input select"
                    defaultValue=""
                    onChange={e => {
                      if (e.target.value) {
                        handleQuickAssign(d.id, e.target.value, 'panel');
                        e.target.value = '';
                      }
                    }}
                    style={{ width: '100%', fontSize: '0.75rem', padding: '3px 6px' }}
                  >
                    <option value="" disabled>+ Add judge to room...</option>
                    {unallocatedAdjudicators.map((u: any) => (
                      <option key={u.id} value={u.id}>
                        {u.name} {u.institution_code ? `(${u.institution_code})` : ''}
                      </option>
                    ))}
                  </select>
                </div>
              )}

              <ConflictBadges slug={slug} debateId={d.id} />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
