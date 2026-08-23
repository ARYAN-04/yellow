import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Plus, FileSpreadsheet, ChevronUp, ChevronDown } from 'lucide-react';
import { useOutletContext } from 'react-router-dom';
import { fetchAPI, type AdminContext } from '../../lib/api';

const precedenceLabels: Record<string, string> = {
  points: 'Team Points',
  speaker_points: 'Speaker Points',
  margin: 'Margin',
};

export default function Setup() {
  const { slug } = useParams<{ slug: string }>();
  const { isReadOnly } = useOutletContext<AdminContext>();
  const [activeTab, setActiveTab] = useState<'institutions' | 'teams' | 'adjudicators' | 'config' | 'categories'>('institutions');

  const [institutions, setInstitutions] = useState<any[]>([]);
  const [teams, setTeams] = useState<any[]>([]);
  const [adjudicators, setAdjudicators] = useState<any[]>([]);
  const [breakCategories, setBreakCategories] = useState<any[]>([]);

  // Forms
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
  const [bcName, setBcName] = useState('');
  const [bcSeq, setBcSeq] = useState('0');
  const [bcSize, setBcSize] = useState('');
  const [bcBasePoints, setBcBasePoints] = useState('');
  const [editingBcId, setEditingBcId] = useState<string | null>(null);

  // Config
  const [cfgSides, setCfgSides] = useState('');
  const [cfgScoreMin, setCfgScoreMin] = useState(0);
  const [cfgScoreMax, setCfgScoreMax] = useState(100);
  const [cfgPrecedence, setCfgPrecedence] = useState<string[]>(['points', 'speaker_points', 'margin']);
  const [cfgResultsPublic, setCfgResultsPublic] = useState(true);

  useEffect(() => {
    loadSetupData();
  }, [slug]);

  const loadSetupData = () => {
    fetchAPI(`/api/t/${slug}/institutions`).then(d => setInstitutions(d || [])).catch(console.error);
    fetchAPI(`/api/t/${slug}/teams`).then(d => setTeams(d || [])).catch(console.error);
    fetchAPI(`/api/t/${slug}/adjudicators`).then(d => setAdjudicators(d || [])).catch(console.error);
    fetchAPI(`/api/t/${slug}/break-categories`).then(d => setBreakCategories(d || [])).catch(console.error);
    fetchAPI(`/api/t/${slug}/config`).then((d) => {
      if (!d) return;
      setCfgSides(d.sides || '');
      setCfgScoreMin(Number(d.score_min ?? 0));
      setCfgScoreMax(Number(d.score_max ?? 100));
      if (Array.isArray(d.ranking_precedence) && d.ranking_precedence.length > 0) setCfgPrecedence(d.ranking_precedence);
      if (d.public_features && typeof d.public_features.results_public === 'boolean') setCfgResultsPublic(d.public_features.results_public);
    }).catch(console.error);
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
    } catch (err: any) { alert(err.message); }
  };

  const handleDeleteTeam = async (id: string) => {
    if (confirm("Delete this team?")) {
      try {
        await fetchAPI(`/api/t/${slug}/teams/${id}`, 'DELETE');
        loadSetupData();
      } catch (err: any) { alert(err.message); }
    }
  };

  const toggleTeamFlag = async (team: any, flag: 'is_novice' | 'is_esl' | 'is_efl') => {
    try {
      await fetchAPI(`/api/t/${slug}/teams/${team.id}`, 'PUT', { [flag]: !team[flag] });
      loadSetupData();
    } catch (err: any) { alert(err.message); }
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

  // --- Break Categories ---
  const resetBcForm = () => {
    setEditingBcId(null);
    setBcName('');
    setBcSeq('0');
    setBcSize('');
    setBcBasePoints('');
  };

  const handleSaveCategory = async (e: React.FormEvent) => {
    e.preventDefault();
    const body: any = { name: bcName, seq: Number(bcSeq) || 0 };
    if (bcSize !== '') body.size = Number(bcSize);
    if (bcBasePoints !== '') body.base_points = Number(bcBasePoints);
    try {
      if (editingBcId) {
        await fetchAPI(`/api/t/${slug}/break-categories/${editingBcId}`, 'PUT', body);
      } else {
        await fetchAPI(`/api/t/${slug}/break-categories`, 'POST', body);
      }
      resetBcForm();
      loadSetupData();
    } catch (err: any) { alert(err.message); }
  };

  const handleEditCategory = (c: any) => {
    setEditingBcId(c.id);
    setBcName(c.name);
    setBcSeq(String(c.seq ?? 0));
    setBcSize(c.size == null ? '' : String(c.size));
    setBcBasePoints(c.base_points == null ? '' : String(c.base_points));
  };

  const handleDeleteCategory = async (id: string) => {
    if (confirm("Delete this break category?")) {
      try {
        await fetchAPI(`/api/t/${slug}/break-categories/${id}`, 'DELETE');
        if (editingBcId === id) resetBcForm();
        loadSetupData();
      } catch (err: any) { alert(err.message); }
    }
  };

  // --- Config ---
  const movePrecedenceRule = (idx: number, dir: -1 | 1) => {
    const next = [...cfgPrecedence];
    const target = idx + dir;
    if (target < 0 || target >= next.length) return;
    [next[idx], next[target]] = [next[target], next[idx]];
    setCfgPrecedence(next);
  };

  const handleSaveConfig = async () => {
    try {
      await fetchAPI(`/api/t/${slug}/config`, 'PUT', {
        sides: cfgSides,
        score_min: Number(cfgScoreMin),
        score_max: Number(cfgScoreMax),
        ranking_precedence: cfgPrecedence,
        public_features: { results_public: cfgResultsPublic }
      });
      alert('Configuration saved.');
    } catch (err: any) { alert(err.message); }
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
      alert(`Imported ${type} successfully!`);
    } catch (err: any) {
      alert("CSV Import failed: " + err.message);
    }
  };

  return (
    <div>
      <h2>Setup</h2>

      <div className="tabs">
        <button className={`tab-btn ${activeTab === 'institutions' ? 'active' : ''}`} onClick={() => setActiveTab('institutions')}>
          Institutions ({institutions.length})
        </button>
        <button className={`tab-btn ${activeTab === 'teams' ? 'active' : ''}`} onClick={() => setActiveTab('teams')}>
          Teams ({teams.length})
        </button>
        <button className={`tab-btn ${activeTab === 'adjudicators' ? 'active' : ''}`} onClick={() => setActiveTab('adjudicators')}>
          Adjudicators ({adjudicators.length})
        </button>
        <button className={`tab-btn ${activeTab === 'config' ? 'active' : ''}`} onClick={() => setActiveTab('config')}>
          Configuration
        </button>
        <button className={`tab-btn ${activeTab === 'categories' ? 'active' : ''}`} onClick={() => setActiveTab('categories')}>
          Break Categories ({breakCategories.length})
        </button>
      </div>

      {/* --- TAB: INSTITUTIONS --- */}
      {activeTab === 'institutions' && (
        <div className="card" style={{ maxWidth: '640px' }}>
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
                <button type="submit" className="btn btn-primary"><Plus size={16} /> Add</button>
              </form>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '1rem', borderTop: '1px solid var(--border)', paddingTop: '1rem' }}>
                <FileSpreadsheet size={16} />
                <span style={{ fontSize: '0.8rem' }}>CSV Bulk Import:</span>
                <input type="file" accept=".csv" onChange={e => handleCSVUpload(e, 'institutions')} style={{ fontSize: '0.75rem', width: '100%' }} />
              </div>
            </>
          )}
          <div style={{ maxHeight: '420px', overflowY: 'auto' }}>
            {institutions.map(i => (
              <div key={i.id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.5rem 0', borderBottom: '1px solid var(--border)' }}>
                <span>{i.name} ({i.code})</span>
                {!isReadOnly && (
                  <button className="btn btn-danger" style={{ padding: '2px 6px', fontSize: '0.75rem' }} onClick={() => handleDeleteInst(i.id)}>Delete</button>
                )}
              </div>
            ))}
            {institutions.length === 0 && (
              <div style={{ padding: '2rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>No institutions yet.</div>
            )}
          </div>
        </div>
      )}

      {/* --- TAB: TEAMS --- */}
      {activeTab === 'teams' && (
        <div className="card" style={{ maxWidth: '640px' }}>
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
                <button type="submit" className="btn btn-primary"><Plus size={16} /> Add Team</button>
              </form>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '1rem', borderTop: '1px solid var(--border)', paddingTop: '1rem' }}>
                <FileSpreadsheet size={16} />
                <span style={{ fontSize: '0.8rem' }}>CSV Import:</span>
                <input type="file" accept=".csv" onChange={e => handleCSVUpload(e, 'teams')} style={{ fontSize: '0.75rem', width: '100%' }} />
              </div>
            </>
          )}
          <div style={{ maxHeight: '420px', overflowY: 'auto' }}>
            {teams.map(t => (
              <div key={t.id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.5rem 0', borderBottom: '1px solid var(--border)' }}>
                <div>
                  <div style={{ fontWeight: '500' }}>{t.name}</div>
                  <div style={{ fontSize: '0.75rem', color: 'var(--text-mute)' }}>Inst: {t.institution_code || 'None'}</div>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.25rem' }}>
                  {!isReadOnly && (
                    <>
                      {(['is_novice', 'is_esl', 'is_efl'] as const).map(flag => (
                        <button
                          key={flag}
                          title={flag.replace('is_', '').toUpperCase()}
                          onClick={() => toggleTeamFlag(t, flag)}
                          style={{
                            fontSize: '0.65rem',
                            fontWeight: '600',
                            padding: '2px 7px',
                            borderRadius: '999px',
                            border: '1px solid ' + (t[flag] ? 'var(--accent)' : 'var(--border)'),
                            background: t[flag] ? 'var(--accent)' : 'transparent',
                            color: t[flag] ? '#fff' : 'var(--text-mute)',
                            cursor: 'pointer'
                          }}
                        >
                          {flag.replace('is_', '').toUpperCase()}
                        </button>
                      ))}
                      <button className="btn btn-danger" style={{ padding: '2px 6px', fontSize: '0.75rem' }} onClick={() => handleDeleteTeam(t.id)}>Delete</button>
                    </>
                  )}
                  {isReadOnly && (
                    <>
                      {t.is_novice && <span style={{ fontSize: '0.65rem', fontWeight: 600, color: 'var(--accent)' }}>NOVICE</span>}
                      {t.is_esl && <span style={{ fontSize: '0.65rem', fontWeight: 600, color: 'var(--accent)' }}>ESL</span>}
                      {t.is_efl && <span style={{ fontSize: '0.65rem', fontWeight: 600, color: 'var(--accent)' }}>EFL</span>}
                    </>
                  )}
                </div>
              </div>
            ))}
            {teams.length === 0 && (
              <div style={{ padding: '2rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>No teams yet.</div>
            )}
          </div>
        </div>
      )}

      {/* --- TAB: ADJUDICATORS --- */}
      {activeTab === 'adjudicators' && (
        <div className="card" style={{ maxWidth: '640px' }}>
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
                <button type="submit" className="btn btn-primary"><Plus size={16} /> Add Judge</button>
              </form>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '1rem', borderTop: '1px solid var(--border)', paddingTop: '1rem' }}>
                <FileSpreadsheet size={16} />
                <span style={{ fontSize: '0.8rem' }}>CSV Import:</span>
                <input type="file" accept=".csv" onChange={e => handleCSVUpload(e, 'adjudicators')} style={{ fontSize: '0.75rem', width: '100%' }} />
              </div>
            </>
          )}
          <div style={{ maxHeight: '420px', overflowY: 'auto' }}>
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
            {adjudicators.length === 0 && (
              <div style={{ padding: '2rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>No adjudicators yet.</div>
            )}
          </div>
        </div>
      )}

      {/* --- TAB: CONFIGURATION --- */}
      {activeTab === 'config' && (
        <div className="card" style={{ maxWidth: '640px' }}>
          <h3>Configuration</h3>
          <div className="form-group">
            <label style={{ fontSize: '0.8rem', fontWeight: 500 }}>Sides (comma-separated)</label>
            <input type="text" className="input" placeholder="OG,OO,CG,CO" value={cfgSides} disabled={isReadOnly} onChange={e => setCfgSides(e.target.value)} />
          </div>
          <div className="form-group" style={{ display: 'flex', gap: '0.75rem' }}>
            <div style={{ flex: 1 }}>
              <label style={{ fontSize: '0.8rem', fontWeight: 500 }}>Speaker Score Min</label>
              <input type="number" className="input" value={cfgScoreMin} disabled={isReadOnly} onChange={e => setCfgScoreMin(Number(e.target.value))} />
            </div>
            <div style={{ flex: 1 }}>
              <label style={{ fontSize: '0.8rem', fontWeight: 500 }}>Speaker Score Max</label>
              <input type="number" className="input" value={cfgScoreMax} disabled={isReadOnly} onChange={e => setCfgScoreMax(Number(e.target.value))} />
            </div>
          </div>
          <div className="form-group">
            <label style={{ fontSize: '0.8rem', fontWeight: 500 }}>Ranking Precedence (highest first)</label>
            <div style={{ border: '1px solid var(--border)', borderRadius: '6px' }}>
              {cfgPrecedence.map((rule, idx) => (
                <div key={rule} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.5rem 0.75rem', borderBottom: idx < cfgPrecedence.length - 1 ? '1px solid var(--border)' : 'none' }}>
                  <span style={{ fontSize: '0.85rem' }}>{idx + 1}. {precedenceLabels[rule] || rule}</span>
                  {!isReadOnly && (
                    <span style={{ display: 'flex', gap: '0.15rem' }}>
                      <button className="btn" style={{ padding: '2px 4px' }} disabled={idx === 0} onClick={() => movePrecedenceRule(idx, -1)}><ChevronUp size={14} /></button>
                      <button className="btn" style={{ padding: '2px 4px' }} disabled={idx === cfgPrecedence.length - 1} onClick={() => movePrecedenceRule(idx, 1)}><ChevronDown size={14} /></button>
                    </span>
                  )}
                </div>
              ))}
            </div>
          </div>
          <div className="form-group">
            <label style={{ fontSize: '0.8rem', fontWeight: 500 }}>Public Portal Features</label>
            <div style={{ marginTop: '0.25rem' }}>
              <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.85rem' }}>
                <input type="checkbox" checked={cfgResultsPublic} disabled={isReadOnly} onChange={e => setCfgResultsPublic(e.target.checked)} />
                Results public
              </label>
            </div>
          </div>
          {!isReadOnly && (
            <button className="btn btn-primary" onClick={handleSaveConfig}>Save Configuration</button>
          )}
        </div>
      )}

      {/* --- TAB: BREAK CATEGORIES --- */}
      {activeTab === 'categories' && (
        <div className="card" style={{ maxWidth: '720px' }}>
          <h3>Break Categories</h3>
          {!isReadOnly && (
            <form onSubmit={handleSaveCategory} style={{ marginBottom: '1.5rem' }}>
              <div className="form-group" style={{ display: 'flex', gap: '0.5rem' }}>
                <input type="text" className="input" placeholder="Name (e.g. Open, Novice)" value={bcName} onChange={e => setBcName(e.target.value)} required />
                <input type="number" className="input" placeholder="Seq" title="Sequence" value={bcSeq} onChange={e => setBcSeq(e.target.value)} style={{ width: '80px' }} />
                <input type="number" className="input" placeholder="Size" title="Cap on qualifiers (optional)" value={bcSize} onChange={e => setBcSize(e.target.value)} style={{ width: '80px' }} />
                <input type="number" className="input" placeholder="Min Pts" title="Base points eligibility (optional)" value={bcBasePoints} onChange={e => setBcBasePoints(e.target.value)} style={{ width: '90px' }} />
              </div>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <button type="submit" className="btn btn-primary"><Plus size={16} /> {editingBcId ? 'Update' : 'Add Category'}</button>
                {editingBcId && (
                  <button type="button" className="btn" onClick={resetBcForm}>Cancel</button>
                )}
              </div>
            </form>
          )}
          <div style={{ maxHeight: '420px', overflowY: 'auto' }}>
            {breakCategories.map(c => (
              <div key={c.id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.5rem 0', borderBottom: '1px solid var(--border)' }}>
                <div>
                  <div style={{ fontWeight: '500' }}>{c.name}</div>
                  <div style={{ fontSize: '0.75rem', color: 'var(--text-mute)' }}>
                    Seq: {c.seq}{c.size != null ? ` | Size: ${c.size}` : ''}{c.base_points != null ? ` | Min Pts: ${c.base_points}` : ''}
                  </div>
                </div>
                {!isReadOnly && (
                  <div style={{ display: 'flex', gap: '0.35rem' }}>
                    <button className="btn" style={{ padding: '2px 8px', fontSize: '0.75rem' }} onClick={() => handleEditCategory(c)}>Edit</button>
                    <button className="btn btn-danger" style={{ padding: '2px 6px', fontSize: '0.75rem' }} onClick={() => handleDeleteCategory(c.id)}>Delete</button>
                  </div>
                )}
              </div>
            ))}
            {breakCategories.length === 0 && (
              <div style={{ padding: '2rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>No break categories yet.</div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
