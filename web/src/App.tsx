import React, { useState, useEffect } from 'react';
import { 
  BrowserRouter, 
  Routes, 
  Route, 
  Link, 
  useParams, 
  useNavigate 
} from 'react-router-dom';
import { 
  Plus, 
  Users, 
  Calendar, 
  FileSpreadsheet, 
  Lock, 
  LogOut, 
  TrendingUp,
  MapPin
} from 'lucide-react';
import { useQuery, useMutation, useQueryClient, QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useForm } from '@tanstack/react-form';

// API request helper
async function fetchAPI(url: string, method = 'GET', body: any = null, isMultipart = false) {
  const options: RequestInit = { method };
  if (body) {
    if (isMultipart) {
      options.body = body;
    } else {
      options.headers = { 'Content-Type': 'application/json' };
      options.body = JSON.stringify(body);
    }
  }
  
  const res = await fetch(url, options);
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error || `Request failed with status ${res.status}`);
  }
  if (res.status === 244 || res.status === 204) {
    return null;
  }
  return res.json();
}

// --- LANDING PAGE ---
function LandingPage() {
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
          GoTabs
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
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', marginTop: '1rem' }}>
            {tournaments.map((t: any) => (
              <Link 
                key={t.slug} 
                to={`/t/${t.slug}/admin`}
                style={{ 
                  display: 'flex', 
                  justifyContent: 'space-between', 
                  alignItems: 'center',
                  padding: '1rem',
                  background: 'rgba(0,0,0,0.02)',
                  border: '1px solid var(--border)',
                  borderRadius: '8px',
                  color: 'var(--text-h)'
                }}
              >
                <span>{t.name}</span>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  {t.is_archived ? (
                    <span className="badge badge-success" style={{ fontSize: '0.7rem' }}>Archived</span>
                  ) : (
                    <span className="badge badge-info" style={{ fontSize: '0.7rem' }}>Active</span>
                  )}
                  <span style={{ fontSize: '0.8rem', color: 'var(--text-mute)' }}>/t/{t.slug}/admin</span>
                </div>
              </Link>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// --- ADMIN DASHBOARD ---
function AdminDashboard() {
  const { slug } = useParams<{ slug: string }>();
  const [isAdmin, setIsAdmin] = useState(false);
  const [password, setPassword] = useState('');
  const [loginError, setLoginError] = useState('');
  const [activeTab, setActiveTab] = useState<'setup' | 'rounds' | 'ballots' | 'standings'>('setup');

  const [tournamentInfo, setTournamentInfo] = useState<any>(null);

  // Setup tab state
  const [institutions, setInstitutions] = useState<any[]>([]);
  const [teams, setTeams] = useState<any[]>([]);
  const [adjudicators, setAdjudicators] = useState<any[]>([]);
  
  // Setup forms
  const [instName, setInstName] = useState('');
  const [instCode, setInstCode] = useState('');
  const [teamName, setTeamName] = useState('');
  const [teamCode, setTeamCode] = useState('');
  const [teamInst, setTeamInst] = useState('');
  const [teamSp1, setTeamSp1] = useState('');
  const [teamSp2, setTeamSp2] = useState('');
  const [adjName, setAdjName] = useState('');
  const [adjInst, setAdjInst] = useState('');
  const [adjScore, setAdjScore] = useState(5.0);

  // Rounds state
  const [rounds, setRounds] = useState<any[]>([]);
  const [newRoundName, setNewRoundName] = useState('');
  const [newRoundSeq, setNewRoundSeq] = useState(1);
  const [newRoundStage, setNewRoundStage] = useState('preliminary');
  const [selectedRoundDraw, setSelectedRoundDraw] = useState<any[]>([]);
  const [activeRoundID, setActiveRoundID] = useState<string | null>(null);

  // Standings state
  const [standings, setStandings] = useState<any[]>([]);

  // Ballots state
  const [debateBallot, setDebateBallot] = useState<any | null>(null);
  const [ballotResults, setBallotResults] = useState<Record<string, { points: number, speakerPoints: number }>>({});

  // Check login and archive status
  useEffect(() => {
    fetchAPI('/api/tournaments')
      .then(list => {
        const current = list.find((t: any) => t.slug === slug);
        setTournamentInfo(current);
        if (current && current.is_archived) {
          setIsAdmin(true); // Bypass login screen for archived public views
        } else {
          // Check if authenticated
          fetchAPI(`/api/t/${slug}/institutions`)
            .then(() => setIsAdmin(true))
            .catch(() => setIsAdmin(false));
        }
      })
      .catch(() => {
        // Fallback check
        fetchAPI(`/api/t/${slug}/institutions`)
          .then(() => setIsAdmin(true))
          .catch(() => setIsAdmin(false));
      });
  }, [slug]);

  // Load Setup Data
  useEffect(() => {
    if (isAdmin) {
      loadSetupData();
      loadRounds();
      loadStandings();
    }
  }, [isAdmin, slug]);

  const loadSetupData = () => {
    fetchAPI(`/api/t/${slug}/institutions`).then(d => setInstitutions(d || [])).catch(console.error);
    fetchAPI(`/api/t/${slug}/teams`).then(d => setTeams(d || [])).catch(console.error);
    fetchAPI(`/api/t/${slug}/adjudicators`).then(d => setAdjudicators(d || [])).catch(console.error);
  };

  const loadRounds = () => {
    fetchAPI(`/api/t/${slug}/rounds`).then(r => {
      setRounds(r || []);
      if (r && r.length > 0) {
        setNewRoundSeq(r.length + 1);
      }
    }).catch(console.error);
  };

  const loadStandings = () => {
    fetchAPI(`/api/t/${slug}/standings`).then(d => setStandings(d || [])).catch(console.error);
  };

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoginError('');
    try {
      await fetchAPI('/api/login', 'POST', { password });
      setIsAdmin(true);
    } catch (err: any) {
      setLoginError(err.message);
    }
  };

  // --- CRUD Actions ---
  const handleCreateInst = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await fetchAPI(`/api/t/${slug}/institutions`, 'POST', { name: instName, code: instCode });
      setInstName('');
      setInstCode('');
      loadSetupData();
    } catch (err: any) { alert(err.message); }
  };

  const handleDeleteInst = async (id: string) => {
    if (confirm("Delete this institution?")) {
      try {
        await fetchAPI(`/api/t/${slug}/institutions/${id}`, 'DELETE');
        loadSetupData();
      } catch (err: any) { alert(err.message); }
    }
  };

  const handleCreateTeam = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await fetchAPI(`/api/t/${slug}/teams`, 'POST', {
        name: teamName,
        code: teamCode,
        institution_id: teamInst,
        speakers: [{ name: teamSp1 }, { name: teamSp2 }]
      });
      setTeamName('');
      setTeamCode('');
      setTeamSp1('');
      setTeamSp2('');
      loadSetupData();
      loadStandings();
    } catch (err: any) { alert(err.message); }
  };

  const handleDeleteTeam = async (id: string) => {
    if (confirm("Delete this team?")) {
      try {
        await fetchAPI(`/api/t/${slug}/teams/${id}`, 'DELETE');
        loadSetupData();
        loadStandings();
      } catch (err: any) { alert(err.message); }
    }
  };

  const handleCreateAdj = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await fetchAPI(`/api/t/${slug}/adjudicators`, 'POST', {
        name: adjName,
        institution_id: adjInst,
        test_score: Number(adjScore)
      });
      setAdjName('');
      setAdjScore(5.0);
      loadSetupData();
    } catch (err: any) { alert(err.message); }
  };

  const handleDeleteAdj = async (id: string) => {
    if (confirm("Delete this adjudicator?")) {
      try {
        await fetchAPI(`/api/t/${slug}/adjudicators/${id}`, 'DELETE');
        loadSetupData();
      } catch (err: any) { alert(err.message); }
    }
  };

  // --- CSV Bulk Uploads ---
  const handleCSVUpload = async (e: React.ChangeEvent<HTMLInputElement>, type: 'institutions' | 'teams' | 'adjudicators') => {
    const file = e.target.files?.[0];
    if (!file) return;

    const formData = new FormData();
    formData.append('file', file);

    try {
      await fetchAPI(`/api/t/${slug}/import/${type}`, 'POST', formData, true);
      loadSetupData();
      loadStandings();
      alert(`Imported ${type} successfully!`);
    } catch (err: any) {
      alert("CSV Import failed: " + err.message);
    }
  };

  // --- Round Actions ---
  const handleCreateRound = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await fetchAPI(`/api/t/${slug}/rounds`, 'POST', {
        name: newRoundName,
        seq: Number(newRoundSeq),
        stage: newRoundStage
      });
      setNewRoundName('');
      loadRounds();
    } catch (err: any) { alert(err.message); }
  };

  const handleViewRoundDraw = async (roundID: string) => {
    setActiveRoundID(roundID);
    try {
      const draw = await fetchAPI(`/api/t/${slug}/rounds/${roundID}/draw`);
      setSelectedRoundDraw(draw || []);
    } catch (err: any) {
      setSelectedRoundDraw([]);
    }
  };

  const handleGenerateRoundDraw = async (roundID: string) => {
    try {
      await fetchAPI(`/api/t/${slug}/rounds/${roundID}/draw`, 'POST');
      handleViewRoundDraw(roundID);
    } catch (err: any) {
      alert("Failed to generate draw: " + err.message);
    }
  };

  // --- Ballot Submission ---
  const openBallotForm = (debate: any) => {
    setDebateBallot(debate);
    const initialResults: Record<string, { points: number, speakerPoints: number }> = {};
    debate.teams.forEach((t: any) => {
      initialResults[t.team_id] = { points: 0, speakerPoints: 150 };
    });
    setBallotResults(initialResults);
  };

  const submitBallot = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!debateBallot) return;

    const resultsPayload = Object.keys(ballotResults).map(tid => ({
      team_id: tid,
      points: ballotResults[tid].points,
      speaker_points: ballotResults[tid].speakerPoints
    }));

    try {
      const res = await fetchAPI(`/api/t/${slug}/debates/${debateBallot.id}/ballots`, 'POST', {
        submitter_type: 'organizer',
        submitter_id: 'admin',
        results: resultsPayload
      });
      
      await fetchAPI(`/api/t/${slug}/ballots/${res.id}/confirm`, 'POST');
      
      setDebateBallot(null);
      if (activeRoundID) {
        handleViewRoundDraw(activeRoundID);
      }
      loadStandings();
      alert("Ballot entered and confirmed!");
    } catch (err: any) {
      alert("Failed to submit ballot: " + err.message);
    }
  };

  const isReadOnly = tournamentInfo?.is_archived === 1 || tournamentInfo?.is_archived === true;

  if (!isAdmin) {
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

  return (
    <div>
      <div className="header">
        <div className="header-inner">
          <div className="logo-text">
            GoTabs {isReadOnly ? 'Record' : 'Admin'} <span style={{ color: 'var(--accent)', fontSize: '0.8rem', fontWeight: '500' }}>/{slug}</span>
          </div>
          <div style={{ display: 'flex', gap: '1rem' }}>
            <Link to="/" className="btn btn-secondary" style={{ padding: '0.4rem 0.8rem', fontSize: '0.8rem' }}>
              <LogOut size={14} /> Exit {isReadOnly ? 'View' : ''}
            </Link>
          </div>
        </div>
      </div>

      <div className="container">
        {isReadOnly && (
          <div className="badge badge-success" style={{ width: '100%', padding: '0.75rem', borderRadius: '8px', marginBottom: '1.5rem', fontSize: '0.9rem', justifyContent: 'center' }}>
            This tournament record is archived and read-only.
          </div>
        )}

        <div className="tabs">
          <button className={`tab-btn ${activeTab === 'setup' ? 'active' : ''}`} onClick={() => setActiveTab('setup')}>
            <Users size={16} /> Tournament Setup
          </button>
          <button className={`tab-btn ${activeTab === 'rounds' ? 'active' : ''}`} onClick={() => setActiveTab('rounds')}>
            <Calendar size={16} /> Rounds & Matchups
          </button>
          <button className={`tab-btn ${activeTab === 'standings' ? 'active' : ''}`} onClick={() => setActiveTab('standings')}>
            <TrendingUp size={16} /> Standings
          </button>
        </div>

        {/* --- TAB: SETUP --- */}
        {activeTab === 'setup' && (
          <div className="grid grid-cols-3">
            {/* Institutions column */}
            <div className="card">
              <h3>Institutions</h3>
              {!isReadOnly && (
                <>
                  <form onSubmit={handleCreateInst} style={{ marginBottom: '1.5rem' }}>
                    <div className="form-group">
                      <input type="text" className="input" placeholder="Name (e.g. Oxford)" value={instName} onChange={e => setInstName(e.target.value)} />
                    </div>
                    <div className="form-group">
                      <input type="text" className="input" placeholder="Code (e.g. OXF)" value={instCode} onChange={e => setInstCode(e.target.value)} />
                    </div>
                    <button type="submit" className="btn btn-primary" style={{ width: '100%' }}><Plus size={16} /> Add</button>
                  </form>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '1rem', borderTop: '1px solid var(--border)', paddingTop: '1rem' }}>
                    <FileSpreadsheet size={16} />
                    <span style={{ fontSize: '0.8rem' }}>CSV Bulk Import:</span>
                    <input type="file" accept=".csv" onChange={e => handleCSVUpload(e, 'institutions')} style={{ fontSize: '0.75rem', width: '100%' }} />
                  </div>
                </>
              )}
              <div style={{ maxHeight: '350px', overflowY: 'auto' }}>
                {institutions.map(i => (
                  <div key={i.id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.5rem 0', borderBottom: '1px solid var(--border)' }}>
                    <span>{i.name} ({i.code})</span>
                    {!isReadOnly && (
                      <button className="btn btn-danger" style={{ padding: '2px 6px', fontSize: '0.75rem' }} onClick={() => handleDeleteInst(i.id)}>Delete</button>
                    )}
                  </div>
                ))}
              </div>
            </div>

            {/* Teams column */}
            <div className="card">
              <h3>Teams</h3>
              {!isReadOnly && (
                <>
                  <form onSubmit={handleCreateTeam} style={{ marginBottom: '1.5rem' }}>
                    <div className="form-group">
                      <input type="text" className="input" placeholder="Team Name" value={teamName} onChange={e => setTeamName(e.target.value)} />
                    </div>
                    <div className="form-group">
                      <input type="text" className="input" placeholder="Short Code (e.g. A)" value={teamCode} onChange={e => setTeamCode(e.target.value)} />
                    </div>
                    <div className="form-group">
                      <select className="input select" value={teamInst} onChange={e => setTeamInst(e.target.value)}>
                        <option value="">Select Institution</option>
                        {institutions.map(i => <option key={i.id} value={i.id}>{i.name}</option>)}
                      </select>
                    </div>
                    <div className="form-group" style={{ display: 'flex', gap: '0.5rem' }}>
                      <input type="text" className="input" placeholder="Speaker 1" value={teamSp1} onChange={e => setTeamSp1(e.target.value)} />
                      <input type="text" className="input" placeholder="Speaker 2" value={teamSp2} onChange={e => setTeamSp2(e.target.value)} />
                    </div>
                    <button type="submit" className="btn btn-primary" style={{ width: '100%' }}><Plus size={16} /> Add Team</button>
                  </form>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '1rem', borderTop: '1px solid var(--border)', paddingTop: '1rem' }}>
                    <FileSpreadsheet size={16} />
                    <span style={{ fontSize: '0.8rem' }}>CSV Import:</span>
                    <input type="file" accept=".csv" onChange={e => handleCSVUpload(e, 'teams')} style={{ fontSize: '0.75rem', width: '100%' }} />
                  </div>
                </>
              )}
              <div style={{ maxHeight: '350px', overflowY: 'auto' }}>
                {teams.map(t => (
                  <div key={t.id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.5rem 0', borderBottom: '1px solid var(--border)' }}>
                    <div>
                      <div style={{ fontWeight: '500' }}>{t.name}</div>
                      <div style={{ fontSize: '0.75rem', color: 'var(--text-mute)' }}>Inst: {t.institution_code || 'None'}</div>
                    </div>
                    {!isReadOnly && (
                      <button className="btn btn-danger" style={{ padding: '2px 6px', fontSize: '0.75rem' }} onClick={() => handleDeleteTeam(t.id)}>Delete</button>
                    )}
                  </div>
                ))}
              </div>
            </div>

            {/* Adjudicators column */}
            <div className="card">
              <h3>Adjudicators</h3>
              {!isReadOnly && (
                <>
                  <form onSubmit={handleCreateAdj} style={{ marginBottom: '1.5rem' }}>
                    <div className="form-group">
                      <input type="text" className="input" placeholder="Judge Name" value={adjName} onChange={e => setAdjName(e.target.value)} />
                    </div>
                    <div className="form-group">
                      <select className="input select" value={adjInst} onChange={e => setAdjInst(e.target.value)}>
                        <option value="">Institution Connection</option>
                        {institutions.map(i => <option key={i.id} value={i.id}>{i.name}</option>)}
                      </select>
                    </div>
                    <div className="form-group">
                      <input type="number" step="0.1" className="input" placeholder="Judge Score (e.g. 5.0)" value={adjScore} onChange={e => setAdjScore(Number(e.target.value))} />
                    </div>
                    <button type="submit" className="btn btn-primary" style={{ width: '100%' }}><Plus size={16} /> Add Judge</button>
                  </form>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '1rem', borderTop: '1px solid var(--border)', paddingTop: '1rem' }}>
                    <FileSpreadsheet size={16} />
                    <span style={{ fontSize: '0.8rem' }}>CSV Import:</span>
                    <input type="file" accept=".csv" onChange={e => handleCSVUpload(e, 'adjudicators')} style={{ fontSize: '0.75rem', width: '100%' }} />
                  </div>
                </>
              )}
              <div style={{ maxHeight: '350px', overflowY: 'auto' }}>
                {adjudicators.map(a => (
                  <div key={a.id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.5rem 0', borderBottom: '1px solid var(--border)' }}>
                    <div>
                      <div style={{ fontWeight: '500' }}>{a.name}</div>
                      <div style={{ fontSize: '0.75rem', color: 'var(--text-mute)' }}>Score: {a.test_score} | Inst: {a.institution_code || 'None'}</div>
                    </div>
                    {!isReadOnly && (
                      <button className="btn btn-danger" style={{ padding: '2px 6px', fontSize: '0.75rem' }} onClick={() => handleDeleteAdj(a.id)}>Delete</button>
                    )}
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* --- TAB: ROUNDS --- */}
        {activeTab === 'rounds' && (
          <div className="grid grid-cols-3">
            {/* Rounds list card */}
            <div className="card" style={{ gridColumn: 'span 1' }}>
              {!isReadOnly && (
                <div style={{ marginBottom: '1.5rem', borderBottom: '1px solid var(--border)', paddingBottom: '1.5rem' }}>
                  <h3>Add Round</h3>
                  <form onSubmit={handleCreateRound}>
                    <div className="form-group">
                      <input type="text" className="input" placeholder="Round Name (e.g. Round 1)" value={newRoundName} onChange={e => setNewRoundName(e.target.value)} />
                    </div>
                    <div className="form-group">
                      <input type="number" className="input" placeholder="Sequence (e.g. 1)" value={newRoundSeq} onChange={e => setNewRoundSeq(Number(e.target.value))} />
                    </div>
                    <div className="form-group">
                      <select className="input select" value={newRoundStage} onChange={e => setNewRoundStage(e.target.value)}>
                        <option value="preliminary">Preliminary</option>
                        <option value="elimination">Elimination (Break)</option>
                      </select>
                    </div>
                    <button type="submit" className="btn btn-primary" style={{ width: '100%' }}><Plus size={16} /> Create Round</button>
                  </form>
                </div>
              )}
              
              <h3>Rounds List</h3>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                {rounds.map(r => (
                  <button 
                    key={r.id} 
                    className={`btn ${activeRoundID === r.id ? 'btn-primary' : 'btn-secondary'}`}
                    style={{ width: '100%', justifyContent: 'space-between' }}
                    onClick={() => handleViewRoundDraw(r.id)}
                  >
                    <span>{r.name}</span>
                    <span style={{ fontSize: '0.8rem', opacity: 0.8 }}>Seq: {r.seq}</span>
                  </button>
                ))}
              </div>
            </div>

            {/* Matchups list card */}
            <div className="card" style={{ gridColumn: 'span 2' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
                <h3>Round Matchups</h3>
                {activeRoundID && !isReadOnly && (
                  <button className="btn btn-primary" onClick={() => handleGenerateRoundDraw(activeRoundID)}>
                    Generate/Reset Draw
                  </button>
                )}
              </div>

              {activeRoundID && (
                (() => {
                  const currentRound = rounds.find(r => r.id === activeRoundID);
                  if (!currentRound) return null;
                  return (
                    <div style={{ display: 'flex', gap: '1.5rem', marginBottom: '1.5rem', paddingBottom: '1rem', borderBottom: '1px solid var(--border)', flexWrap: 'wrap', alignItems: 'center' }}>
                      <span style={{ fontWeight: '600', fontSize: '0.85rem', color: 'var(--text-h)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Round Controls:</span>
                      <label style={{ display: 'inline-flex', alignItems: 'center', gap: '0.4rem', cursor: isReadOnly ? 'not-allowed' : 'pointer', fontSize: '0.85rem', color: 'var(--text-h)' }}>
                        <input 
                          type="checkbox" 
                          disabled={isReadOnly}
                          checked={currentRound.draw_released}
                          onChange={async e => {
                            try {
                              await fetchAPI(`/api/t/${slug}/rounds/${activeRoundID}`, 'PUT', { draw_released: e.target.checked });
                              loadRounds();
                            } catch (err: any) { alert(err.message); }
                          }}
                        />
                        Draw Released
                      </label>
                      <label style={{ display: 'inline-flex', alignItems: 'center', gap: '0.4rem', cursor: isReadOnly ? 'not-allowed' : 'pointer', fontSize: '0.85rem', color: 'var(--text-h)' }}>
                        <input 
                          type="checkbox" 
                          disabled={isReadOnly}
                          checked={currentRound.silent}
                          onChange={async e => {
                            try {
                              await fetchAPI(`/api/t/${slug}/rounds/${activeRoundID}`, 'PUT', { silent: e.target.checked });
                              loadRounds();
                            } catch (err: any) { alert(err.message); }
                          }}
                        />
                        Silent Round
                      </label>
                      <label style={{ display: 'inline-flex', alignItems: 'center', gap: '0.4rem', cursor: isReadOnly ? 'not-allowed' : 'pointer', fontSize: '0.85rem', color: 'var(--text-h)' }}>
                        <input 
                          type="checkbox" 
                          disabled={isReadOnly}
                          checked={currentRound.results_released}
                          onChange={async e => {
                            try {
                              await fetchAPI(`/api/t/${slug}/rounds/${activeRoundID}`, 'PUT', { results_released: e.target.checked });
                              loadRounds();
                              loadStandings();
                            } catch (err: any) { alert(err.message); }
                          }}
                        />
                        Results Released
                      </label>
                    </div>
                  );
                })()
              )}

              {!activeRoundID ? (
                <div style={{ padding: '3rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>
                  Select a round from the left column to view matches.
                </div>
              ) : selectedRoundDraw.length === 0 ? (
                <div style={{ padding: '3rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>
                  No matches generated for this round yet.
                </div>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                  {selectedRoundDraw.map(d => (
                    <div key={d.id} className="card" style={{ padding: '1.25rem', background: 'rgba(0,0,0,0.01)', border: '1px solid var(--border)' }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '1px solid var(--border)', paddingBottom: '0.5rem', marginBottom: '0.75rem' }}>
                        <span style={{ fontWeight: '600', color: 'var(--text-h)', display: 'inline-flex', alignItems: 'center', gap: '0.4rem' }}>
                          <MapPin size={16} /> {d.venue}
                        </span>
                        {!isReadOnly && (
                          <button className="btn btn-secondary" style={{ padding: '4px 10px', fontSize: '0.75rem' }} onClick={() => openBallotForm(d)}>
                            Record Ballot
                          </button>
                        )}
                      </div>

                      {/* BP Grid representation */}
                      <div className="grid grid-cols-2" style={{ gap: '0.75rem' }}>
                        {d.teams.map((t: any) => (
                          <div key={t.team_id} style={{ display: 'flex', justifyContent: 'space-between', padding: '0.5rem', background: 'rgba(0,0,0,0.02)', borderRadius: '6px' }}>
                            <span style={{ fontSize: '0.85rem' }}>{t.team_name}</span>
                            <span className="badge badge-info" style={{ textTransform: 'uppercase' }}>{t.side}</span>
                          </div>
                        ))}
                      </div>

                      <div style={{ marginTop: '0.75rem', fontSize: '0.8rem', color: 'var(--text-mute)' }}>
                        Chair Panel: {d.adjudicators.filter((a: any) => a.role === 'chair').map((a: any) => a.adjudicator_name).join(', ') || 'No chair assigned'}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}

        {/* --- TAB: STANDINGS --- */}
        {activeTab === 'standings' && (
          <div className="card">
            <h3>Cumulative Team Standings</h3>
            <div className="table-wrapper">
              <table className="table">
                <thead>
                  <tr>
                    <th>Rank</th>
                    <th>Team Name</th>
                    <th>Institution</th>
                    <th>Wins/Points</th>
                    <th>Speaker Points</th>
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
                    </tr>
                  ))}
                  {standings.length === 0 && (
                    <tr>
                      <td colSpan={5} style={{ textAlign: 'center', padding: '2rem 0', color: 'var(--text-mute)' }}>
                        No results logged yet.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* --- BALLOT ENTRY MODAL --- */}
        {debateBallot && !isReadOnly && (
          <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, background: 'rgba(0,0,0,0.8)', zIndex: 1000, display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
            <div className="card" style={{ maxWidth: '500px', width: '90%', maxHeight: '90vh', overflowY: 'auto' }}>
              <h3>Enter Ballot: {debateBallot.venue}</h3>
              <form onSubmit={submitBallot}>
                {debateBallot.teams.map((t: any) => (
                  <div key={t.team_id} style={{ padding: '1rem', borderBottom: '1px solid var(--border)', marginBottom: '1rem' }}>
                    <div style={{ fontWeight: '600', color: 'var(--text-h)', marginBottom: '0.5rem' }}>
                      {t.team_name} ({t.side.toUpperCase()})
                    </div>
                    <div className="grid grid-cols-2" style={{ gap: '0.75rem' }}>
                      <div>
                        <label className="label">Wins / Rank Points</label>
                        <select 
                          className="input select"
                          value={ballotResults[t.team_id]?.points || 0}
                          onChange={e => setBallotResults({
                            ...ballotResults,
                            [t.team_id]: { ...ballotResults[t.team_id], points: Number(e.target.value) }
                          })}
                        >
                          <option value="3">3 Points (1st Place)</option>
                          <option value="2">2 Points (2nd Place)</option>
                          <option value="1">1 Point (3rd Place)</option>
                          <option value="0">0 Points (4th Place)</option>
                        </select>
                      </div>
                      <div>
                        <label className="label">Total Speaker Points</label>
                        <input 
                          type="number" 
                          step="0.5"
                          className="input" 
                          value={ballotResults[t.team_id]?.speakerPoints || 150}
                          onChange={e => setBallotResults({
                            ...ballotResults,
                            [t.team_id]: { ...ballotResults[t.team_id], speakerPoints: Number(e.target.value) }
                          })}
                        />
                      </div>
                    </div>
                  </div>
                ))}
                
                <div style={{ display: 'flex', gap: '1rem', marginTop: '1rem' }}>
                  <button type="submit" className="btn btn-primary" style={{ flex: 1 }}>Submit & Confirm</button>
                  <button type="button" className="btn btn-secondary" onClick={() => setDebateBallot(null)}>Cancel</button>
                </div>
              </form>
            </div>
          </div>
        )}

      </div>
    </div>
  );
}

// --- PARTICIPANT ROUTE ---
function ParticipantDashboard() {
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
    </div>
  );
}

// --- JUDGE ROUTE ---
function JudgeDashboard() {
  const { token } = useParams<{ token: string }>();
  const [tokenInfo, setTokenInfo] = useState<any>(null);
  const [debates, setDebates] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [selectedDebate, setSelectedDebate] = useState<any | null>(null);
  const [ballotResults, setBallotResults] = useState<Record<string, { points: number, speakerPoints: number }>>({});

  const loadData = () => {
    setLoading(true);
    fetchAPI(`/api/token/${token}`)
      .then(info => {
        setTokenInfo(info);
        return fetchAPI(`/api/token/${token}/debates`);
      })
      .then(d => {
        setDebates(d || []);
        setLoading(false);
      })
      .catch(err => {
        setError(err.message);
        setLoading(false);
      });
  };

  useEffect(() => {
    loadData();
  }, [token]);

  const openBallotForm = (debate: any) => {
    setSelectedDebate(debate);
    const initialResults: Record<string, { points: number, speakerPoints: number }> = {};
    debate.teams.forEach((t: any) => {
      initialResults[t.team_id] = { points: 0, speakerPoints: 150 };
    });
    setBallotResults(initialResults);
  };

  const handleBallotSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedDebate) return;

    const resultsPayload = Object.keys(ballotResults).map(tid => ({
      team_id: tid,
      points: ballotResults[tid].points,
      speaker_points: ballotResults[tid].speakerPoints
    }));

    try {
      await fetchAPI(`/api/token/${token}/debates/${selectedDebate.id}/ballots`, 'POST', {
        results: resultsPayload
      });
      setSelectedDebate(null);
      loadData();
      alert("Ballot submitted successfully! It is pending organizer confirmation.");
    } catch (err: any) {
      alert("Failed to submit ballot: " + err.message);
    }
  };

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

  return (
    <div className="container" style={{ maxWidth: '720px' }}>
      <div className="card" style={{ marginBottom: '2rem', textAlign: 'center' }}>
        <span className="badge badge-warning" style={{ marginBottom: '0.75rem', textTransform: 'uppercase' }}>Adjudicator Portal</span>
        <h2 style={{ margin: '0 0 0.25rem 0' }}>{tokenInfo.owner_name}</h2>
        <p style={{ color: 'var(--text-mute)', fontSize: '0.85rem' }}>Tournament: {tokenInfo.slug} | Scoped Access Link</p>
      </div>

      <div className="card">
        <h3>Your Assigned Debates</h3>
        {debates.length === 0 ? (
          <div style={{ padding: '2rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>
            No debates assigned to you in active rounds.
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
            {debates.map(d => (
              <div key={d.id} className="card" style={{ padding: '1.25rem', background: 'rgba(0,0,0,0.01)', border: '1px solid var(--border)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '1px solid var(--border)', paddingBottom: '0.5rem', marginBottom: '0.75rem' }}>
                  <span style={{ fontWeight: '600', color: 'var(--text-h)' }}>{d.round_name} — {d.venue}</span>
                  <span className="badge badge-info" style={{ textTransform: 'uppercase' }}>Role: {d.role}</span>
                </div>

                <div className="grid grid-cols-2" style={{ gap: '0.5rem', marginBottom: '1rem' }}>
                  {d.teams.map((t: any) => (
                    <div key={t.team_id} style={{ display: 'flex', justifyContent: 'space-between', padding: '0.4rem', background: 'rgba(0,0,0,0.02)', borderRadius: '6px', fontSize: '0.85rem' }}>
                      <span>{t.team_name}</span>
                      <span style={{ textTransform: 'uppercase', color: 'var(--text-mute)' }}>{t.side}</span>
                    </div>
                  ))}
                </div>

                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderTop: '1px solid var(--border)', paddingTop: '0.75rem' }}>
                  <span style={{ fontSize: '0.8rem', color: 'var(--text-mute)' }}>
                    Ballot status: <span style={{ fontWeight: '600', color: d.ballot_status === 'confirmed' ? 'var(--success)' : d.ballot_status === 'submitted' ? 'var(--warning)' : 'var(--text-mute)' }}>{d.ballot_status}</span>
                  </span>
                  {d.ballot_status === 'unsubmitted' && d.role === 'chair' && (
                    <button className="btn btn-primary" style={{ padding: '4px 10px', fontSize: '0.8rem' }} onClick={() => openBallotForm(d)}>
                      Enter Ballot
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {selectedDebate && (
        <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, background: 'rgba(0,0,0,0.8)', zIndex: 1000, display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
          <div className="card" style={{ maxWidth: '500px', width: '90%', maxHeight: '90vh', overflowY: 'auto' }}>
            <h3>Submit Ballot: {selectedDebate.round_name} ({selectedDebate.venue})</h3>
            <form onSubmit={handleBallotSubmit}>
              {selectedDebate.teams.map((t: any) => (
                <div key={t.team_id} style={{ padding: '1rem', borderBottom: '1px solid var(--border)', marginBottom: '1rem' }}>
                  <div style={{ fontWeight: '600', color: 'var(--text-h)', marginBottom: '0.5rem' }}>
                    {t.team_name} ({t.side.toUpperCase()})
                  </div>
                  <div className="grid grid-cols-2" style={{ gap: '0.75rem' }}>
                    <div>
                      <label className="label">Wins / Rank Points</label>
                      <select 
                        className="input select"
                        value={ballotResults[t.team_id]?.points || 0}
                        onChange={e => setBallotResults({
                          ...ballotResults,
                          [t.team_id]: { ...ballotResults[t.team_id], points: Number(e.target.value) }
                        })}
                      >
                        <option value="3">3 Points (1st Place)</option>
                        <option value="2">2 Points (2nd Place)</option>
                        <option value="1">1 Point (3rd Place)</option>
                        <option value="0">0 Points (4th Place)</option>
                      </select>
                    </div>
                    <div>
                      <label className="label">Total Speaker Points</label>
                      <input 
                        type="number" 
                        step="0.5"
                        className="input" 
                        value={ballotResults[t.team_id]?.speakerPoints || 150}
                        onChange={e => setBallotResults({
                          ...ballotResults,
                          [t.team_id]: { ...ballotResults[t.team_id], speakerPoints: Number(e.target.value) }
                        })}
                      />
                    </div>
                  </div>
                </div>
              ))}
              
              <div style={{ display: 'flex', gap: '1rem', marginTop: '1rem' }}>
                <button type="submit" className="btn btn-primary" style={{ flex: 1 }}>Submit Ballot</button>
                <button type="button" className="btn btn-secondary" onClick={() => setSelectedDebate(null)}>Cancel</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

const queryClient = new QueryClient();

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<LandingPage />} />
          <Route path="/t/:slug/admin" element={<AdminDashboard />} />
          <Route path="/p/:token" element={<ParticipantDashboard />} />
          <Route path="/j/:token" element={<JudgeDashboard />} />
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}

export default App;
