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

  const toggle = async (row: CheckinRow) => {
    try {
      await fetchAPI(`/api/t/${slug}/checkins/${row.entity_type}/${row.entity_id}`, 'POST', { checked_in: !row.checked_in });
      load();
    } catch (err: any) { alert(err.message); }
  };

  const visible = rows.filter(r => r.entity_type === filter);

  return (
    <div>
      <h2>Check-ins</h2>
      <p style={{ color: 'var(--text-mute)', fontSize: '0.85rem', marginTop: '-0.5rem' }}>
        Each team and adjudicator has a personal QR code. Scanning it opens a self-service check-in page; availability for rounds can be synced from these check-ins.
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
        <h3><QrCode size={16} style={{ verticalAlign: '-2px' }} /> {filter === 'team' ? 'Teams' : 'Adjudicators'}</h3>
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
            <div style={{ padding: '2rem 0', color: 'var(--text-mute)', textAlign: 'center' }}>Nothing to check in yet.</div>
          )}
        </div>
      </div>
    </div>
  );
}
