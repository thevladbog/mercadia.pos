import { Navigate, Route, Routes } from 'react-router-dom';

import { AuthProvider } from './auth/AuthProvider.js';
import { LoginPage } from './auth/LoginPage.js';
import { ProtectedRoute } from './auth/ProtectedRoute.js';
import { RequireCentralAdmin } from './auth/RequireCentralAdmin.js';
import { UnauthorizedBridge } from './auth/UnauthorizedBridge.js';
import { AppLayout } from './layout/AppLayout.js';
import { CentralReportingPage } from './pages/CentralReportingPage.js';
import { CentralUsersPage } from './pages/CentralUsersPage.js';
import { CreateCentralUserPage } from './pages/CreateCentralUserPage.js';
import { EditCentralUserPage } from './pages/EditCentralUserPage.js';
import { StoreMonitoringPage } from './pages/StoreMonitoringPage.js';

export function App() {
  return (
    <AuthProvider>
      <UnauthorizedBridge />
      <Routes>
        <Route element={<LoginPage />} path="/login" />
        <Route element={<ProtectedRoute />}>
          <Route element={<AppLayout />}>
            <Route element={<CentralReportingPage />} path="/central/reporting" />
            <Route element={<StoreMonitoringPage />} path="/store/monitoring" />
            <Route element={<RequireCentralAdmin />}>
              <Route element={<CentralUsersPage />} path="/central/users" />
              <Route element={<CreateCentralUserPage />} path="/central/users/new" />
              <Route element={<EditCentralUserPage />} path="/central/users/:userId" />
            </Route>
          </Route>
        </Route>
        <Route element={<Navigate replace to="/central/reporting" />} path="/" />
        <Route element={<Navigate replace to="/central/reporting" />} path="*" />
      </Routes>
    </AuthProvider>
  );
}
