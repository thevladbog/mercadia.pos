import { Navigate, Route, Routes } from 'react-router-dom';

import { isSeniorCashier } from '@/auth/permissions.js';
import { useAuth } from '@/auth/useAuth.js';

function RoleAwareRedirect() {
  const { roles } = useAuth();
  const target = isSeniorCashier(roles) ? '/senior-cashier/dashboard' : '/central/dashboard';
  return <Navigate replace to={target} />;
}

import { AuthProvider } from '@/auth/AuthProvider.js';
import { LoginPage } from '@/auth/LoginPage.js';
import { ProtectedRoute } from '@/auth/ProtectedRoute.js';
import { RequireCentralAdmin } from '@/auth/RequireCentralAdmin.js';
import { RequireSeniorCashierOrAdmin } from '@/auth/RequireSeniorCashierOrAdmin.js';
import { UnauthorizedBridge } from '@/auth/UnauthorizedBridge.js';
import { AppLayout } from '@/layout/AppLayout.js';
import { CentralCatalogPage } from '@/pages/CentralCatalogPage.js';
import { CentralColorSchemesPage } from '@/pages/CentralColorSchemesPage.js';
import { CentralDashboardPage } from '@/pages/CentralDashboardPage.js';
import { CentralLayoutTemplatesPage } from '@/pages/CentralLayoutTemplatesPage.js';
import { CentralReportingPage } from '@/pages/CentralReportingPage.js';
import { CentralStoresPage } from '@/pages/CentralStoresPage.js';
import { CentralSyncExplorerPage } from '@/pages/CentralSyncExplorerPage.js';
import { SyncEntityDetailPage } from '@/pages/SyncEntityDetailPage.js';
import { CentralUsersPage } from '@/pages/CentralUsersPage.js';
import { CreateCentralUserPage } from '@/pages/CreateCentralUserPage.js';
import { CreateColorSchemePage } from '@/pages/CreateColorSchemePage.js';
import { CreateLayoutTemplatePage } from '@/pages/CreateLayoutTemplatePage.js';
import { EditCentralUserPage } from '@/pages/EditCentralUserPage.js';
import { EditColorSchemePage } from '@/pages/EditColorSchemePage.js';
import { EditLayoutTemplatePage } from '@/pages/EditLayoutTemplatePage.js';
import { RegisterStorePage } from '@/pages/RegisterStorePage.js';
import { StoreMonitoringPage } from '@/pages/StoreMonitoringPage.js';
import { StoreSafePage } from '@/pages/StoreSafePage.js';
import { StoreEodPage } from '@/pages/StoreEodPage.js';
import { StoreCredentialManagementPage } from '@/pages/StoreCredentialManagementPage.js';
import { StoreSettingsPage } from '@/pages/StoreSettingsPage.js';
import { TerminalMonitoringDetailPage } from '@/pages/TerminalMonitoringDetailPage.js';
import { StoreReportingPage } from '@/pages/StoreReportingPage.js';
import { SeniorCashierDashboardPage } from '@/pages/SeniorCashierDashboardPage.js';
import { IssueChangeFundPage } from '@/pages/IssueChangeFundPage.js';
import { ReceiveCashPage } from '@/pages/ReceiveCashPage.js';
import { SafeRecountPage } from '@/pages/SafeRecountPage.js';
import { BankCollectionPage } from '@/pages/BankCollectionPage.js';
import { BusinessExpensePage } from '@/pages/BusinessExpensePage.js';
import { FinalCollectionPage } from '@/pages/FinalCollectionPage.js';
import { OperationJournalPage } from '@/pages/OperationJournalPage.js';
import { ShiftHandoverPage } from '@/pages/ShiftHandoverPage.js';

export function App() {
  return (
    <AuthProvider>
      <UnauthorizedBridge />
      <Routes>
        <Route element={<LoginPage />} path="/login" />
        <Route element={<ProtectedRoute />}>
          <Route element={<AppLayout />}>
            <Route element={<CentralDashboardPage />} path="/central/dashboard" />
            <Route element={<CentralReportingPage />} path="/central/reporting" />
            <Route element={<StoreReportingPage />} path="/central/reporting/stores/:storeId" />
            <Route element={<RequireSeniorCashierOrAdmin />}>
              <Route element={<SeniorCashierDashboardPage />} path="/senior-cashier/dashboard" />
              <Route element={<IssueChangeFundPage />} path="/senior-cashier/change-fund" />
              <Route element={<ReceiveCashPage />} path="/senior-cashier/receive-cash" />
              <Route element={<SafeRecountPage />} path="/senior-cashier/safe-recount" />
              <Route element={<BankCollectionPage />} path="/senior-cashier/bank-collection" />
              <Route element={<BusinessExpensePage />} path="/senior-cashier/expense" />
              <Route element={<FinalCollectionPage />} path="/senior-cashier/collection" />
              <Route element={<OperationJournalPage />} path="/senior-cashier/journal" />
              <Route element={<ShiftHandoverPage />} path="/senior-cashier/handover" />
            </Route>
            <Route element={<CentralStoresPage />} path="/central/stores" />
            <Route element={<CentralSyncExplorerPage />} path="/central/sync" />
            <Route element={<CentralCatalogPage />} path="/central/catalog" />
            <Route
              element={<SyncEntityDetailPage />}
              path="/central/sync/stores/:storeId/payments/:paymentId"
            />
            <Route
              element={<SyncEntityDetailPage />}
              path="/central/sync/stores/:storeId/cash-movements/:cashMovementId"
            />
            <Route
              element={<SyncEntityDetailPage />}
              path="/central/sync/stores/:storeId/fiscal-documents/:fiscalDocumentId"
            />
            <Route
              element={<SyncEntityDetailPage />}
              path="/central/sync/stores/:storeId/returns/:returnId"
            />
            <Route
              element={<SyncEntityDetailPage />}
              path="/central/sync/stores/:storeId/operational-days/:operationalDayId"
            />
            <Route element={<StoreMonitoringPage />} path="/store/monitoring" />
            <Route element={<StoreSafePage />} path="/store/safe" />
            <Route element={<StoreEodPage />} path="/store/eod" />
            <Route element={<StoreCredentialManagementPage />} path="/store/credentials" />
            <Route element={<StoreSettingsPage />} path="/store/settings" />
            <Route
              element={<TerminalMonitoringDetailPage />}
              path="/store/monitoring/stores/:storeId/terminals/:terminalId"
            />
            <Route element={<RequireCentralAdmin />}>
              <Route element={<RegisterStorePage />} path="/central/stores/new" />
              <Route element={<CentralUsersPage />} path="/central/users" />
              <Route element={<CreateCentralUserPage />} path="/central/users/new" />
              <Route element={<EditCentralUserPage />} path="/central/users/:userId" />
              <Route element={<CentralColorSchemesPage />} path="/central/color-schemes" />
              <Route element={<CreateColorSchemePage />} path="/central/color-schemes/new" />
              <Route element={<EditColorSchemePage />} path="/central/color-schemes/:schemeId" />
              <Route element={<CentralLayoutTemplatesPage />} path="/central/layout-templates" />
              <Route element={<CreateLayoutTemplatePage />} path="/central/layout-templates/new" />
              <Route
                element={<EditLayoutTemplatePage />}
                path="/central/layout-templates/:templateId"
              />
            </Route>
          </Route>
        </Route>
        <Route element={<ProtectedRoute />}>
          <Route element={<RoleAwareRedirect />} path="/" />
        </Route>
        <Route element={<RoleAwareRedirect />} path="*" />
      </Routes>
    </AuthProvider>
  );
}
