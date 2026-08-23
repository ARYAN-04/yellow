import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import LandingPage from './pages/LandingPage';
import ParticipantDashboard from './pages/ParticipantDashboard';
import JudgeDashboard from './pages/JudgeDashboard';
import CheckinPage from './pages/CheckinPage';
import AdminLayout from './components/AdminLayout';
import AdminOverview from './pages/admin/AdminOverview';
import Standings from './pages/admin/Standings';
import Setup from './pages/admin/Setup';
import Conflicts from './pages/admin/Conflicts';
import Checkins from './pages/admin/Checkins';
import Feedback from './pages/admin/Feedback';
import Breaks from './pages/admin/Breaks';
import Brackets from './pages/admin/Brackets';
import RoundLayout from './pages/admin/rounds/RoundLayout';
import RoundAvailability from './pages/admin/rounds/RoundAvailability';
import RoundDraw from './pages/admin/rounds/RoundDraw';
import RoundResults from './pages/admin/rounds/RoundResults';
import RoundMotions from './pages/admin/rounds/RoundMotions';

const queryClient = new QueryClient();

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<LandingPage />} />
          <Route path="/p/:token" element={<ParticipantDashboard />} />
          <Route path="/j/:token" element={<JudgeDashboard />} />
          <Route path="/checkin/:token" element={<CheckinPage />} />
          <Route path="/t/:slug/admin" element={<AdminLayout />}>
            <Route index element={<AdminOverview />} />
            <Route path="feedback" element={<Feedback />} />
            <Route path="standings" element={<Standings />} />
            <Route path="breaks" element={<Breaks />} />
            <Route path="brackets" element={<Brackets />} />
            <Route path="setup" element={<Setup />} />
            <Route path="conflicts" element={<Conflicts />} />
            <Route path="checkins" element={<Checkins />} />
            <Route path="rounds/:roundId" element={<RoundLayout />}>
              <Route index element={<Navigate to="draw" replace />} />
              <Route path="availability" element={<RoundAvailability />} />
              <Route path="draw" element={<RoundDraw />} />
              <Route path="results" element={<RoundResults />} />
              <Route path="motions" element={<RoundMotions />} />
            </Route>
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}

export default App;
