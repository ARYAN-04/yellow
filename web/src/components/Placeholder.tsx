import { Construction } from 'lucide-react';

interface PlaceholderPageProps {
  title: string;
  description?: string;
}

export default function PlaceholderPage({ title, description }: PlaceholderPageProps) {
  return (
    <div>
      <h2 style={{ marginBottom: '1.5rem' }}>{title}</h2>
      <div className="card" style={{ textAlign: 'center', padding: '4rem 2rem' }}>
        <div style={{ display: 'inline-flex', padding: '1rem', borderRadius: '50%', background: 'rgba(0, 0, 0, 0.05)', color: 'var(--text-mute)', marginBottom: '1.5rem' }}>
          <Construction size={28} />
        </div>
        <h3 style={{ marginBottom: '0.5rem' }}>Coming in a later update</h3>
        <p style={{ fontSize: '0.9rem', color: 'var(--text-mute)', margin: 0 }}>
          {description || 'This section is planned for an upcoming release.'}
        </p>
      </div>
    </div>
  );
}
