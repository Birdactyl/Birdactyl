import { useState, useRef, useEffect, useMemo } from 'react';
import { startLoading, finishLoading } from '../lib/pageLoader';
import { notify } from '../components/feedback/Notification';

type ServerFetchFn<T, F> = (page: number, perPage: number, search: string, filter: F) => Promise<{
  success: boolean;
  data?: { items?: T[]; users?: T[]; page: number; total_pages: number; total: number };
  error?: string;
}>;

type ClientFetchFn<T> = () => Promise<{ success: boolean; data?: T[]; error?: string }>;

interface ServerModeOptions<T, F> {
  mode: 'server';
  fetchFn: ServerFetchFn<T, F>;
  defaultFilter: F;
  defaultPerPage?: number;
  itemsKey?: 'items' | 'users';
}

interface ClientModeOptions<T, F> {
  mode: 'client';
  fetchFn: ClientFetchFn<T>;
  filterFn: (item: T, search: string, filter: F) => boolean;
  defaultFilter: F;
  defaultPerPage?: number;
}

type UseTableOptions<T, F> = ServerModeOptions<T, F> | ClientModeOptions<T, F>;

export function useTable<T extends { id: string }, F = string>(options: UseTableOptions<T, F>) {
  const { defaultFilter, defaultPerPage = 20 } = options;
  const isServer = options.mode === 'server';

  const [allItems, setAllItems] = useState<T[]>([]);
  const [page, setPageState] = useState(1);
  const [perPage, setPerPage] = useState(defaultPerPage);
  const [serverTotalPages, setServerTotalPages] = useState(1);
  const [serverTotal, setServerTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [ready, setReady] = useState(false);
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [filter, setFilter] = useState<F>(defaultFilter);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const requestId = useRef(0);

  const filtered = useMemo(() => {
    if (isServer) return allItems;
    const opts = options as ClientModeOptions<T, F>;
    return allItems.filter(item => opts.filterFn(item, search, filter));
  }, [allItems, search, filter, isServer]);

  const clientTotalPages = Math.ceil(filtered.length / perPage) || 1;
  const pageItems = isServer ? allItems : filtered.slice((page - 1) * perPage, page * perPage);
  const totalPages = isServer ? serverTotalPages : clientTotalPages;
  const total = isServer ? serverTotal : filtered.length;

  const loadServer = async (p: number, pp: number, s: string, f: F, initial = false) => {
    const opts = options as ServerModeOptions<T, F>;
    const currentRequest = ++requestId.current;
    setLoading(true);
    setSelected(new Set());
    const res = await opts.fetchFn(p, pp, s, f);
    if (currentRequest !== requestId.current) return;
    if (res.success && res.data) {
      const itemsList = opts.itemsKey === 'users' ? res.data.users : res.data.items;
      setAllItems((itemsList || []) as T[]);
      setPageState(res.data.page);
      setServerTotalPages(res.data.total_pages);
      setServerTotal(res.data.total);
    } else {
      notify('Error', res.error || 'Failed to load data', 'error');
    }
    setLoading(false);
    if (initial) { setReady(true); finishLoading(); }
  };

  const loadClient = async (initial = false) => {
    const opts = options as ClientModeOptions<T, F>;
    setLoading(true);
    const res = await opts.fetchFn();
    if (res.success && res.data) setAllItems(res.data as T[]);
    setLoading(false);
    if (initial) { setReady(true); finishLoading(); }
  };

  useEffect(() => {
    startLoading();
    if (isServer) {
      loadServer(1, perPage, '', defaultFilter, true);
    } else {
      loadClient(true);
    }
  }, []);

  const setPage = (p: number) => {
    if (isServer) {
      loadServer(p, perPage, search, filter);
    } else {
      setPageState(p);
    }
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setSearch(searchInput);
    setSelected(new Set());
    if (isServer) {
      loadServer(1, perPage, searchInput, filter);
    } else {
      setPageState(1);
    }
  };

  const handlePerPageChange = (pp: number) => {
    setPerPage(pp);
    if (isServer) {
      loadServer(1, pp, search, filter);
    } else {
      setPageState(1);
    }
  };

  const handleFilterChange = (f: F) => {
    setFilter(f);
    setSelected(new Set());
    if (isServer) {
      loadServer(1, perPage, search, f);
    } else {
      setPageState(1);
    }
  };

  const reload = () => {
    if (isServer) {
      loadServer(page, perPage, search, filter);
    } else {
      loadClient();
    }
  };

  const toggleSelect = (id: string) => {
    setSelected(prev => {
      const next = new Set(prev);
      next.has(id) ? next.delete(id) : next.add(id);
      return next;
    });
  };

  const toggleSelectAll = () => {
    setSelected(prev => prev.size === pageItems.length ? new Set() : new Set(pageItems.map(i => i.id)));
  };

  const clearSelection = () => setSelected(new Set());

  const allSelected = pageItems.length > 0 && selected.size === pageItems.length;
  const someSelected = selected.size > 0 && selected.size < pageItems.length;

  return {
    items: pageItems,
    allItems,
    setAllItems,
    filtered,
    page,
    setPage,
    perPage,
    totalPages,
    total,
    loading,
    ready,
    search,
    searchInput,
    setSearchInput,
    filter,
    selected,
    selectedItems: allItems.filter(i => selected.has(i.id)),
    handleSearch,
    handlePerPageChange,
    handleFilterChange,
    reload,
    toggleSelect,
    toggleSelectAll,
    allSelected,
    someSelected,
    clearSelection,
  };
}
