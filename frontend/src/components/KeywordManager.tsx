import { FormEvent, useState } from 'react';
import type { Keyword } from '../types/models';

type KeywordManagerProps = {
  keywords: Keyword[];
  onAdd: (value: string) => Promise<void>;
  onDelete: (id: number) => Promise<void>;
};

export function KeywordManager({ keywords, onAdd, onDelete }: KeywordManagerProps) {
  const [value, setValue] = useState('');
  const [isSaving, setIsSaving] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const trimmedValue = value.trim();
    if (!trimmedValue) return;

    setIsSaving(true);
    try {
      await onAdd(trimmedValue);
      setValue('');
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div className="mb-4">
        <h2 className="text-lg font-semibold text-slate-900">Monitored Keywords</h2>
        <p className="text-sm text-slate-500">Add brands or suspicious words to detect in certificate domains.</p>
      </div>

      <form onSubmit={handleSubmit} className="flex gap-2">
        <input
          value={value}
          onChange={(event) => setValue(event.target.value)}
          placeholder="example: paypal, microsoft, acme"
          className="flex-1 rounded-xl border border-slate-300 px-4 py-2 text-sm outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
        />
        <button
          disabled={isSaving}
          className="rounded-xl bg-slate-900 px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
        >
          {isSaving ? 'Adding...' : 'Add'}
        </button>
      </form>

      <div className="mt-4 flex flex-wrap gap-2">
        {keywords.length === 0 ? (
          <p className="text-sm text-slate-500">No keywords yet. Add one to start matching certificates.</p>
        ) : (
          keywords.map((keyword) => (
            <span key={keyword.id} className="inline-flex items-center gap-2 rounded-full bg-slate-100 px-3 py-1 text-sm text-slate-700">
              {keyword.value}
              <button onClick={() => onDelete(keyword.id)} className="text-slate-400 hover:text-red-600" aria-label={`Delete ${keyword.value}`}>
                ×
              </button>
            </span>
          ))
        )}
      </div>
    </section>
  );
}
