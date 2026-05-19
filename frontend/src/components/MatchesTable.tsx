import type { MatchedCertificate } from '../types/models';
import { api } from '../api/client';

type MatchesTableProps = {
  matches: MatchedCertificate[];
};

export function MatchesTable({ matches }: MatchesTableProps) {
  return (
    <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div className="mb-4 flex items-center justify-between gap-4">
        <div>
          <h2 className="text-lg font-semibold text-slate-900">Matched Certificates</h2>
          <p className="text-sm text-slate-500">Domains matching your monitored keywords are highlighted.</p>
        </div>
        <a href={api.exportUrl} className="rounded-xl bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700">
          Export CSV
        </a>
      </div>

      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-slate-200 text-sm">
          <thead className="bg-slate-50 text-left text-xs uppercase tracking-wide text-slate-500">
            <tr>
              <th className="px-4 py-3">Domain</th>
              <th className="px-4 py-3">Matched Keyword</th>
              <th className="px-4 py-3">Issuer</th>
              <th className="px-4 py-3">Detected At</th>
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
                <tr key={match.id} className="border-l-4 border-red-500 bg-red-50/70">
                  <td className="max-w-sm px-4 py-3 font-medium text-slate-900">{match.domain}</td>
                  <td className="px-4 py-3">
                    <span className="rounded-full bg-red-100 px-2 py-1 text-xs font-semibold text-red-700">{match.matchedKeyword}</span>
                  </td>
                  <td className="max-w-sm px-4 py-3 text-slate-600">{match.issuer}</td>
                  <td className="px-4 py-3 text-slate-600">{new Date(match.createdAt).toLocaleString()}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </section>
  );
}
