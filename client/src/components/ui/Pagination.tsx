import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from './DropdownMenu';
import { Icons } from '../Icons';

interface PaginationProps {
  page: number;
  totalPages: number;
  total: number;
  perPage: number;
  perPageOptions?: number[];
  onPageChange: (page: number) => void;
  onPerPageChange: (perPage: number) => void;
  loading?: boolean;
}

export default function Pagination({
  page,
  totalPages,
  total,
  perPage,
  perPageOptions = [10, 20, 50, 100],
  onPageChange,
  onPerPageChange,
  loading,
}: PaginationProps) {
  const start = (page - 1) * perPage + 1;
  const end = Math.min(page * perPage, total);

  return (
    <div className="flex items-center gap-3 text-xs text-neutral-400">
      <div className="flex items-center gap-2">
        <span className="whitespace-nowrap">Rows</span>
        <div className="w-20">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button
                type="button"
                className="w-full rounded-lg border border-neutral-800 px-3 py-2 text-xs text-neutral-100 transition hover:border-neutral-500 focus:outline-none focus:ring-2 focus:ring-neutral-100 focus:ring-offset-2 focus:ring-offset-[#171717] bg-neutral-800/80 flex items-center justify-between"
              >
                <span className="truncate">{perPage}</span>
                <Icons.selector className="w-4 h-4 text-neutral-500" />
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              {perPageOptions.map((opt) => (
                <DropdownMenuItem key={opt} onSelect={() => onPerPageChange(opt)}>
                  {opt}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      <span className="hidden sm:inline">{start}–{end} of {total}</span>

      <div className="flex items-center gap-2">
        <button
          onClick={() => onPageChange(page - 1)}
          disabled={page <= 1 || loading}
          className="h-9 w-9 rounded-lg inline-flex items-center justify-center text-lg font-medium text-neutral-400 hover:bg-neutral-800/80 hover:text-neutral-100 disabled:opacity-60 disabled:cursor-not-allowed transition-colors"
          aria-label="Previous page"
        >
          ‹
        </button>
        <span className="font-semibold text-neutral-300 min-w-[4rem] text-center">{page} / {totalPages}</span>
        <button
          onClick={() => onPageChange(page + 1)}
          disabled={page >= totalPages || loading}
          className="h-9 w-9 rounded-lg inline-flex items-center justify-center text-lg font-medium text-neutral-400 hover:bg-neutral-800/80 hover:text-neutral-100 disabled:opacity-60 disabled:cursor-not-allowed transition-colors"
          aria-label="Next page"
        >
          ›
        </button>
      </div>
    </div>
  );
}
