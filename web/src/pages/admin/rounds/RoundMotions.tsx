import { useState, useEffect } from 'react';
import { useOutletContext } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Trash2, Edit2, CheckCircle, Eye, EyeOff, FileText, BarChart2, Award, Info, Printer, Tv } from 'lucide-react';
import { fetchAPI, type RoundContext } from '../../../lib/api';

export default function RoundMotions() {
  const { slug, roundId, isReadOnly, isAssistant } = useOutletContext<RoundContext>();
  const queryClient = useQueryClient();
  const isRestricted = isReadOnly || isAssistant;

  const [activeTab, setActiveTab] = useState<'motions' | 'vetoes' | 'stats'>('motions');

  // Motion Form State
  const [showAddModal, setShowAddModal] = useState(false);
  const [editingMotionId, setEditingMotionId] = useState<string | null>(null);
  const [motionSeq, setMotionSeq] = useState('1');
  const [motionReference, setMotionReference] = useState('');
  const [motionText, setMotionText] = useState('');
  const [motionInfoSlide, setMotionInfoSlide] = useState('');

  // Vetoes State
  const [selectedDebateId, setSelectedDebateId] = useState<string>('');
  const [vetoInputs, setVetoInputs] = useState<Record<string, Record<string, number>>>({}); // teamId -> motionId -> preference

  // Queries
  const { data: motions = [], isLoading: loadingMotions } = useQuery<any[]>({
    queryKey: ['round-motions', slug, roundId],
    queryFn: () => fetchAPI(`/api/t/${slug}/rounds/${roundId}/motions`),
  });

  const { data: draw = [] } = useQuery<any[]>({
    queryKey: ['draw', slug, roundId],
    queryFn: () => fetchAPI(`/api/t/${slug}/rounds/${roundId}/draw`),
  });

  const { data: debateVetoes = [], refetch: refetchDebateVetoes } = useQuery<any[]>({
    queryKey: ['debate-vetoes', slug, selectedDebateId],
    queryFn: () => fetchAPI(`/api/t/${slug}/debates/${selectedDebateId}/vetoes`),
    enabled: !!selectedDebateId,
  });

  const { data: motionStats = [], isLoading: loadingStats } = useQuery<any[]>({
    queryKey: ['motion-stats', slug],
    queryFn: () => fetchAPI(`/api/t/${slug}/motions/statistics`),
    enabled: activeTab === 'stats',
  });

  // Default select first debate when draw loads
  useEffect(() => {
    if (draw.length > 0 && !selectedDebateId) {
      setSelectedDebateId(draw[0].id);
    }
  }, [draw, selectedDebateId]);

  // Sync existing debate vetoes into form state when selected debate vetoes load
  useEffect(() => {
    if (!selectedDebateId) return;
    const initial: Record<string, Record<string, number>> = {};
    debateVetoes.forEach((v: any) => {
      if (!initial[v.team_id]) initial[v.team_id] = {};
      initial[v.team_id][v.motion_id] = v.preference;
    });
    setVetoInputs(initial);
  }, [debateVetoes, selectedDebateId]);

  // Actions
  const resetMotionForm = () => {
    setShowAddModal(false);
    setEditingMotionId(null);
    setMotionSeq(String(motions.length + 1));
    setMotionReference('');
    setMotionText('');
    setMotionInfoSlide('');
  };

  const handleOpenAdd = () => {
    resetMotionForm();
    setMotionSeq(String(motions.length + 1));
    setShowAddModal(true);
  };

  const handleOpenEdit = (m: any) => {
    setEditingMotionId(m.id);
    setMotionSeq(String(m.seq ?? 1));
    setMotionReference(m.reference || '');
    setMotionText(m.text || '');
    setMotionInfoSlide(m.info_slide || '');
    setShowAddModal(true);
  };

  const saveMotionMutation = useMutation({
    mutationFn: async () => {
      const payload = {
        seq: Number(motionSeq) || 1,
        reference: motionReference.trim(),
        text: motionText.trim(),
        info_slide: motionInfoSlide.trim(),
      };
      if (editingMotionId) {
        return fetchAPI(`/api/t/${slug}/motions/${editingMotionId}`, 'PUT', payload);
      }
      return fetchAPI(`/api/t/${slug}/rounds/${roundId}/motions`, 'POST', payload);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['round-motions', slug, roundId] });
      queryClient.invalidateQueries({ queryKey: ['motion-stats', slug] });
      resetMotionForm();
    },
    onError: (err: any) => alert('Failed to save motion: ' + err.message),
  });

  const deleteMotionMutation = useMutation({
    mutationFn: (id: string) => fetchAPI(`/api/t/${slug}/motions/${id}`, 'DELETE'),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['round-motions', slug, roundId] });
      queryClient.invalidateQueries({ queryKey: ['motion-stats', slug] });
    },
    onError: (err: any) => alert('Failed to delete motion: ' + err.message),
  });

  const toggleReleaseMutation = useMutation({
    mutationFn: (release: boolean) =>
      fetchAPI(`/api/t/${slug}/rounds/${roundId}/motions/release`, 'POST', { release }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['round-motions', slug, roundId] });
    },
    onError: (err: any) => alert('Failed to update release status: ' + err.message),
  });

  const saveVetoesMutation = useMutation({
    mutationFn: async () => {
      const vetoList: { team_id: string; motion_id: string; preference: number }[] = [];
      Object.entries(vetoInputs).forEach(([teamId, motionPrefs]) => {
        Object.entries(motionPrefs).forEach(([motionId, pref]) => {
          if (pref) {
            vetoList.push({ team_id: teamId, motion_id: motionId, preference: Number(pref) });
          }
        });
      });
      return fetchAPI(`/api/t/${slug}/debates/${selectedDebateId}/vetoes`, 'POST', { vetoes: vetoList });
    },
    onSuccess: () => {
      refetchDebateVetoes();
      queryClient.invalidateQueries({ queryKey: ['motion-stats', slug] });
      alert('Motion preferences and vetoes saved successfully.');
    },
    onError: (err: any) => alert('Failed to save vetoes: ' + err.message),
  });

  const isRoundReleased = motions.length > 0 && motions.some(m => !!m.released_at);
  const selectedDebate = draw.find(d => d.id === selectedDebateId);

  // Compute preferred motion preview for selected debate
  const motionPreferenceSums: Record<string, number> = {};
  if (selectedDebate && motions.length > 0) {
    motions.forEach(m => {
      let sum = 0;
      selectedDebate.teams.forEach((t: any) => {
        const pref = vetoInputs[t.team_id]?.[m.id] ?? 0;
        sum += pref;
      });
      motionPreferenceSums[m.id] = sum;
    });
  }

  return (
    <div>
      <div className="tabs" style={{ marginBottom: '1.25rem' }}>
        <button
          className={`tab-btn ${activeTab === 'motions' ? 'active' : ''}`}
          onClick={() => setActiveTab('motions')}
        >
          <FileText size={15} style={{ marginRight: '0.35rem', verticalAlign: 'middle' }} />
          Motions & Info Slides ({motions.length})
        </button>
        <button
          className={`tab-btn ${activeTab === 'vetoes' ? 'active' : ''}`}
          onClick={() => setActiveTab('vetoes')}
        >
          <Award size={15} style={{ marginRight: '0.35rem', verticalAlign: 'middle' }} />
          Motion Vetoes & Preferences
        </button>
        <button
          className={`tab-btn ${activeTab === 'stats' ? 'active' : ''}`}
          onClick={() => setActiveTab('stats')}
        >
          <BarChart2 size={15} style={{ marginRight: '0.35rem', verticalAlign: 'middle' }} />
          Motion Balance Statistics
        </button>
      </div>

      {/* ================= TAB 1: MOTIONS & INFO SLIDES ================= */}
      {activeTab === 'motions' && (
        <div>
          <div className="card" style={{ marginBottom: '1.5rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
              <div>
                <h3 style={{ margin: 0 }}>Round Motions</h3>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginTop: '0.35rem' }}>
                  <span style={{ fontSize: '0.8rem', color: 'var(--text-mute)' }}>Release Status:</span>
                  {isRoundReleased ? (
                    <span className="badge" style={{ background: '#dcfce7', color: '#15803d', display: 'inline-flex', alignItems: 'center', gap: '0.25rem' }}>
                      <CheckCircle size={12} /> Released (Live to participants)
                    </span>
                  ) : (
                    <span className="badge" style={{ background: 'rgba(0,0,0,0.06)', color: 'var(--text-mute)' }}>
                      Draft (Hidden from participants)
                    </span>
                  )}
                </div>
              </div>

              <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap', alignItems: 'center' }}>
                <button
                  className="btn btn-secondary"
                  style={{ fontSize: '0.8rem', padding: '0.35rem 0.75rem', display: 'inline-flex', alignItems: 'center', gap: '0.3rem' }}
                  onClick={() => window.open(`/t/${slug}/admin/rounds/${roundId}/print/motions`, '_blank')}
                  title="Open printable motion slips in new tab"
                >
                  <Printer size={14} /> Print Motion Slips
                </button>
                <button
                  className="btn btn-secondary"
                  style={{ fontSize: '0.8rem', padding: '0.35rem 0.75rem', display: 'inline-flex', alignItems: 'center', gap: '0.3rem' }}
                  onClick={() => window.open(`/t/${slug}/display/motion?round_id=${roundId}`, '_blank')}
                  title="Open live projector view in new tab"
                >
                  <Tv size={14} /> Projector Mode
                </button>

                {!isRestricted && (
                  <>
                    <button
                      className="btn btn-secondary"
                      onClick={() => toggleReleaseMutation.mutate(!isRoundReleased)}
                      disabled={motions.length === 0 || toggleReleaseMutation.isPending}
                      style={{ display: 'inline-flex', alignItems: 'center', gap: '0.35rem' }}
                    >
                      {isRoundReleased ? <EyeOff size={15} /> : <Eye size={15} />}
                      {isRoundReleased ? 'Unrelease Motions' : 'Release Motions'}
                    </button>
                    <button
                      className="btn btn-primary"
                      onClick={handleOpenAdd}
                      style={{ display: 'inline-flex', alignItems: 'center', gap: '0.35rem' }}
                    >
                      <Plus size={15} /> Add Motion
                    </button>
                  </>
                )}
              </div>
            </div>
          </div>

          {/* Add / Edit Motion Inline Modal / Form */}
          {showAddModal && (
            <div className="card" style={{ marginBottom: '1.5rem', background: '#fafafa', border: '1px solid var(--border-focus)' }}>
              <h4 style={{ margin: '0 0 1rem 0' }}>{editingMotionId ? 'Edit Motion' : 'Add New Motion'}</h4>
              <form
                onSubmit={e => {
                  e.preventDefault();
                  saveMotionMutation.mutate();
                }}
              >
                <div style={{ display: 'flex', gap: '0.75rem', marginBottom: '0.75rem' }}>
                  <div style={{ width: '90px' }}>
                    <label style={{ fontSize: '0.75rem', fontWeight: 600, display: 'block', marginBottom: '0.25rem' }}>Seq</label>
                    <input
                      type="number"
                      className="input"
                      value={motionSeq}
                      onChange={e => setMotionSeq(e.target.value)}
                      required
                    />
                  </div>
                  <div style={{ flex: 1 }}>
                    <label style={{ fontSize: '0.75rem', fontWeight: 600, display: 'block', marginBottom: '0.25rem' }}>Reference Code (optional)</label>
                    <input
                      type="text"
                      className="input"
                      placeholder="e.g. R1-M1 or THW Taxes"
                      value={motionReference}
                      onChange={e => setMotionReference(e.target.value)}
                    />
                  </div>
                </div>

                <div className="form-group" style={{ marginBottom: '0.75rem' }}>
                  <label style={{ fontSize: '0.75rem', fontWeight: 600, display: 'block', marginBottom: '0.25rem' }}>Motion Text *</label>
                  <textarea
                    className="input"
                    rows={3}
                    placeholder="This House would..."
                    value={motionText}
                    onChange={e => setMotionText(e.target.value)}
                    required
                    style={{ width: '100%', resize: 'vertical' }}
                  />
                </div>

                <div className="form-group" style={{ marginBottom: '1rem' }}>
                  <label style={{ fontSize: '0.75rem', fontWeight: 600, display: 'block', marginBottom: '0.25rem' }}>Info Slide / Background Context (optional)</label>
                  <textarea
                    className="input"
                    rows={3}
                    placeholder="Provide definitions or contextual background for debaters..."
                    value={motionInfoSlide}
                    onChange={e => setMotionInfoSlide(e.target.value)}
                    style={{ width: '100%', resize: 'vertical' }}
                  />
                </div>

                <div style={{ display: 'flex', gap: '0.5rem' }}>
                  <button type="submit" className="btn btn-primary" disabled={saveMotionMutation.isPending}>
                    {editingMotionId ? 'Save Changes' : 'Create Motion'}
                  </button>
                  <button type="button" className="btn" onClick={resetMotionForm}>
                    Cancel
                  </button>
                </div>
              </form>
            </div>
          )}

          {/* Motions List */}
          {loadingMotions ? (
            <div style={{ padding: '3rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>Loading motions...</div>
          ) : motions.length === 0 ? (
            <div className="card" style={{ padding: '3rem 0', textAlign: 'center', color: 'var(--text-mute)' }}>
              No motions added for this round yet. Click <strong>Add Motion</strong> to configure one.
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              {motions.map((m: any) => (
                <div key={m.id} className="card" style={{ border: '1px solid var(--border)', padding: '1.25rem' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: '1rem' }}>
                    <div style={{ flex: 1 }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.5rem' }}>
                        <span className="badge badge-info" style={{ fontWeight: 600 }}>
                          #{m.seq}
                        </span>
                        {m.reference && (
                          <span style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-mute)' }}>
                            [{m.reference}]
                          </span>
                        )}
                        {m.released_at ? (
                          <span style={{ fontSize: '0.7rem', color: '#15803d' }}>
                            Released: {new Date(m.released_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                          </span>
                        ) : (
                          <span style={{ fontSize: '0.7rem', color: 'var(--text-mute)' }}>Draft</span>
                        )}
                      </div>

                      <div style={{ fontSize: '1.05rem', fontWeight: 600, color: 'var(--text-h)', lineHeight: 1.4 }}>
                        {m.text}
                      </div>

                      {m.info_slide && (
                        <div style={{ marginTop: '0.75rem', padding: '0.75rem', background: 'rgba(0,0,0,0.02)', borderLeft: '3px solid var(--accent)', borderRadius: '4px' }}>
                          <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-h)', display: 'inline-flex', alignItems: 'center', gap: '0.3rem', marginBottom: '0.25rem' }}>
                            <Info size={13} /> Info Slide:
                          </div>
                          <div style={{ fontSize: '0.85rem', color: 'var(--text)', whiteSpace: 'pre-wrap' }}>
                            {m.info_slide}
                          </div>
                        </div>
                      )}
                    </div>

                    {!isRestricted && (
                      <div style={{ display: 'flex', gap: '0.35rem' }}>
                        <button
                          className="btn"
                          style={{ padding: '4px 8px' }}
                          title="Edit Motion"
                          onClick={() => handleOpenEdit(m)}
                        >
                          <Edit2 size={14} />
                        </button>
                        <button
                          className="btn btn-danger"
                          style={{ padding: '4px 8px' }}
                          title="Delete Motion"
                          onClick={() => {
                            if (confirm('Delete this motion?')) {
                              deleteMotionMutation.mutate(m.id);
                            }
                          }}
                        >
                          <Trash2 size={14} />
                        </button>
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* ================= TAB 2: MOTION VETOES & PREFERENCES ================= */}
      {activeTab === 'vetoes' && (
        <div>
          <div className="card" style={{ marginBottom: '1.5rem' }}>
            <h3 style={{ margin: '0 0 0.5rem 0' }}>Motion Vetoes & Preferences</h3>
            <p style={{ fontSize: '0.8rem', color: 'var(--text-mute)', margin: '0 0 1rem 0' }}>
              Record team preference rankings (e.g., 1 = 1st choice, 2 = 2nd choice, 3 = veto) for multi-motion rounds.
            </p>

            {draw.length === 0 ? (
              <div style={{ padding: '1rem 0', color: 'var(--text-mute)' }}>
                Please generate the round draw first to allocate debates.
              </div>
            ) : motions.length === 0 ? (
              <div style={{ padding: '1rem 0', color: 'var(--text-mute)' }}>
                No motions created for this round. Please add motions under the Motions tab first.
              </div>
            ) : (
              <div>
                <label style={{ fontSize: '0.8rem', fontWeight: 600, display: 'block', marginBottom: '0.35rem' }}>
                  Select Debate
                </label>
                <select
                  className="input select"
                  value={selectedDebateId}
                  onChange={e => setSelectedDebateId(e.target.value)}
                  style={{ maxWidth: '400px' }}
                >
                  {(draw || []).map((d: any, idx: number) => (
                    <option key={d.id} value={d.id}>
                      Debate {idx + 1}: {d.venue} ({(d.teams || []).map((t: any) => t.team_name).join(' vs ')})
                    </option>
                  ))}
                </select>
              </div>
            )}
          </div>

          {selectedDebate && (motions || []).length > 0 && (
            <div style={{ display: 'grid', gridTemplateColumns: 'minmax(0, 1fr) 300px', gap: '1.5rem', alignItems: 'start' }}>
              {/* Preferences Entry Grid */}
              <div className="card">
                <h4 style={{ margin: '0 0 1rem 0' }}>Team Preferences for {selectedDebate.venue}</h4>

                <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
                  {(selectedDebate.teams || []).map((t: any) => (
                    <div key={t.team_id} style={{ border: '1px solid var(--border)', borderRadius: '6px', padding: '0.75rem' }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
                        <span style={{ fontWeight: 600, fontSize: '0.9rem' }}>{t.team_name}</span>
                        <span className="badge badge-info" style={{ textTransform: 'uppercase' }}>{t.side}</span>
                      </div>

                      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '0.75rem' }}>
                        {motions.map((m: any) => {
                          const currentPref = vetoInputs[t.team_id]?.[m.id] ?? '';
                          return (
                            <div key={m.id} style={{ background: 'rgba(0,0,0,0.02)', padding: '0.5rem', borderRadius: '4px' }}>
                              <div style={{ fontSize: '0.75rem', fontWeight: 600, marginBottom: '0.25rem', color: 'var(--text-h)' }}>
                                #{m.seq} {m.reference ? `[${m.reference}]` : ''}
                              </div>
                              <div style={{ fontSize: '0.75rem', color: 'var(--text-mute)', marginBottom: '0.5rem', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                                {m.text}
                              </div>
                              <select
                                className="input select"
                                value={currentPref}
                                disabled={isReadOnly}
                                onChange={e => {
                                  const val = Number(e.target.value);
                                  setVetoInputs(prev => ({
                                    ...prev,
                                    [t.team_id]: {
                                      ...(prev[t.team_id] || {}),
                                      [m.id]: val,
                                    },
                                  }));
                                }}
                                style={{ width: '100%', fontSize: '0.8rem', padding: '0.25rem 0.5rem' }}
                              >
                                <option value="">Select rank...</option>
                                {motions.map((_, idx) => (
                                  <option key={idx + 1} value={idx + 1}>
                                    {idx === 0 ? '1 (1st Choice)' : idx === motions.length - 1 ? `${idx + 1} (Veto / Last)` : `${idx + 1} Choice`}
                                  </option>
                                ))}
                              </select>
                            </div>
                          );
                        })}
                      </div>
                    </div>
                  ))}
                </div>

                {!isReadOnly && (
                  <div style={{ marginTop: '1.25rem', paddingTop: '1rem', borderTop: '1px solid var(--border)' }}>
                    <button
                      className="btn btn-primary"
                      onClick={() => saveVetoesMutation.mutate()}
                      disabled={saveVetoesMutation.isPending}
                    >
                      Save Debate Preferences
                    </button>
                  </div>
                )}
              </div>

              {/* Veto Calculation Summary Sidebar */}
              <div className="card" style={{ background: '#fafafa' }}>
                <h4 style={{ margin: '0 0 0.75rem 0', display: 'inline-flex', alignItems: 'center', gap: '0.35rem' }}>
                  <Award size={16} /> Resulting Motion
                </h4>
                <p style={{ fontSize: '0.75rem', color: 'var(--text-mute)', margin: '0 0 1rem 0' }}>
                  The motion with the lowest total preference score is chosen for this debate.
                </p>

                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                  {motions.map((m: any) => {
                    const score = motionPreferenceSums[m.id] ?? 0;
                    const isLowest = score > 0 && Math.min(...Object.values(motionPreferenceSums).filter(v => v > 0)) === score;
                    return (
                      <div
                        key={m.id}
                        style={{
                          padding: '0.6rem',
                          borderRadius: '6px',
                          border: `1px solid ${isLowest ? 'var(--accent)' : 'var(--border)'}`,
                          background: isLowest ? 'rgba(0,0,0,0.03)' : '#fff',
                        }}
                      >
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                          <span style={{ fontSize: '0.8rem', fontWeight: 600 }}>
                            #{m.seq} {m.reference ? `[${m.reference}]` : ''}
                          </span>
                          <span className="badge" style={{ background: isLowest ? 'var(--accent)' : 'rgba(0,0,0,0.06)', color: isLowest ? '#fff' : 'var(--text-mute)' }}>
                            Score: {score || '-'}
                          </span>
                        </div>
                        <div style={{ fontSize: '0.75rem', color: 'var(--text-mute)', marginTop: '0.25rem', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                          {m.text}
                        </div>
                        {isLowest && (
                          <div style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--accent)', marginTop: '0.3rem' }}>
                            ⭐ Selected Motion for Debate
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* ================= TAB 3: MOTION BALANCE STATISTICS ================= */}
      {activeTab === 'stats' && (
        <div className="card">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
            <div>
              <h3 style={{ margin: 0 }}>Motion Balance & Side Statistics</h3>
              <p style={{ fontSize: '0.8rem', color: 'var(--text-mute)', margin: '0.25rem 0 0 0' }}>
                Track side win rates (1st places) and position distributions across confirmed debates to detect side bias.
              </p>
            </div>
          </div>

          {loadingStats ? (
            <div style={{ padding: '3rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>Loading motion statistics...</div>
          ) : motionStats.length === 0 ? (
            <div style={{ padding: '3rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>
              No debate results or motion statistics recorded yet.
            </div>
          ) : (
            <div style={{ overflowX: 'auto' }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.85rem' }}>
                <thead>
                  <tr style={{ borderBottom: '2px solid var(--border)', textAlign: 'left' }}>
                    <th style={{ padding: '0.6rem 0.75rem' }}>Motion</th>
                    <th style={{ padding: '0.6rem 0.75rem' }}>Round</th>
                    <th style={{ padding: '0.6rem 0.75rem', textAlign: 'center' }}>Debates</th>
                    <th style={{ padding: '0.6rem 0.75rem' }}>Side Win Rates</th>
                    <th style={{ padding: '0.6rem 0.75rem' }}>Position Distribution</th>
                  </tr>
                </thead>
                <tbody>
                  {motionStats.map((s: any) => (
                    <tr key={s.motion_id} style={{ borderBottom: '1px solid var(--border)' }}>
                      <td style={{ padding: '0.75rem', maxWidth: '300px' }}>
                        <div style={{ fontWeight: 600, color: 'var(--text-h)' }}>
                          {s.reference ? `[${s.reference}] ` : ''}{s.text}
                        </div>
                      </td>
                      <td style={{ padding: '0.75rem', whiteSpace: 'nowrap', color: 'var(--text-mute)' }}>
                        {s.round_name}
                      </td>
                      <td style={{ padding: '0.75rem', textAlign: 'center', fontWeight: 600 }}>
                        {s.total_debates}
                      </td>
                      <td style={{ padding: '0.75rem', minWidth: '220px' }}>
                        {s.total_debates === 0 ? (
                          <span style={{ color: 'var(--text-mute)' }}>No confirmed ballots</span>
                        ) : (
                          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.35rem' }}>
                            {Object.entries(s.side_percentages || {}).map(([side, pct]: [string, any]) => (
                              <span
                                key={side}
                                className="badge"
                                style={{
                                  background: 'rgba(0,0,0,0.04)',
                                  border: '1px solid var(--border)',
                                  fontSize: '0.75rem',
                                }}
                              >
                                <strong>{side}:</strong> {pct}% ({s.side_wins?.[side] || 0})
                              </span>
                            ))}
                          </div>
                        )}
                      </td>
                      <td style={{ padding: '0.75rem', minWidth: '200px' }}>
                        {s.total_debates === 0 ? (
                          <span style={{ color: 'var(--text-mute)' }}>-</span>
                        ) : (
                          <div style={{ fontSize: '0.75rem', color: 'var(--text-mute)', display: 'flex', flexDirection: 'column', gap: '0.2rem' }}>
                            {Object.entries(s.positional_counts || {}).map(([side, ranks]: [string, any]) => (
                              <div key={side} style={{ display: 'flex', gap: '0.5rem' }}>
                                <span style={{ fontWeight: 600, width: '35px' }}>{side}:</span>
                                <span>
                                  {Object.entries(ranks || {}).map(([rank, count]) => `${rank}st: ${count}`).join(' | ')}
                                </span>
                              </div>
                            ))}
                          </div>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
