import { useOutletContext } from 'react-router-dom';
import PlaceholderPage from '../../../components/Placeholder';
import type { RoundContext } from '../../../lib/api';

export default function RoundMotions() {
  const { isReadOnly } = useOutletContext<RoundContext>();
  return (
    <PlaceholderPage
      title="Motions"
      description={isReadOnly ? 'Motion data is not available for archived records.' : 'Motion editing and release for this round is coming in a later update.'}
    />
  );
}
