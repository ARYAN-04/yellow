import React, { useState } from 'react';
import { Lock } from 'lucide-react';
import { fetchAPI } from '../lib/api';

interface LoginPageProps {
  onSuccess: () => void;
}

export default function LoginPage({ onSuccess }: LoginPageProps) {
  const [password, setPassword] = useState('');
  const [loginError, setLoginError] = useState('');

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoginError('');
    try {
      await fetchAPI('/api/login', 'POST', { password });
      onSuccess();
    } catch (err: any) {
      setLoginError(err.message);
    }
  };

  return (
    <div className="container" style={{ maxWidth: '420px', marginTop: '6rem' }}>
      <div className="card" style={{ textAlign: 'center' }}>
        <div style={{ display: 'inline-flex', padding: '1rem', borderRadius: '50%', background: 'rgba(0, 0, 0, 0.05)', color: 'var(--accent)', marginBottom: '1.5rem' }}>
          <Lock size={32} />
        </div>
        <h2>Admin Authentication</h2>
        <p style={{ fontSize: '0.9rem', marginBottom: '1.5rem' }}>Access is password-gated for tournament organizers.</p>
        {loginError && <div style={{ color: 'var(--danger)', marginBottom: '1rem', fontSize: '0.85rem' }}>{loginError}</div>}
        <form onSubmit={handleLogin}>
          <div className="form-group">
            <label className="label">Organizer Password</label>
            <input
              type="password"
              className="input"
              placeholder="Enter password (default: admin)"
              value={password}
              onChange={e => setPassword(e.target.value)}
            />
          </div>
          <button type="submit" className="btn btn-primary" style={{ width: '100%' }}>
            Unlock Console
          </button>
        </form>
      </div>
    </div>
  );
}
