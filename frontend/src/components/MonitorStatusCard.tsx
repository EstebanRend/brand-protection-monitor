import type { MonitorStatus } from '../types/models';

type MonitorStatusCardProps = {
  status: MonitorStatus | null;
  onRunOnce: () => Promise<void>;
  isRunning: boolean;
};

export function MonitorStatusCard({ status, onRunOnce, isRunning }: MonitorStatusCardProps) {
  const isActive = status?.status === 'active';

  return (
    <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h2 className="text-lg font-semibold text-slate-900">Monitor Status</h2>
          <p className="text-sm text-slate-500">Background worker checks the latest CT log entries periodically.</p>
        </div>
        <span className={`rounded-full px-3 py-1 text-xs font-semibold ${isActive ? 'bg-green-100 text-green-700' : 'bg-amber-100 text-amber-700'}`}>
          {status?.status ?? 'loading'}
        </span>
      </div>

      <div className="mt-5 grid grid-cols-1 gap-3 sm:grid-cols-3">
        <Metric label="Last Tree Size" value={status?.lastTreeSize?.toLocaleString() ?? '-'} />
        <Metric label="Processed Last Cycle" value={status?.processedLastCycle?.toString() ?? '-'} />
        <Metric label="Last Processed" value={status?.lastProcessedAt ? new Date(status.lastProcessedAt).toLocaleString() : '-'} />
      </div>

      <button
        onClick={onRunOnce}
        disabled={isRunning}
        className="mt-5 rounded-xl border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50 disabled:opacity-50"
      >
        {isRunning ? 'Running...' : 'Run Monitor Once'}
      </button>
    </section>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl bg-slate-50 p-4">
      <p className="text-xs font-medium uppercase tracking-wide text-slate-500">{label}</p>
      <p className="mt-1 text-sm font-semibold text-slate-900">{value}</p>
    </div>
  );
}
