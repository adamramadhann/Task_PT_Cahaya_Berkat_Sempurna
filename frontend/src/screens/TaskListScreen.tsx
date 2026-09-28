import {useState} from 'react';
import {StyleSheet, Text, TouchableOpacity, View} from 'react-native';
import {SafeAreaView} from 'react-native-safe-area-context';

import {EditTaskModal} from '../components/EditTaskModal';
import {Pagination} from '../components/Pagination';
import {SearchInput} from '../components/SearchInput';
import {StatusFilter} from '../components/StatusFilter';
import {TaskList} from '../components/TaskList';
import type {Task} from '../types/task';
import {useTasks} from '../hooks/useTasks';

/**
 * Satu-satunya layar (Keputusan #3): komposisi search, filter, list,
 * pagination, dan edit modal. Tanpa react-navigation.
 */
export function TaskListScreen() {
  const {
    tasks,
    meta,
    loading,
    error,
    status,
    keyword,
    applySearch,
    applyStatus,
    changePage,
    refetch,
  } = useTasks();
  const [editing, setEditing] = useState<Task | null>(null);

  return (
    <SafeAreaView style={styles.container} edges={['top', 'bottom']}>
      <Text style={styles.heading}>Tasks</Text>
      <SearchInput value={keyword} onChangeDebounced={applySearch} />
      <StatusFilter value={status} onChange={applyStatus} />
      {error !== null && (
        <View style={styles.errorBanner}>
          <Text style={styles.errorBannerText}>{error}</Text>
          <TouchableOpacity style={styles.retryButton} onPress={refetch}>
            <Text style={styles.retryButtonText}>Coba lagi</Text>
          </TouchableOpacity>
        </View>
      )}
      <TaskList tasks={tasks} loading={loading} onPressItem={setEditing} />
      {meta !== null && meta.total_pages > 0 && (
        <Pagination
          page={meta.page}
          totalPages={meta.total_pages}
          disabled={loading}
          onChangePage={changePage}
        />
      )}
      <EditTaskModal
        task={editing}
        onClose={() => setEditing(null)}
        onSaved={() => {
          setEditing(null);
          refetch(); // bug fix "refresh list after update"
        }}
      />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f9fafb',
  },
  heading: {
    fontSize: 24,
    fontWeight: '700',
    color: '#111827',
    marginHorizontal: 12,
    marginTop: 8,
    marginBottom: 12,
  },
  errorBanner: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: 8,
    backgroundColor: '#fee2e2',
    borderRadius: 8,
    paddingVertical: 8,
    paddingHorizontal: 12,
    marginHorizontal: 12,
    marginBottom: 8,
  },
  errorBannerText: {
    flexShrink: 1,
    color: '#b91c1c',
  },
  retryButton: {
    paddingHorizontal: 10,
    paddingVertical: 6,
    borderRadius: 6,
    backgroundColor: '#b91c1c',
  },
  retryButtonText: {
    color: '#fff',
    fontWeight: '600',
  },
});
