import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { fetchAPI } from '../../lib/api';

const categoryTabs = [
  { key: '', label: 'Open' },
  { key: 'novice', label: 'Novice' },
  { key: 'esl', label: 'ESL' },
  { key: 'efl', label: 'EFL' },
];

export default function Standings() {
  const { slug } = useParams<{ slug: string }>();
  const [standings, setStandings] = useState<any[]>([]);
  const [category, setCategory] = useState('');

  useEffect(() => {
    const qs = category ? `?category=${category}` : '';
    fetchAPI(`/api/t/${slug}/standings${qs}`).then(d => setStandings(d || [])).catch(console.error);
  }, [slug, category]);

  return (
    <div className="card">
      <h3>Cumulative Team Standings</h3>
      <div className="tabs" style={{ marginBottom: '1rem' }}>
        {categoryTabs.map(t => (
          <button key={t.key} className={`tab-btn ${category === t.key ? 'active' : ''}`} onClick={() => setCategory(t.key)}>
            {t.label}
          </button>
        ))}
      </div>
      <div className="table-wrapper">
        <table className="table">
          <thead>
            <tr>
              <th>Rank</th>
              <th>Team Name</th>
              <th>Institution</th>
              <th>Wins/Points</th>
              <th>Speaker Points</th>
              <th>Margin</th>
            </tr>
          </thead>
          <tbody>
            {standings.map((team, idx) => (
              <tr key={team.team_id}>
                <td>{idx + 1}</td>
                <td style={{ fontWeight: '500', color: 'var(--text-h)' }}>{team.team_name}</td>
                <td>{team.institution_code || 'Independent'}</td>
                <td style={{ fontWeight: '600', color: 'var(--accent)' }}>{team.points}</td>
                <td>{team.speaker_points.toFixed(1)}</td>
                <td>{(team.margin || 0).toFixed(1)}</td>
              </tr>
            ))}
            {standings.length === 0 && (
              <tr>
                <td colSpan={6} style={{ textAlign: 'center', padding: '2rem 0', color: 'var(--text-mute)' }}>
                  No results logged yet.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
