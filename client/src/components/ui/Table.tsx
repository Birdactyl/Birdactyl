import { ReactNode, Fragment, useRef } from 'react';
import { ContextMenuZone } from './ContextMenu';
import { useVirtualizer } from '@tanstack/react-virtual';

interface ContextMenuItem {
  label: ReactNode;
  icon?: ReactNode;
  onClick?: () => void;
  variant?: 'default' | 'danger';
  disabled?: boolean;
}

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
  contextMenu?: (item: T) => (ContextMenuItem | 'separator')[];
  virtualize?: boolean;
}

function RowCells<T>({ item, columns }: { item: T; columns: Column<T>[] }) {
  return (
    <>
      {columns.map(col => (
        <td key={col.key} className={`px-4 py-3 whitespace-nowrap ${col.align === 'right' ? 'text-right' : col.align === 'center' ? 'text-center' : ''}`}>
          {col.render(item)}
        </td>
      ))}
    </>
  );
}

export default function Table<T>({ columns, data, keyField, loading, emptyText = 'No data found', onRowClick, rowClassName, expandable, contextMenu, virtualize = true }: TableProps<T>) {
  const parentRef = useRef<HTMLDivElement>(null);

  const virtualizer = useVirtualizer({
    count: data.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 56,
    overscan: 10,
  });

  const virtualItems = virtualizer.getVirtualItems();
  const shouldVirtualize = virtualize && data.length > 50;

  const displayItems = shouldVirtualize ? virtualItems.map(vi => ({ item: data[vi.index], index: vi.index, virtualRow: vi })) : data.map((item, index) => ({ item, index, virtualRow: undefined }));

  const paddingTop = shouldVirtualize && virtualItems.length > 0 ? virtualItems[0]?.start || 0 : 0;
  const paddingBottom = shouldVirtualize && virtualItems.length > 0 ? virtualizer.getTotalSize() - (virtualItems[virtualItems.length - 1]?.end || 0) : 0;

  return (
    <div ref={parentRef} className="rounded-lg border border-neutral-800 overflow-y-auto overflow-x-auto max-h-[70vh] relative min-h-[300px]">
      <table className="w-full min-w-full relative">
        <thead className="bg-neutral-900/90 backdrop-blur top-0 sticky z-10">
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
          ) : (
            <>
              {paddingTop > 0 && <tr><td style={{ height: `${paddingTop}px` }} colSpan={columns.length} className="p-0 border-0" /></tr>}
              {displayItems.map(({ item, index }) => {
                const key = String(item[keyField]);
                const isExpanded = expandable?.isExpanded(item);
                const menuItems = contextMenu?.(item);
                const rowClass = `hover:bg-neutral-800/30 transition-colors ${onRowClick ? 'cursor-pointer' : ''} ${rowClassName?.(item) || ''}`;

                const rowRef = shouldVirtualize ? virtualizer.measureElement : undefined;

                return (
                  <Fragment key={key}>
                    {menuItems && menuItems.length > 0 ? (
                      <ContextMenuZone
                        as="tr"
                        ref={rowRef}
                        data-index={index}
                        items={menuItems}
                        className={rowClass}
                        onClick={onRowClick ? () => onRowClick(item) : undefined}
                      >
                        <RowCells item={item} columns={columns} />
                      </ContextMenuZone>
                    ) : (
                      <tr
                        ref={rowRef as any}
                        data-index={index}
                        onClick={() => onRowClick?.(item)}
                        className={rowClass}
                      >
                        <RowCells item={item} columns={columns} />
                      </tr>
                    )}
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
              {paddingBottom > 0 && <tr><td style={{ height: `${paddingBottom}px` }} colSpan={columns.length} className="p-0 border-0" /></tr>}
            </>
          )}
        </tbody>
      </table>
    </div>
  );
}
