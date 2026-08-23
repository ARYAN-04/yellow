import { useState, useEffect } from 'react';
import { Link, useParams } from 'react-router-dom';
import { fetchAPI } from '../lib/api';
import FeedbackSection from '../components/FeedbackSection';

export default function ParticipantDashboard() {
  const { token } = useParams<{ token: string }>();
  const [tokenInfo, setTokenInfo] = useState<any>(null);
  const [debates, setDebates] = useState<any[]>([]);
  const [standings, setStandings] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    setLoading(true);
    fetchAPI(`/api/token/${token}`)
      .then(info => {
        setTokenInfo(info);
        const loadDebates = fetchAPI(`/api/token/${token}/debates`);
        const loadStandings = fetchAPI(`/api/t/${info.slug}/standings`).catch(() => []);
        return Promise.all([loadDebates, loadStandings, info]);
      })
      .then(([d, s]) => {
        setDebates(d || []);
        setStandings(s || []);
        setLoading(false);
      })
      .catch(err => {
        setError(err.message);
        setLoading(false);
      });
  }, [token]);

  if (loading) {
    return <div className="container" style={{ textAlign: 'center', marginTop: '6rem' }}>Loading Portal...</div>;
  }

  if (error) {
    return (
      <div className="container" style={{ maxWidth: '420px', marginTop: '6rem' }}>
        <div className="card" style={{ textAlign: 'center', borderColor: 'var(--danger)' }}>
          <h2 style={{ color: 'var(--danger)' }}>Access Error</h2>
          <p>{error}</p>
          <div style={{ marginTop: '1.5rem' }}>
            <Link to="/" className="btn btn-secondary">Go Home</Link>
          </div>
        </div>
      </div>
    );
  }

  const myStanding = standings.find((team: any) => team.team_id === tokenInfo?.owner_id);
  const myRank = standings.findIndex((team: any) => team.team_id === tokenInfo?.owner_id) + 1;

  return (
    <div className="container" style={{ maxWidth: '720px' }}>
      <div className="card" style={{ marginBottom: '2rem', textAlign: 'center' }}>
        <span className="badge badge-info" style={{ marginBottom: '0.75rem', textTransform: 'uppercase' }}>Debater Portal</span>
        <h2 style={{ margin: '0 0 0.25rem 0' }}>{tokenInfo.owner_name}</h2>
        <p style={{ color: 'var(--text-mute)', fontSize: '0.85rem' }}>Tournament: {tokenInfo.slug} | Scoped Access Link</p>
      </div>

      {myStanding && (
        <div className="card" style={{ marginBottom: '2rem', textAlign: 'center' }}>
          <h3>Cumulative Standing</h3>
          <div className="grid grid-cols-3" style={{ marginTop: '1rem' }}>
            <div style={{ borderRight: '1px solid var(--border)' }}>
              <div style={{ fontSize: '0.75rem', textTransform: 'uppercase', color: 'var(--text-mute)', marginBottom: '0.25rem' }}>Rank</div>
              <div style={{ fontSize: '1.5rem', fontWeight: '700', color: 'var(--accent)' }}>#{myRank}</div>
            </div>
            <div style={{ borderRight: '1px solid var(--border)' }}>
              <div style={{ fontSize: '0.75rem', textTransform: 'uppercase', color: 'var(--text-mute)', marginBottom: '0.25rem' }}>Wins / Points</div>
              <div style={{ fontSize: '1.5rem', fontWeight: '700', color: 'var(--accent)' }}>{myStanding.points}</div>
            </div>
            <div>
              <div style={{ fontSize: '0.75rem', textTransform: 'uppercase', color: 'var(--text-mute)', marginBottom: '0.25rem' }}>Speaker Points</div>
              <div style={{ fontSize: '1.5rem', fontWeight: '700', color: 'var(--accent)' }}>{myStanding.speaker_points.toFixed(1)}</div>
            </div>
          </div>
        </div>
      )}

      <div className="card">
        <h3>Match Schedule</h3>
        {debates.length === 0 ? (
          <div style={{ padding: '2rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>
            No matches released for your team yet.
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            {debates.map(d => (
              <div key={d.id} style={{ padding: '1rem', background: 'rgba(0,0,0,0.01)', border: '1px solid var(--border)', borderRadius: '8px' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid var(--border)', paddingBottom: '0.5rem', marginBottom: '0.75rem' }}>
                  <span style={{ fontWeight: '600', color: 'var(--text-h)' }}>{d.round_name}</span>
                  <span style={{ fontSize: '0.85rem', color: 'var(--accent)', fontWeight: '500' }}>Venue: {d.venue}</span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
                  <span style={{ fontSize: '0.85rem' }}>Your Position:</span>
                  <span className="badge badge-info" style={{ textTransform: 'uppercase' }}>{d.side}</span>
                </div>
                <div style={{ fontSize: '0.8rem', color: 'var(--text-mute)', marginBottom: d.points !== undefined ? '0.5rem' : '0' }}>
                  Room Opponents: {d.teams.filter((t: any) => t.team_id !== tokenInfo.owner_id).map((t: any) => `${t.team_name} (${t.side})`).join(', ')}
                </div>
                {d.points !== undefined && (
                  <div style={{ marginTop: '0.75rem', paddingTop: '0.75rem', borderTop: '1px dashed var(--border)', display: 'flex', justifyContent: 'space-between', fontSize: '0.85rem' }}>
                    <span style={{ color: 'var(--text-mute)' }}>Ballot Result:</span>
                    <span style={{ fontWeight: '600', color: 'var(--success)' }}>
                      {d.points} Points (Ranks) | {d.speaker_points?.toFixed(1)} Speaker Pts
                    </span>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      <FeedbackSection token={token!} />
    </div>
  );
}
