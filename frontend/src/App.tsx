import { useEffect, useState } from 'react';
import { api } from './api/client';
import { KeywordManager } from './components/KeywordManager';
import { MatchesTable } from './components/MatchesTable';
import { MonitorStatusCard } from './components/MonitorStatusCard';
import type { Keyword, MatchedCertificate, MonitorStatus } from './types/models';

export default function App() {
  const [keywords, setKeywords] = useState<Keyword[]>([]);
  const [matches, setMatches] = useState<MatchedCertificate[]>([]);
  const [status, setStatus] = useState<MonitorStatus | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isRunning, setIsRunning] = useState(false);

  async function loadDashboard() {
    try {
      const [keywordData, matchData, statusData] = await Promise.all([
        api.listKeywords(),
        api.listMatches(),
        api.getStatus(),
      ]);
      setKeywords(keywordData);
      setMatches(matchData);
      setStatus(statusData);
      setError(null);
    } catch (unknownError) {
      const message = unknownError instanceof Error ? unknownError.message : 'Failed to load dashboard';
      setError(message);
    }
  }

  useEffect(() => {
    void loadDashboard();
    const intervalId = window.setInterval(() => void loadDashboard(), 10000);
    return () => window.clearInterval(intervalId);
  }, []);

  async function handleAddKeyword(value: string) {
    await api.createKeyword(value);
    await loadDashboard();
  }

  async function handleDeleteKeyword(id: number) {
    await api.deleteKeyword(id);
    await loadDashboard();
  }

  async function handleRunOnce() {
    setIsRunning(true);
    try {
      await api.runOnce();
      await loadDashboard();
    } finally {
      setIsRunning(false);
    }
  }

  return (
    <main className="min-h-screen px-6 py-8 text-slate-900">
      <div className="mx-auto max-w-7xl space-y-6">
        <header className="rounded-3xl bg-slate-950 p-8 text-white shadow-sm">
          <p className="text-sm font-semibold uppercase tracking-wide text-blue-300">Certificate Transparency Monitor</p>
          <h1 className="mt-2 text-3xl font-bold">Brand Protection Dashboard</h1>
          <p className="mt-3 max-w-3xl text-slate-300">
            Monitor recent CT log entries, match suspicious domains against configurable keywords, and export findings for review.
          </p>
        </header>

        {error && <div className="rounded-2xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</div>}

        <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <KeywordManager keywords={keywords} onAdd={handleAddKeyword} onDelete={handleDeleteKeyword} />
          <MonitorStatusCard status={status} onRunOnce={handleRunOnce} isRunning={isRunning} />
        </div>

        <MatchesTable matches={matches} />
      </div>
    </main>
  );
}
