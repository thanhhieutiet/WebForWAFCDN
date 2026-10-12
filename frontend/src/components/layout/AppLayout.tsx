import React from 'react';
import { Outlet } from 'react-router-dom';
import { Navbar } from './Navbar';
import { EnvironmentBar } from './EnvironmentBar';
import { Footer } from './Footer';
import { LabConsole } from '@/components/lab/LabConsole';
import { ErrorBoundary } from '@/components/common/ErrorBoundary';

export function AppLayout(): React.JSX.Element {
  return (
    <div className="flex min-h-screen flex-col bg-slate-50 text-slate-900 dark:bg-slate-950 dark:text-slate-100 font-sans selection:bg-teal-100 selection:text-teal-900">
      <EnvironmentBar />
      <Navbar />
      <main className="flex-1 pb-16">
        <ErrorBoundary>
          <Outlet />
        </ErrorBoundary>
      </main>
      <Footer />
      <LabConsole />
    </div>
  );
}
