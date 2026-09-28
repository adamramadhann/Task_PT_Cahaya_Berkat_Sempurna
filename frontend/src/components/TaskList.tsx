import {ActivityIndicator, FlatList, StyleSheet, Text, View} from 'react-native';

import type {Task} from '../types/task';
import {TaskItem} from './TaskItem';

interface TaskListProps {
  tasks: Task[];
  /** Saat fetch berjalan: spinner penuh jika list masih kosong, footer jika sedang berisi. */
  loading: boolean;
  onPressItem: (task: Task) => void;
}

/** FlatList task + spinner loading + empty state. */
export function TaskList({tasks, loading, onPressItem}: TaskListProps) {
  if (loading && tasks.length === 0) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" />
      </View>
    );
  }
  if (tasks.length === 0) {
    return (
      <View style={styles.center}>
        <Text style={styles.empty}>Tidak ada task.</Text>
      </View>
    );
  }
  return (
    <FlatList
      data={tasks}
      keyExtractor={item => String(item.id)}
      renderItem={({item}) => <TaskItem task={item} onPress={onPressItem} />}
      ListFooterComponent={loading ? <ActivityIndicator style={styles.footer} /> : undefined}
      contentContainerStyle={styles.listContent}
    />
  );
}

const styles = StyleSheet.create({
  center: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    padding: 24,
  },
  empty: {
    fontSize: 15,
    color: '#6b7280',
  },
  listContent: {
    paddingHorizontal: 12,
    paddingBottom: 12,
  },
  footer: {
    marginVertical: 12,
  },
});
