import {StyleSheet, Text, TouchableOpacity, View} from 'react-native';

interface PaginationProps {
  page: number;
  totalPages: number;
  /** true saat list loading — kedua tombol nonaktif. */
  disabled: boolean;
  onChangePage: (page: number) => void;
}

/** Prev/next + posisi halaman; angka diambil dari `meta` respons BE (Fase FE-3). */
export function Pagination({page, totalPages, disabled, onChangePage}: PaginationProps) {
  const prevDisabled = disabled || page <= 1;
  const nextDisabled = disabled || page >= totalPages;
  return (
    <View style={styles.row}>
      <TouchableOpacity
        style={[styles.button, prevDisabled && styles.buttonDisabled]}
        disabled={prevDisabled}
        onPress={() => onChangePage(page - 1)}>
        <Text style={styles.buttonText}>{'‹ Prev'}</Text>
      </TouchableOpacity>
      <Text style={styles.pageLabel}>
        Halaman {page} dari {totalPages}
      </Text>
      <TouchableOpacity
        style={[styles.button, nextDisabled && styles.buttonDisabled]}
        disabled={nextDisabled}
        onPress={() => onChangePage(page + 1)}>
        <Text style={styles.buttonText}>{'Next ›'}</Text>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: 8,
    paddingHorizontal: 12,
    paddingVertical: 8,
    borderTopWidth: StyleSheet.hairlineWidth,
    borderTopColor: '#e5e7eb',
    backgroundColor: '#fff',
  },
  button: {
    borderRadius: 8,
    backgroundColor: '#2563eb',
    paddingHorizontal: 14,
    paddingVertical: 8,
  },
  buttonDisabled: {
    backgroundColor: '#c7d2fe',
  },
  buttonText: {
    color: '#fff',
    fontWeight: '600',
  },
  pageLabel: {
    fontSize: 13,
    color: '#374151',
  },
});
