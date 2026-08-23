import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { CheckCircle2 } from 'lucide-react';
import { fetchAPI } from '../lib/api';

interface CheckinInfo {
  entity_type: string;
  entity_name: string;
  checked_in: boolean;
}

export default function CheckinPage() {
  const { token } = useParams<{ token: string }>();
  const [info, setInfo] = useState<CheckinInfo | null>(null);
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    fetchAPI(`/api/checkin/${token}`)
      .then(setInfo)
      .catch(err => setError(err.message));
  }, [token]);

  const checkIn = async () => {
    setSubmitting(true);
    try {
      const updated = await fetchAPI(`/api/checkin/${token}`, 'POST');
      setInfo(updated);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div style={{
      minHeight: '100vh',
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      gap: '1.5rem',
      padding: '2rem',
      textAlign: 'center',
    }}>
      <div style={{ fontSize: '0.85rem', fontWeight: 700, letterSpacing: '0.08em', color: 'var(--text-mute)' }}>GOTABS CHECK-IN</div>

      {error && (
        <>
          <p style={{ color: '#dc2626' }}>{error}</p>
          <Link to="/" className="btn btn-secondary">Go Home</Link>
        </>
      )}

      {!error && !info && <p style={{ color: 'var(--text-mute)' }}>Loading…</p>}

      {!error && info && (
        <>
          <h1 style={{ margin: 0 }}>{info.entity_name}</h1>
          <p style={{ margin: 0, color: 'var(--text-mute)', textTransform: 'capitalize' }}>{info.entity_type}</p>
          {info.checked_in ? (
            <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '0.75rem' }}>
              <span className="badge badge-success" style={{ fontSize: '1rem', padding: '0.6rem 1.4rem', gap: '0.5rem' }}>
                <CheckCircle2 size={18} /> Checked in
              </span>
              <p style={{ margin: 0, color: 'var(--text-mute)', fontSize: '0.85rem' }}>You're all set — see the tab room if your name is wrong.</p>
            </div>
          ) : (
            <button
              type="button"
              className="btn btn-primary"
              style={{ fontSize: '1.15rem', padding: '1rem 3rem', borderRadius: '12px' }}
              onClick={checkIn}
              disabled={submitting}
            >
              {submitting ? 'Checking in…' : 'Check In'}
            </button>
          )}
        </>
      )}
    </div>
  );
}
