import { useState, useEffect } from 'react';
import { useParams, useOutletContext } from 'react-router-dom';
import { Copy, Check, QrCode } from 'lucide-react';
import { QRCodeSVG } from 'qrcode.react';
import { fetchAPI, type AdminContext } from '../../lib/api';

interface CheckinRow {
  entity_type: string;
  entity_id: string;
  entity_name: string;
  checked_in: boolean;
  checkin_token: string;
}

function CheckinQR({ row }: { row: CheckinRow }) {
  const [copied, setCopied] = useState(false);
  const url = `${window.location.origin}/checkin/${row.checkin_token}`;

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(url);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch { /* clipboard unavailable */ }
  };

  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
      <div style={{ background: '#fff', padding: '6px', borderRadius: '6px', border: '1px solid var(--border)' }}>
        <QRCodeSVG value={url} size={72} />
      </div>
      <button type="button" className="btn btn-secondary" style={{ padding: '4px 8px', fontSize: '0.75rem' }} onClick={copy} title="Copy check-in URL">
        {copied ? <Check size={13} /> : <Copy size={13} />} {copied ? 'Copied' : 'URL'}
      </button>
    </div>
  );
}

export default function Checkins() {
  const { slug } = useParams<{ slug: string }>();
  const { isReadOnly } = useOutletContext<AdminContext>();
  const [rows, setRows] = useState<CheckinRow[]>([]);
  const [filter, setFilter] = useState<'team' | 'adjudicator'>('team');

  const load = () => {
    fetchAPI(`/api/t/${slug}/checkins`).then(d => setRows(d || [])).catch(console.error);
  };

  useEffect(() => {
    load();
  }, [slug]);

  const [search, setSearch] = useState('');

  const toggle = async (row: CheckinRow) => {
    try {
      await fetchAPI(`/api/t/${slug}/checkins/${row.entity_type}/${row.entity_id}`, 'POST', { checked_in: !row.checked_in });
      load();
    } catch (err: any) { alert(err.message); }
  };

  const handleBulkCheckin = async (checkedIn: boolean) => {
    const targets = visible.filter(r => r.checked_in !== checkedIn);
    if (targets.length === 0) return;
    try {
      await Promise.all(
        targets.map(r =>
          fetchAPI(`/api/t/${slug}/checkins/${r.entity_type}/${r.entity_id}`, 'POST', { checked_in: checkedIn })
        )
      );
      load();
    } catch (err: any) {
      alert("Bulk check-in failed: " + err.message);
    }
  };

  const visible = rows
    .filter(r => r.entity_type === filter)
    .filter(r => !search || r.entity_name.toLowerCase().includes(search.toLowerCase()));

  const totalCurrent = rows.filter(r => r.entity_type === filter).length;
  const checkedInCount = rows.filter(r => r.entity_type === filter && r.checked_in).length;

  return (
    <div>
      <h2>Check-ins &amp; Attendance</h2>
      <p style={{ color: 'var(--text-mute)', fontSize: '0.85rem', marginTop: '-0.5rem' }}>
        Manage participant attendance directly or provide self-service QR codes. Round availability can be synced from these records.
      </p>

      <div className="tabs" style={{ marginBottom: '1rem' }}>
        <button type="button" className={`tab-btn ${filter === 'team' ? 'active' : ''}`} onClick={() => setFilter('team')}>
          Teams ({rows.filter(r => r.entity_type === 'team').length})
        </button>
        <button type="button" className={`tab-btn ${filter === 'adjudicator' ? 'active' : ''}`} onClick={() => setFilter('adjudicator')}>
          Adjudicators ({rows.filter(r => r.entity_type === 'adjudicator').length})
        </button>
      </div>

      <div className="card" style={{ maxWidth: '860px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem', flexWrap: 'wrap', gap: '0.5rem' }}>
          <div>
            <h3 style={{ margin: 0 }}>
              <QrCode size={16} style={{ verticalAlign: '-2px' }} /> {filter === 'team' ? 'Teams' : 'Adjudicators'}
            </h3>
            <span style={{ fontSize: '0.8rem', color: 'var(--text-mute)' }}>
              {checkedInCount} / {totalCurrent} checked in
            </span>
          </div>

          <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
            <input
              type="text"
              className="input"
              placeholder={`Search ${filter}s...`}
              value={search}
              onChange={e => setSearch(e.target.value)}
              style={{ width: '180px', padding: '4px 8px', fontSize: '0.8rem' }}
            />
            {!isReadOnly && (
              <>
                <button
                  type="button"
                  className="btn btn-primary"
                  style={{ padding: '4px 10px', fontSize: '0.78rem' }}
                  onClick={() => handleBulkCheckin(true)}
                  disabled={checkedInCount === totalCurrent}
                >
                  Check In All
                </button>
                <button
                  type="button"
                  className="btn btn-secondary"
                  style={{ padding: '4px 10px', fontSize: '0.78rem' }}
                  onClick={() => handleBulkCheckin(false)}
                  disabled={checkedInCount === 0}
                >
                  Undo All
                </button>
              </>
            )}
          </div>
        </div>

        <div style={{ maxHeight: '520px', overflowY: 'auto' }}>
          {visible.map(row => (
            <div key={`${row.entity_type}:${row.entity_id}`} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.65rem 0', borderBottom: '1px solid var(--border)' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
                <span style={{ fontWeight: 500 }}>{row.entity_name}</span>
                {row.checked_in && <span className="badge badge-success">Checked in</span>}
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                {!isReadOnly && (
                  <button
                    type="button"
                    className={`btn ${row.checked_in ? 'btn-secondary' : 'btn-primary'}`}
                    style={{ padding: '4px 10px', fontSize: '0.78rem' }}
                    onClick={() => toggle(row)}
                  >
                    {row.checked_in ? 'Undo' : 'Check In'}
                  </button>
                )}
                <CheckinQR row={row} />
              </div>
            </div>
          ))}
          {visible.length === 0 && (
            <div style={{ padding: '2rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>
              {search ? 'No matches found.' : 'Nothing to check in yet.'}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
