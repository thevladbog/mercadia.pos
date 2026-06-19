import { Navigate, Route, Routes } from 'react-router-dom';

import { AuthProvider } from './auth/AuthProvider.js';
import { LoginPage } from './auth/LoginPage.js';
import { ProtectedRoute } from './auth/ProtectedRoute.js';
import { UnauthorizedBridge } from './auth/UnauthorizedBridge.js';
import { AppLayout } from './layout/AppLayout.js';
import { CentralReportingPage } from './pages/CentralReportingPage.js';

export function App() {
  return (
    <AuthProvider>
      <UnauthorizedBridge />
      <Routes>
        <Route element={<LoginPage />} path="/login" />
        <Route element={<ProtectedRoute />}>
          <Route element={<AppLayout />}>
            <Route element={<CentralReportingPage />} path="/central/reporting" />
          </Route>
        </Route>
        <Route element={<Navigate replace to="/central/reporting" />} path="/" />
        <Route element={<Navigate replace to="/central/reporting" />} path="*" />
      </Routes>
    </AuthProvider>
  );
}
