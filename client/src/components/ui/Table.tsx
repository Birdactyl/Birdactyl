import { ReactNode, Fragment } from 'react';

interface Column<T> {
  key: string;
  header: ReactNode;
  align?: 'left' | 'right' | 'center';
  render: (item: T) => ReactNode;
  className?: string;
}

interface TableProps<T> {
  columns: Column<T>[];
  data: T[];
  keyField: keyof T;
  loading?: boolean;
  emptyText?: string;
  onRowClick?: (item: T) => void;
  rowClassName?: (item: T) => string;
  expandable?: {
    render: (item: T) => ReactNode;
    isExpanded: (item: T) => boolean;
  };
}

export default function Table<T>({ columns, data, keyField, loading, emptyText = 'No data found', onRowClick, rowClassName, expandable }: TableProps<T>) {
  return (
    <div className="rounded-lg border border-neutral-800 overflow-x-auto">
      <table className="w-full min-w-max">
        <thead className="bg-neutral-900/50">
          <tr>
            {columns.map(col => (
              <th key={col.key} className={`px-4 py-3 text-xs font-medium text-neutral-400 uppercase tracking-wider ${col.align === 'right' ? 'text-right' : col.align === 'center' ? 'text-center' : 'text-left'} ${col.className || ''}`}>
                {col.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className={`bg-neutral-900/50 divide-y divide-neutral-800 transition-opacity duration-150 ${loading ? 'opacity-50' : 'opacity-100'}`}>
          {data.length === 0 ? (
            <tr><td colSpan={columns.length} className="px-4 py-8 text-center text-sm text-neutral-400">{loading ? 'Loading...' : emptyText}</td></tr>
          ) : data.map(item => {
            const key = String(item[keyField]);
            const isExpanded = expandable?.isExpanded(item);
            return (
              <Fragment key={key}>
                <tr onClick={() => onRowClick?.(item)} className={`hover:bg-neutral-800/30 transition-colors ${onRowClick ? 'cursor-pointer' : ''} ${rowClassName?.(item) || ''}`}>
                  {columns.map(col => (
                    <td key={col.key} className={`px-4 py-3 whitespace-nowrap ${col.align === 'right' ? 'text-right' : col.align === 'center' ? 'text-center' : ''}`}>
                      {col.render(item)}
                    </td>
                  ))}
                </tr>
                {expandable && isExpanded && (
                  <tr className="bg-neutral-800/20">
                    <td colSpan={columns.length} className="px-4 py-4">
                      {expandable.render(item)}
                    </td>
                  </tr>
                )}
              </Fragment>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
