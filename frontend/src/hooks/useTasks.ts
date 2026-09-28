import {useCallback, useEffect, useState} from 'react';

import {ApiError} from '../api/client';
import {listTasks} from '../api/tasks';
import type {ListMeta, ListTasksQuery, Status, Task} from '../types/task';

interface UseTasksResult {
  tasks: Task[];
  meta: ListMeta | null;
  /** true selama request list berjalan (fetch awal, ganti filter/halaman, refetch). */
  loading: boolean;
  error: string | null;
  /** Filter status aktif; '' berarti semua. */
  status: Status | '';
  keyword: string;
  /** Terapkan hasil debounce SearchInput — selalu reset ke halaman 1. */
  applySearch: (keyword: string) => void;
  /** Terapkan pilihan StatusFilter — selalu reset ke halaman 1. */
  applyStatus: (status: Status | '') => void;
  changePage: (page: number) => void;
  /** Ambil ulang list dengan query yang sama (retry error, refresh setelah edit). */
  refetch: () => void;
}

/**
 * Satu-satunya sumber data layar (Keputusan #4: plain hooks, tanpa state library).
 * Setiap perubahan query membatalkan request sebelumnya lewat AbortController,
 * sehingga respons basi tidak pernah menimpa yang baru (race-safe).
 */
export function useTasks(): UseTasksResult {
  const [status, setStatus] = useState<Status | ''>('');
  const [keyword, setKeyword] = useState('');
  const [page, setPage] = useState(1);
  const [data, setData] = useState<{tasks: Task[]; meta: ListMeta} | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [reloadKey, setReloadKey] = useState(0);

  useEffect(() => {
    const query: ListTasksQuery = {
      status: status === '' ? undefined : status,
      keyword: keyword === '' ? undefined : keyword,
      page,
    };
    const controller = new AbortController();
    setLoading(true);
    setError(null);
    listTasks(query, controller.signal)
      .then(res => setData({tasks: res.data, meta: res.meta}))
      .catch(err => {
        // Dibatalkan karena digantikan request lebih baru / unmount — abaikan.
        if (controller.signal.aborted) {
          return;
        }
        setError(err instanceof ApiError ? err.message : 'Terjadi kesalahan tak terduga.');
      })
      .finally(() => {
        if (!controller.signal.aborted) {
          setLoading(false);
        }
      });
    return () => controller.abort();
  }, [status, keyword, page, reloadKey]);

  const applySearch = useCallback((kw: string) => {
    setKeyword(kw);
    setPage(1);
  }, []);

  const applyStatus = useCallback((s: Status | '') => {
    setStatus(s);
    setPage(1);
  }, []);

  const changePage = useCallback((p: number) => setPage(p), []);

  const refetch = useCallback(() => setReloadKey(key => key + 1), []);

  return {
    tasks: data?.tasks ?? [],
    meta: data?.meta ?? null,
    loading,
    error,
    status,
    keyword,
    applySearch,
    applyStatus,
    changePage,
    refetch,
  };
}
