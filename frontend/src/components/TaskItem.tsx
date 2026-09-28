import {Pressable, StyleSheet, Text, View} from 'react-native';

import type {Status, Task} from '../types/task';

/** Warna badge per status — cukup map statis, tanpa sistem tema. */
const STATUS_COLORS: Record<Status, string> = {
  todo: '#6b7280',
  in_progress: '#d97706',
  done: '#16a34a',
};

interface TaskItemProps {
  task: Task;
  onPress: (task: Task) => void;
}

/** Satu baris task di list. Tap di mana pun → buka edit modal. */
export function TaskItem({task, onPress}: TaskItemProps) {
  return (
    <Pressable style={styles.card} onPress={() => onPress(task)}>
      <View style={styles.headerRow}>
        <Text style={styles.title} numberOfLines={1}>
          {task.title}
        </Text>
        <Text style={[styles.badge, {backgroundColor: STATUS_COLORS[task.status]}]}>
          {task.status}
        </Text>
      </View>
      {task.assignee !== null && <Text style={styles.meta}>Assignee: {task.assignee}</Text>}
      {task.due_date !== null && <Text style={styles.meta}>Due: {task.due_date}</Text>}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: '#fff',
    borderRadius: 8,
    borderWidth: StyleSheet.hairlineWidth,
    borderColor: '#e5e7eb',
    padding: 12,
    marginBottom: 8,
  },
  headerRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: 8,
  },
  title: {
    flexShrink: 1,
    fontSize: 16,
    fontWeight: '600',
    color: '#111827',
  },
  badge: {
    color: '#fff',
    fontSize: 11,
    fontWeight: '600',
    borderRadius: 999,
    overflow: 'hidden',
    paddingHorizontal: 8,
    paddingVertical: 3,
  },
  meta: {
    marginTop: 4,
    fontSize: 13,
    color: '#6b7280',
  },
});
