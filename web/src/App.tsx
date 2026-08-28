import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import ErrorBoundary from './components/ErrorBoundary';
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

// Public Portal Pages
import PublicStandings from './pages/public/PublicStandings';
import PublicResults from './pages/public/PublicResults';
import PublicDraw from './pages/public/PublicDraw';
import PublicMotions from './pages/public/PublicMotions';

// Printable offline pages
import PrintMasterDraw from './pages/admin/print/PrintMasterDraw';
import PrintBallots from './pages/admin/print/PrintBallots';
import PrintMotions from './pages/admin/print/PrintMotions';

// Projector / Display pages
import ProjectorMotion from './pages/display/ProjectorMotion';
import ProjectorDraw from './pages/display/ProjectorDraw';

const queryClient = new QueryClient();

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ErrorBoundary fallbackTitle="Application Error">
        <BrowserRouter>
          <Routes>
            <Route path="/" element={<LandingPage />} />
            <Route path="/p/:token" element={<ParticipantDashboard />} />
            <Route path="/j/:token" element={<JudgeDashboard />} />
            <Route path="/checkin/:token" element={<CheckinPage />} />

            {/* Public Portal Routes */}
            <Route path="/t/:slug/standings" element={<PublicStandings />} />
            <Route path="/t/:slug/results" element={<PublicResults />} />
            <Route path="/t/:slug/draw" element={<PublicDraw />} />
            <Route path="/t/:slug/motions" element={<PublicMotions />} />

            {/* Printable Routes (Clean full page outside standard admin layout) */}
            <Route path="/t/:slug/admin/rounds/:roundId/print/draw" element={<PrintMasterDraw />} />
            <Route path="/t/:slug/admin/rounds/:roundId/print/ballots" element={<PrintBallots />} />
            <Route path="/t/:slug/admin/rounds/:roundId/print/motions" element={<PrintMotions />} />

            {/* Projector / Display Presentation Routes */}
            <Route path="/t/:slug/display/motion" element={<ProjectorMotion />} />
            <Route path="/t/:slug/display/draw" element={<ProjectorDraw />} />

            {/* Admin Workspace Layout & Pages */}
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
      </ErrorBoundary>
    </QueryClientProvider>
  );
}

export default App;
