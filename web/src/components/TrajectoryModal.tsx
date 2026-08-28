import { useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { X, MapPin, User, Users, Award } from 'lucide-react';
import { fetchAPI } from '../lib/api';

interface TrajectoryModalProps {
  slug: string;
  type: 'team' | 'speaker' | 'adjudicator';
  id: string | null;
  onClose: () => void;
}

export default function TrajectoryModal({ slug, type, id, onClose }: TrajectoryModalProps) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [onClose]);

  const endpointPath =
    type === 'team'
      ? `teams/${id}/trajectory`
      : type === 'speaker'
      ? `speakers/${id}/trajectory`
      : `adjudicators/${id}/trajectory`;

  const { data, isLoading, error } = useQuery<any>({
    queryKey: ['trajectory', slug, type, id],
    queryFn: () => {
      if (!id) return null;
      return fetchAPI(`/api/t/${slug}/${endpointPath}`);
    },
    enabled: !!id,
  });

  if (!id) return null;

  const getTitle = () => {
    if (type === 'team') return data?.team?.name || 'Team Trajectory';
    if (type === 'speaker') return data?.speaker?.name || 'Speaker Trajectory';
    return data?.adjudicator?.name || 'Adjudicator Trajectory';
  };

  return (
    <div
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        backgroundColor: 'rgba(0, 0, 0, 0.6)',
        backdropFilter: 'blur(4px)',
        zIndex: 9999,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '1.5rem',
      }}
      onClick={onClose}
    >
      <div
        style={{
          background: '#fff',
          borderRadius: '12px',
          width: '100%',
          maxWidth: '780px',
          maxHeight: '88vh',
          display: 'flex',
          flexDirection: 'column',
          boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25)',
          overflow: 'hidden',
        }}
        onClick={e => e.stopPropagation()}
      >
        {/* Modal Header */}
        <div
          style={{
            padding: '1.25rem 1.5rem',
            borderBottom: '1px solid var(--border)',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            background: 'var(--bg-card, #fff)',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
            <div
              style={{
                padding: '0.4rem',
                borderRadius: '8px',
                background: type === 'team' ? '#dbeafe' : type === 'speaker' ? '#fef3c7' : '#dcfce7',
                color: type === 'team' ? '#1d4ed8' : type === 'speaker' ? '#d97706' : '#15803d',
                display: 'flex',
                alignItems: 'center',
              }}
            >
              {type === 'team' ? <Users size={20} /> : type === 'speaker' ? <User size={20} /> : <Award size={20} />}
            </div>
            <div>
              <h3 style={{ margin: 0, fontSize: '1.2rem', fontWeight: 700, color: 'var(--text-h)' }}>
                {getTitle()}
              </h3>
              <div style={{ fontSize: '0.8rem', color: 'var(--text-mute)' }}>
                {type === 'team' ? (
                  <span>Code: {data?.team?.code || 'N/A'} {data?.team?.institution_code ? `• ${data.team.institution_code}` : ''}</span>
                ) : type === 'speaker' ? (
                  <span>Team: <strong>{data?.team_name || 'N/A'}</strong></span>
                ) : (
                  <span>Institution: {data?.adjudicator?.institution_name || data?.adjudicator?.institution_code || 'Independent'}</span>
                )}
              </div>
            </div>
          </div>
          <button
            onClick={onClose}
            style={{
              background: 'transparent',
              border: 'none',
              cursor: 'pointer',
              color: 'var(--text-mute)',
              padding: '0.4rem',
              borderRadius: '6px',
              display: 'flex',
              alignItems: 'center',
            }}
          >
            <X size={20} />
          </button>
        </div>

        {/* Modal Body */}
        <div style={{ padding: '1.5rem', overflowY: 'auto', flex: 1 }}>
          {isLoading ? (
            <div style={{ textAlign: 'center', padding: '3rem 0', color: 'var(--text-mute)' }}>
              Loading trajectory history...
            </div>
          ) : error ? (
            <div style={{ textAlign: 'center', padding: '2rem 0', color: 'var(--danger)' }}>
              Failed to load trajectory: {(error as any).message}
            </div>
          ) : type === 'team' ? (
            <TeamTrajectoryView data={data} />
          ) : type === 'speaker' ? (
            <SpeakerTrajectoryView data={data} />
          ) : (
            <AdjudicatorTrajectoryView data={data} />
          )}
        </div>
      </div>
    </div>
  );
}

function TeamTrajectoryView({ data }: { data: any }) {
  const debates: any[] = data?.debates || [];
  const team = data?.team;

  const totalPoints = debates.reduce((acc, d) => acc + (d.points ?? 0), 0);
  const totalSpeakerPoints = debates.reduce((acc, d) => acc + (d.speaker_points ?? 0), 0);
  const completedDebates = debates.filter(d => d.points != null).length;

  return (
    <div>
      <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1.25rem', flexWrap: 'wrap' }}>
        {team?.is_novice && <span className="badge badge-warning">Novice</span>}
        {team?.is_esl && <span className="badge badge-info">ESL</span>}
        {team?.is_efl && <span className="badge badge-info">EFL</span>}
        {team?.is_standby && <span className="badge badge-secondary">Standby</span>}
      </div>

      <div className="grid grid-cols-3" style={{ gap: '1rem', marginBottom: '1.5rem' }}>
        <div className="card" style={{ padding: '0.75rem 1rem', textAlign: 'center', background: 'rgba(0,0,0,0.02)' }}>
          <div style={{ fontSize: '0.75rem', textTransform: 'uppercase', color: 'var(--text-mute)', marginBottom: '0.25rem' }}>
            Wins / Points
          </div>
          <div style={{ fontSize: '1.4rem', fontWeight: 800, color: 'var(--accent)' }}>
            {totalPoints}
          </div>
        </div>
        <div className="card" style={{ padding: '0.75rem 1rem', textAlign: 'center', background: 'rgba(0,0,0,0.02)' }}>
          <div style={{ fontSize: '0.75rem', textTransform: 'uppercase', color: 'var(--text-mute)', marginBottom: '0.25rem' }}>
            Total Spkr Pts
          </div>
          <div style={{ fontSize: '1.4rem', fontWeight: 800, color: 'var(--text-h)' }}>
            {totalSpeakerPoints.toFixed(1)}
          </div>
        </div>
        <div className="card" style={{ padding: '0.75rem 1rem', textAlign: 'center', background: 'rgba(0,0,0,0.02)' }}>
          <div style={{ fontSize: '0.75rem', textTransform: 'uppercase', color: 'var(--text-mute)', marginBottom: '0.25rem' }}>
            Debates
          </div>
          <div style={{ fontSize: '1.4rem', fontWeight: 800, color: 'var(--text-h)' }}>
            {completedDebates} / {debates.length}
          </div>
        </div>
      </div>

      <h4 style={{ margin: '0 0 0.75rem 0', fontSize: '0.95rem', fontWeight: 700 }}>Debate Timeline</h4>
      {debates.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '2rem 0', color: 'var(--text-mute)', fontSize: '0.85rem' }}>
          No debate pairings found for this team.
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
          {debates.map((d: any) => (
            <div
              key={d.debate_id || d.round_id}
              style={{
                border: '1px solid var(--border)',
                borderRadius: '8px',
                padding: '1rem',
                background: d.points != null ? '#fff' : 'rgba(0,0,0,0.01)',
              }}
            >
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <span style={{ fontWeight: 700, fontSize: '0.9rem', color: 'var(--text-h)' }}>
                    {d.round_name}
                  </span>
                  <span className="badge badge-info" style={{ textTransform: 'uppercase', fontSize: '0.7rem' }}>
                    {d.side} {d.pull_up ? '(PU)' : ''}
                  </span>
                  <span style={{ fontSize: '0.8rem', color: 'var(--text-mute)', display: 'inline-flex', alignItems: 'center', gap: '0.2rem' }}>
                    <MapPin size={12} /> {d.venue}
                  </span>
                </div>
                <div>
                  {d.points != null ? (
                    <span
                      style={{
                        fontWeight: 800,
                        fontSize: '0.9rem',
                        color: d.points > 0 ? '#16a34a' : 'var(--text-h)',
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: '0.3rem',
                      }}
                    >
                      {d.points} Pts ({d.speaker_points?.toFixed(1)} spkrs)
                    </span>
                  ) : (
                    <span style={{ fontSize: '0.75rem', color: 'var(--text-mute)' }}>Pending Results</span>
                  )}
                </div>
              </div>

              <div style={{ fontSize: '0.8rem', color: 'var(--text-mute)', display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                {d.opponents?.length > 0 && (
                  <div>
                    <strong>vs. </strong>
                    {d.opponents.map((opp: any) => `${opp.team_name} (${opp.side})`).join(' • ')}
                  </div>
                )}
                {d.adjudicators?.length > 0 && (
                  <div>
                    <strong>Panel: </strong>
                    {d.adjudicators.map((a: any) => `${a.name}${a.role === 'chair' ? ' (Ⓒ)' : ''}`).join(', ')}
                  </div>
                )}
              </div>

              {d.speaker_scores?.length > 0 && (
                <div
                  style={{
                    marginTop: '0.6rem',
                    paddingTop: '0.5rem',
                    borderTop: '1px solid var(--border)',
                    display: 'flex',
                    gap: '1rem',
                    fontSize: '0.8rem',
                  }}
                >
                  {d.speaker_scores.map((sc: any, idx: number) => (
                    <div key={idx} style={{ display: 'inline-flex', alignItems: 'center', gap: '0.3rem' }}>
                      <span style={{ color: 'var(--text-mute)' }}>{sc.speaker_name}{sc.is_reply ? ' (R)' : ''}:</span>
                      <strong style={{ color: 'var(--text-h)' }}>{sc.score.toFixed(1)}</strong>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function SpeakerTrajectoryView({ data }: { data: any }) {
  const speeches: any[] = data?.speeches || [];
  const speaker = data?.speaker;

  const validScores = speeches.filter(s => s.score != null).map(s => s.score);
  const totalScore = validScores.reduce((acc, sc) => acc + sc, 0);
  const avgScore = validScores.length > 0 ? totalScore / validScores.length : 0;

  let trimmedAvg = avgScore;
  if (validScores.length >= 3) {
    const sorted = [...validScores].sort((a, b) => a - b);
    const trimmed = sorted.slice(1, sorted.length - 1);
    trimmedAvg = trimmed.reduce((a, b) => a + b, 0) / trimmed.length;
  }

  return (
    <div>
      <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1.25rem', flexWrap: 'wrap' }}>
        {speaker?.is_novice && <span className="badge badge-warning">Novice</span>}
        {speaker?.is_esl && <span className="badge badge-info">ESL</span>}
        {speaker?.is_efl && <span className="badge badge-info">EFL</span>}
      </div>

      <div className="grid grid-cols-4" style={{ gap: '0.75rem', marginBottom: '1.5rem' }}>
        <div className="card" style={{ padding: '0.75rem 0.5rem', textAlign: 'center', background: 'rgba(0,0,0,0.02)' }}>
          <div style={{ fontSize: '0.7rem', textTransform: 'uppercase', color: 'var(--text-mute)', marginBottom: '0.2rem' }}>
            Speeches
          </div>
          <div style={{ fontSize: '1.3rem', fontWeight: 800, color: 'var(--text-h)' }}>
            {validScores.length}
          </div>
        </div>
        <div className="card" style={{ padding: '0.75rem 0.5rem', textAlign: 'center', background: 'rgba(0,0,0,0.02)' }}>
          <div style={{ fontSize: '0.7rem', textTransform: 'uppercase', color: 'var(--text-mute)', marginBottom: '0.2rem' }}>
            Total Score
          </div>
          <div style={{ fontSize: '1.3rem', fontWeight: 800, color: 'var(--text-h)' }}>
            {totalScore.toFixed(1)}
          </div>
        </div>
        <div className="card" style={{ padding: '0.75rem 0.5rem', textAlign: 'center', background: 'rgba(0,0,0,0.02)' }}>
          <div style={{ fontSize: '0.7rem', textTransform: 'uppercase', color: 'var(--text-mute)', marginBottom: '0.2rem' }}>
            Average
          </div>
          <div style={{ fontSize: '1.3rem', fontWeight: 800, color: 'var(--accent)' }}>
            {avgScore.toFixed(2)}
          </div>
        </div>
        <div className="card" style={{ padding: '0.75rem 0.5rem', textAlign: 'center', background: 'rgba(0,0,0,0.02)' }}>
          <div style={{ fontSize: '0.7rem', textTransform: 'uppercase', color: 'var(--text-mute)', marginBottom: '0.2rem' }}>
            Trimmed Avg
          </div>
          <div style={{ fontSize: '1.3rem', fontWeight: 800, color: '#16a34a' }}>
            {trimmedAvg.toFixed(2)}
          </div>
        </div>
      </div>

      <h4 style={{ margin: '0 0 0.75rem 0', fontSize: '0.95rem', fontWeight: 700 }}>Speech Timeline</h4>
      {speeches.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '2rem 0', color: 'var(--text-mute)', fontSize: '0.85rem' }}>
          No speeches recorded for this speaker yet.
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
          {speeches.map((sp: any, idx: number) => (
            <div
              key={idx}
              style={{
                border: '1px solid var(--border)',
                borderRadius: '8px',
                padding: '0.85rem 1rem',
                background: sp.score != null ? '#fff' : 'rgba(0,0,0,0.01)',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
              }}
            >
              <div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.2rem' }}>
                  <span style={{ fontWeight: 700, fontSize: '0.9rem', color: 'var(--text-h)' }}>
                    {sp.round_name}
                  </span>
                  <span className="badge badge-info" style={{ textTransform: 'uppercase', fontSize: '0.7rem' }}>
                    {sp.side}
                  </span>
                  {sp.is_reply && <span className="badge badge-warning" style={{ fontSize: '0.7rem' }}>Reply</span>}
                </div>
                <div style={{ fontSize: '0.8rem', color: 'var(--text-mute)', display: 'flex', gap: '0.6rem' }}>
                  <span>Venue: {sp.venue}</span>
                  {sp.team_points != null && <span>Team Points: {sp.team_points}</span>}
                </div>
              </div>

              <div>
                {sp.score != null ? (
                  <div style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--accent)', textAlign: 'right' }}>
                    {sp.score.toFixed(1)}
                  </div>
                ) : (
                  <span style={{ fontSize: '0.75rem', color: 'var(--text-mute)' }}>Pending</span>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function AdjudicatorTrajectoryView({ data }: { data: any }) {
  const debates: any[] = data?.debates || [];
  const adj = data?.adjudicator;

  const chairsCount = debates.filter(d => d.role === 'chair').length;
  const panelsCount = debates.filter(d => d.role === 'panel').length;
  const traineesCount = debates.filter(d => d.role === 'trainee').length;

  return (
    <div>
      <div className="grid grid-cols-4" style={{ gap: '0.75rem', marginBottom: '1.5rem' }}>
        <div className="card" style={{ padding: '0.75rem 0.5rem', textAlign: 'center', background: 'rgba(0,0,0,0.02)' }}>
          <div style={{ fontSize: '0.7rem', textTransform: 'uppercase', color: 'var(--text-mute)', marginBottom: '0.2rem' }}>
            Debates Judged
          </div>
          <div style={{ fontSize: '1.3rem', fontWeight: 800, color: 'var(--text-h)' }}>
            {debates.length}
          </div>
        </div>
        <div className="card" style={{ padding: '0.75rem 0.5rem', textAlign: 'center', background: 'rgba(0,0,0,0.02)' }}>
          <div style={{ fontSize: '0.7rem', textTransform: 'uppercase', color: 'var(--text-mute)', marginBottom: '0.2rem' }}>
            Chairs / Panels
          </div>
          <div style={{ fontSize: '1.3rem', fontWeight: 800, color: 'var(--accent)' }}>
            {chairsCount} <span style={{ fontSize: '0.9rem', color: 'var(--text-mute)', fontWeight: 500 }}>/ {panelsCount}</span>
          </div>
        </div>
        <div className="card" style={{ padding: '0.75rem 0.5rem', textAlign: 'center', background: 'rgba(0,0,0,0.02)' }}>
          <div style={{ fontSize: '0.7rem', textTransform: 'uppercase', color: 'var(--text-mute)', marginBottom: '0.2rem' }}>
            Trainees
          </div>
          <div style={{ fontSize: '1.3rem', fontWeight: 800, color: 'var(--text-h)' }}>
            {traineesCount}
          </div>
        </div>
        <div className="card" style={{ padding: '0.75rem 0.5rem', textAlign: 'center', background: 'rgba(0,0,0,0.02)' }}>
          <div style={{ fontSize: '0.7rem', textTransform: 'uppercase', color: 'var(--text-mute)', marginBottom: '0.2rem' }}>
            Test Score
          </div>
          <div style={{ fontSize: '1.3rem', fontWeight: 800, color: '#16a34a' }}>
            {adj?.test_score?.toFixed(1) || '0.0'}
          </div>
        </div>
      </div>

      <h4 style={{ margin: '0 0 0.75rem 0', fontSize: '0.95rem', fontWeight: 700 }}>Judging Timeline</h4>
      {debates.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '2rem 0', color: 'var(--text-mute)', fontSize: '0.85rem' }}>
          No judging assignments recorded for this adjudicator yet.
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
          {debates.map((d: any, idx: number) => (
            <div
              key={idx}
              style={{
                border: '1px solid var(--border)',
                borderRadius: '8px',
                padding: '0.85rem 1rem',
                background: '#fff',
              }}
            >
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.4rem' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <span style={{ fontWeight: 700, fontSize: '0.9rem', color: 'var(--text-h)' }}>
                    {d.round_name}
                  </span>
                  <span className="badge badge-info" style={{ textTransform: 'uppercase', fontSize: '0.7rem' }}>
                    {d.role}
                  </span>
                  <span style={{ fontSize: '0.8rem', color: 'var(--text-mute)', display: 'inline-flex', alignItems: 'center', gap: '0.2rem' }}>
                    <MapPin size={12} /> {d.venue}
                  </span>
                </div>
                {d.ballot_status && (
                  <span className="badge badge-success" style={{ fontSize: '0.7rem', textTransform: 'capitalize' }}>
                    {d.ballot_status}
                  </span>
                )}
              </div>

              {d.teams?.length > 0 && (
                <div style={{ fontSize: '0.8rem', color: 'var(--text-mute)', marginBottom: '0.25rem' }}>
                  <strong>Teams: </strong>
                  {d.teams.map((t: any) => `${t.team_name} (${t.side})`).join(' • ')}
                </div>
              )}

              {d.co_adjudicators?.length > 0 && (
                <div style={{ fontSize: '0.78rem', color: 'var(--text-mute)' }}>
                  <strong>Co-panelists: </strong>
                  {d.co_adjudicators.map((c: any) => `${c.name} (${c.role})`).join(', ')}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
