import type { MatchedCertificate } from '../types/models';
import { api } from '../api/client';

type MatchesTableProps = {
  matches: MatchedCertificate[];
};

export function MatchesTable({ matches }: MatchesTableProps) {
  const formattedCount = matches.length.toLocaleString();

  return (
    <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div className="mb-4 flex items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <h2 className="text-lg font-semibold text-slate-900">Matched Certificates</h2>
            <span className="rounded-full bg-slate-100 px-2.5 py-1 text-xs font-semibold text-slate-700">{formattedCount}</span>
          </div>
          <p className="text-sm text-slate-500">Potential lookalike domains from monitored CT entries.</p>
        </div>
        <a
          href={api.exportUrl}
          className="shrink-0 rounded-xl bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          Export CSV
        </a>
      </div>

      <div className="overflow-hidden rounded-xl border border-slate-200">
        <div className="overflow-x-auto">
          <table className="min-w-full table-fixed divide-y divide-slate-200 text-sm">
          <thead className="bg-slate-50 text-left text-xs uppercase tracking-wide text-slate-500">
            <tr>
              <th className="w-[42%] px-4 py-3">Domain</th>
              <th className="w-[16%] px-4 py-3">Keyword</th>
              <th className="w-[24%] px-4 py-3">Issuer</th>
              <th className="w-[18%] px-4 py-3">Detected</th>
            </tr>
          </thead>
            <tbody className="divide-y divide-slate-100">
            {matches.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-4 py-8 text-center text-slate-500">
                  No matches yet. Add keywords and run the monitor.
                </td>
              </tr>
            ) : (
              matches.map((match) => (
                <tr key={match.id} className="align-top hover:bg-rose-50/40">
                  <td className="px-4 py-3">
                    <p className="break-all font-medium text-slate-900" title={match.domain}>
                      {match.domain}
                    </p>
                    <p className="mt-1 truncate text-xs text-slate-500" title={match.sourceLog}>
                      Source: {match.sourceLog}
                    </p>
                  </td>
                  <td className="px-4 py-3">
                    <span className="inline-flex rounded-full bg-rose-100 px-2.5 py-1 text-xs font-semibold text-rose-700">
                      {match.matchedKeyword}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-slate-600">
                    <p className="break-words" title={match.issuer}>
                      {match.issuer}
                    </p>
                  </td>
                  <td className="px-4 py-3 text-slate-600">
                    <DetectedTimestamp value={match.createdAt} />
                  </td>
                </tr>
              ))
            )}
            </tbody>
          </table>
        </div>
      </div>
    </section>
  );
}

function DetectedTimestamp({ value }: { value: string }) {
  const timestamp = new Date(value);
  return (
    <div className="space-y-0.5">
      <p className="font-medium text-slate-700">{timestamp.toLocaleDateString()}</p>
      <p className="text-xs text-slate-500">{timestamp.toLocaleTimeString()}</p>
    </div>
  );
}
