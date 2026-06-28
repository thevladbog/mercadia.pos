import { Routes, Route, Navigate } from 'react-router-dom';

import { useAuth } from '@/auth/AuthProvider.js';
import { LoginPage } from '@/pages/LoginPage.js';
import { DashboardPage } from '@/pages/DashboardPage.js';
import { IssueChangeFundPage } from '@/pages/IssueChangeFundPage.js';
import { ReceiveCashPage } from '@/pages/ReceiveCashPage.js';
import { FinalCollectionPage } from '@/pages/FinalCollectionPage.js';
import { SafeRecountPage } from '@/pages/SafeRecountPage.js';
import { BankCollectionPage } from '@/pages/BankCollectionPage.js';
import { BusinessExpensePage } from '@/pages/BusinessExpensePage.js';
import { MonitoringPage } from '@/pages/MonitoringPage.js';
import { OperationJournalPage } from '@/pages/OperationJournalPage.js';
import { ShiftHandoverPage } from '@/pages/ShiftHandoverPage.js';

function RequireAuth({ children }: { children: React.ReactNode }) {
  const { session } = useAuth();
  if (!session) {
    return <Navigate to="/login" replace />;
  }
  return <>{children}</>;
}

function RequireSeniorCashier({ children }: { children: React.ReactNode }) {
  const { session } = useAuth();
  if (!session) {
    return <Navigate to="/login" replace />;
  }
  if (!session.roles.includes('senior_cashier') && !session.roles.includes('admin')) {
    return <Navigate to="/login" replace />;
  }
  return <>{children}</>;
}

export function App() {
  const { session } = useAuth();

  return (
    <Routes>
      <Route path="/login" element={session ? <Navigate to="/dashboard" replace /> : <LoginPage />} />

      <Route
        path="/dashboard"
        element={
          <RequireSeniorCashier>
            <DashboardPage />
          </RequireSeniorCashier>
        }
      />

      <Route
        path="/cash/change-fund"
        element={
          <RequireSeniorCashier>
            <IssueChangeFundPage />
          </RequireSeniorCashier>
        }
      />
      <Route
        path="/cash/receive"
        element={
          <RequireSeniorCashier>
            <ReceiveCashPage />
          </RequireSeniorCashier>
        }
      />
      <Route
        path="/cash/final-collection"
        element={
          <RequireSeniorCashier>
            <FinalCollectionPage />
          </RequireSeniorCashier>
        }
      />
      <Route
        path="/cash/safe-recount"
        element={
          <RequireSeniorCashier>
            <SafeRecountPage />
          </RequireSeniorCashier>
        }
      />
      <Route
        path="/cash/bank-collection"
        element={
          <RequireSeniorCashier>
            <BankCollectionPage />
          </RequireSeniorCashier>
        }
      />
      <Route
        path="/cash/expense"
        element={
          <RequireSeniorCashier>
            <BusinessExpensePage />
          </RequireSeniorCashier>
        }
      />

      <Route
        path="/monitoring"
        element={
          <RequireAuth>
            <MonitoringPage />
          </RequireAuth>
        }
      />

      <Route
        path="/journal"
        element={
          <RequireSeniorCashier>
            <OperationJournalPage />
          </RequireSeniorCashier>
        }
      />

      <Route
        path="/handover"
        element={
          <RequireSeniorCashier>
            <ShiftHandoverPage />
          </RequireSeniorCashier>
        }
      />

      <Route path="*" element={<Navigate to={session ? '/dashboard' : '/login'} replace />} />
    </Routes>
  );
}
