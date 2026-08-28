import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Plus, FileSpreadsheet } from 'lucide-react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from '@tanstack/react-form';
import { fetchAPI } from '../lib/api';

export default function LandingPage() {
  const [mode, setMode] = useState<'create' | 'upload'>('create');
  const [error, setError] = useState('');
  const [uploadError, setUploadError] = useState('');
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  // Query tournaments list
  const { data: tournaments = [] } = useQuery<any[]>({
    queryKey: ['tournaments'],
    queryFn: () => fetchAPI('/api/tournaments'),
  });

  // Mutations
  const createMutation = useMutation({
    mutationFn: (body: { name: string; slug: string }) => fetchAPI('/api/tournaments', 'POST', body),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['tournaments'] });
      navigate(`/t/${data.slug}/admin`);
    },
  });

  const uploadMutation = useMutation({
    mutationFn: (body: FormData) => fetchAPI('/api/archive/upload', 'POST', body, true),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['tournaments'] });
      navigate(`/t/${data.slug}/admin`);
    },
  });

  // Forms
  const createForm = useForm({
    defaultValues: {
      name: '',
      slug: '',
    },
    onSubmit: async ({ value }) => {
      setError('');
      try {
        await createMutation.mutateAsync(value);
      } catch (err: any) {
        setError(err.message);
      }
    },
  });

  const uploadForm = useForm({
    defaultValues: {
      name: '',
      slug: '',
      file: null as File | null,
    },
    onSubmit: async ({ value }) => {
      setUploadError('');
      if (!value.file) {
        setUploadError("Please select a database file to upload");
        return;
      }
      const formData = new FormData();
      formData.append('file', value.file);
      formData.append('name', value.name);
      formData.append('slug', value.slug);
      try {
        await uploadMutation.mutateAsync(formData);
      } catch (err: any) {
        setUploadError(err.message);
      }
    },
  });

  return (
    <div className="container" style={{ maxWidth: '680px', marginTop: '4rem' }}>
      <div style={{ textAlign: 'center', marginBottom: '2.5rem' }}>
        <h1 style={{ fontSize: '3.2rem', margin: '0 0 0.5rem 0', fontWeight: '700', letterSpacing: '-0.03em', color: 'var(--text-h)' }}>
          Yellow
        </h1>
        <p style={{ fontSize: '1.1rem', color: 'var(--text)' }}>Portable, Offline-First Tournament Tabbing Software</p>
      </div>

      <div className="tabs" style={{ marginBottom: '1.5rem', justifyContent: 'center' }}>
        <button className={`tab-btn ${mode === 'create' ? 'active' : ''}`} onClick={() => setMode('create')}>
          Create Tournament
        </button>
        <button className={`tab-btn ${mode === 'upload' ? 'active' : ''}`} onClick={() => setMode('upload')}>
          Upload Archive (.db)
        </button>
      </div>

      {mode === 'create' ? (
        <div className="card" style={{ marginBottom: '2rem' }}>
          <h2>Initialize New Tournament</h2>
          {error && <div style={{ color: 'var(--danger)', marginBottom: '1rem', fontSize: '0.9rem' }}>{error}</div>}
          <form onSubmit={(e) => { e.preventDefault(); e.stopPropagation(); createForm.handleSubmit(); }}>
            <createForm.Field
              name="name"
              children={(field) => (
                <div className="form-group">
                  <label className="label">Tournament Name</label>
                  <input
                    type="text"
                    className="input"
                    placeholder="e.g. World Universities Debate Championship"
                    value={field.state.value}
                    onChange={e => {
                      field.handleChange(e.target.value);
                      createForm.setFieldValue('slug', e.target.value.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)/g, ''));
                    }}
                  />
                </div>
              )}
            />
            <createForm.Field
              name="slug"
              children={(field) => (
                <div className="form-group">
                  <label className="label">URL Slug</label>
                  <input
                    type="text"
                    className="input"
                    placeholder="e.g. wudc-2026"
                    value={field.state.value}
                    onChange={e => field.handleChange(e.target.value.toLowerCase())}
                  />
                </div>
              )}
            />
            <button type="submit" className="btn btn-primary" style={{ width: '100%' }}>
              <Plus size={18} /> Initialize Tournament
            </button>
          </form>
        </div>
      ) : (
        <div className="card" style={{ marginBottom: '2rem' }}>
          <h2>Upload Tournament DB</h2>
          <p style={{ fontSize: '0.85rem', color: 'var(--text-mute)', marginBottom: '1.25rem' }}>Upload a local tournament SQLite database file to index it as a permanent public record.</p>
          {uploadError && <div style={{ color: 'var(--danger)', marginBottom: '1rem', fontSize: '0.9rem' }}>{uploadError}</div>}
          <form onSubmit={(e) => { e.preventDefault(); e.stopPropagation(); uploadForm.handleSubmit(); }}>
            <uploadForm.Field
              name="name"
              children={(field) => (
                <div className="form-group">
                  <label className="label">Tournament Name</label>
                  <input
                    type="text"
                    className="input"
                    placeholder="e.g. Oxford Union IV 2026"
                    value={field.state.value}
                    onChange={e => {
                      field.handleChange(e.target.value);
                      uploadForm.setFieldValue('slug', e.target.value.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)/g, ''));
                    }}
                  />
                </div>
              )}
            />
            <uploadForm.Field
              name="slug"
              children={(field) => (
                <div className="form-group">
                  <label className="label">URL Slug</label>
                  <input
                    type="text"
                    className="input"
                    placeholder="e.g. oxford-iv-2026"
                    value={field.state.value}
                    onChange={e => field.handleChange(e.target.value.toLowerCase())}
                  />
                </div>
              )}
            />
            <uploadForm.Field
              name="file"
              children={(field) => (
                <div className="form-group">
                  <label className="label">Select Tournament File (.db)</label>
                  <input
                    type="file"
                    accept=".db"
                    className="input"
                    onChange={e => field.handleChange(e.target.files?.[0] || null)}
                  />
                </div>
              )}
            />
            <button type="submit" className="btn btn-primary" style={{ width: '100%' }}>
              <FileSpreadsheet size={18} /> Upload and Register Archive
            </button>
          </form>
        </div>
      )}

      {tournaments.length > 0 && (
        <div className="card">
          <h2>Tournament Records</h2>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.85rem', marginTop: '1rem' }}>
            {tournaments.map((t: any) => (
              <div
                key={t.slug}
                style={{
                  display: 'flex',
                  flexDirection: 'column',
                  gap: '0.6rem',
                  padding: '1rem 1.25rem',
                  background: 'rgba(0,0,0,0.01)',
                  border: '1px solid var(--border)',
                  borderRadius: '8px',
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <Link
                    to={`/t/${t.slug}/admin`}
                    style={{ fontWeight: 700, fontSize: '1.05rem', color: 'var(--text-h)', textDecoration: 'none' }}
                  >
                    {t.name}
                  </Link>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    {t.is_archived ? (
                      <span className="badge badge-success" style={{ fontSize: '0.7rem' }}>Archived</span>
                    ) : (
                      <span className="badge badge-info" style={{ fontSize: '0.7rem' }}>Active</span>
                    )}
                    <span style={{ fontSize: '0.75rem', color: 'var(--text-mute)' }}>/{t.slug}</span>
                  </div>
                </div>

                <div style={{ display: 'flex', gap: '0.4rem', flexWrap: 'wrap' }}>
                  <Link to={`/t/${t.slug}/standings`} className="btn btn-secondary" style={{ padding: '3px 8px', fontSize: '0.75rem' }}>
                    Standings
                  </Link>
                  <Link to={`/t/${t.slug}/results`} className="btn btn-secondary" style={{ padding: '3px 8px', fontSize: '0.75rem' }}>
                    Results
                  </Link>
                  <Link to={`/t/${t.slug}/draw`} className="btn btn-secondary" style={{ padding: '3px 8px', fontSize: '0.75rem' }}>
                    Draw
                  </Link>
                  <Link to={`/t/${t.slug}/motions`} className="btn btn-secondary" style={{ padding: '3px 8px', fontSize: '0.75rem' }}>
                    Motions
                  </Link>
                  <Link to={`/t/${t.slug}/admin`} className="btn btn-primary" style={{ padding: '3px 10px', fontSize: '0.75rem', marginLeft: 'auto' }}>
                    Admin Workspace →
                  </Link>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
