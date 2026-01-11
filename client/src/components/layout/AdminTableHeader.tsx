import { Button, Input, Icons } from '../';

interface Props {
  title: string;
  total: number;
  selectedCount: number;
  searchInput: string;
  onSearchInputChange: (value: string) => void;
  onSearch: (e: React.FormEvent) => void;
  onCreateClick: () => void;
  createLabel?: string;
  bulkActions?: { label: string; onClick: () => void; icon?: keyof typeof Icons; variant?: 'danger' | 'warning' }[];
  filters?: { label: string; value: string; active: boolean; onClick: () => void }[];
}

export default function AdminTableHeader({
  title,
  total,
  selectedCount,
  searchInput,
  onSearchInputChange,
  onSearch,
  onCreateClick,
  createLabel = 'Create',
  bulkActions,
  filters,
}: Props) {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-neutral-100">{title}</h1>
          <p className="text-sm text-neutral-400 mt-1">{total} total</p>
        </div>
        <Button onClick={onCreateClick}>
          <Icons.plus className="w-4 h-4" />
          {createLabel}
        </Button>
      </div>

      <div className="flex items-center gap-3">
        <form onSubmit={onSearch} className="flex-1">
          <Input
            value={searchInput}
            onChange={e => onSearchInputChange(e.target.value)}
            placeholder="Search..."
            className="w-full"
          />
        </form>

        {filters && filters.length > 0 && (
          <div className="flex items-center gap-2">
            {filters.map(f => (
              <button
                key={f.value}
                onClick={f.onClick}
                className={`px-3 py-2 text-sm font-medium rounded-lg transition-colors ${
                  f.active ? 'bg-neutral-800 text-neutral-100' : 'text-neutral-400 hover:text-neutral-100 hover:bg-neutral-800/50'
                }`}
              >
                {f.label}
              </button>
            ))}
          </div>
        )}
      </div>

      {selectedCount > 0 && bulkActions && (
        <div className="flex items-center gap-2 p-3 rounded-lg bg-neutral-800/50 border border-neutral-700">
          <span className="text-sm text-neutral-300">{selectedCount} selected</span>
          <div className="flex-1" />
          {bulkActions.map((action, i) => {
            const Icon = action.icon ? Icons[action.icon] : null;
            return (
              <button
                key={i}
                onClick={action.onClick}
                className={`inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium rounded-lg transition-colors ${
                  action.variant === 'danger'
                    ? 'text-red-400 hover:bg-red-500/10'
                    : action.variant === 'warning'
                    ? 'text-amber-400 hover:bg-amber-500/10'
                    : 'text-neutral-300 hover:bg-neutral-700'
                }`}
              >
                {Icon && <Icon className="w-4 h-4" />}
                {action.label}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
