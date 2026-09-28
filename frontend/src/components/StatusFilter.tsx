import {StyleSheet, Text, TouchableOpacity, View} from 'react-native';

import type {Status} from '../types/task';

/** Opsi chip: "Semua" (tanpa filter) + tiga status yang dikenal BE. */
const OPTIONS: ReadonlyArray<{value: Status | ''; label: string}> = [
  {value: '', label: 'Semua'},
  {value: 'todo', label: 'Todo'},
  {value: 'in_progress', label: 'In Progress'},
  {value: 'done', label: 'Done'},
];

interface StatusFilterProps<T extends Status | ''> {
  value: T;
  onChange: (status: T) => void;
  /** false di edit modal — status di sana wajib, opsi "Semua" tidak berlaku. */
  withAllOption?: boolean;
}

/**
 * Baris chip status — TouchableOpacity bawaan, tanpa package picker (Keputusan #7).
 * Generic pada `T`: list memakai `Status | ''` (ada "Semua"), edit modal `Status`.
 */
export function StatusFilter<T extends Status | ''>({
  value,
  onChange,
  withAllOption = true,
}: StatusFilterProps<T>) {
  const options = withAllOption ? OPTIONS : OPTIONS.filter(option => option.value !== '');
  return (
    <View style={styles.row}>
      {options.map(option => {
        const active = option.value === value;
        return (
          <TouchableOpacity
            key={option.label}
            style={[styles.chip, active && styles.chipActive]}
            onPress={() => onChange(option.value as T)}>
            <Text style={[styles.label, active && styles.labelActive]}>{option.label}</Text>
          </TouchableOpacity>
        );
      })}
    </View>
  );
}

const styles = StyleSheet.create({
  row: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
    marginHorizontal: 12,
    marginBottom: 8,
  },
  chip: {
    borderRadius: 999,
    backgroundColor: '#e5e7eb',
    paddingHorizontal: 12,
    paddingVertical: 6,
  },
  chipActive: {
    backgroundColor: '#2563eb',
  },
  label: {
    fontSize: 13,
    fontWeight: '600',
    color: '#374151',
  },
  labelActive: {
    color: '#fff',
  },
});
