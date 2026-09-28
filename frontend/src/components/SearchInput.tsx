import {useEffect, useRef, useState} from 'react';
import {StyleSheet, TextInput} from 'react-native';

const DEBOUNCE_MS = 400;

interface SearchInputProps {
  /** Kata kunci aktif di parent (hasil debounce terakhir). */
  value: string;
  /** Dipanggil SEKALI, 400ms setelah ketikan terakhir — bukan per huruf. */
  onChangeDebounced: (keyword: string) => void;
}

/**
 * TextInput dengan debounce (Keputusan #5): timer di-reset setiap ketikan,
 * jadi mengetik cepat hanya mengirim satu nilai akhir. Timer dibersihkan saat
 * unmount, dan disusun fake-timer friendly agar mudah dites.
 */
export function SearchInput({value, onChangeDebounced}: SearchInputProps) {
  const [text, setText] = useState(value);

  // Callback terbaru tanpa me-reset timer yang sedang berjalan.
  const onChangeRef = useRef(onChangeDebounced);
  useEffect(() => {
    onChangeRef.current = onChangeDebounced;
  });

  useEffect(() => {
    if (text === value) {
      return; // tidak ada ketikan yang masih menunggu dikirim
    }
    const timer = setTimeout(() => onChangeRef.current(text), DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [text, value]);

  return (
    <TextInput
      style={styles.input}
      placeholder="Cari judul task..."
      value={text}
      onChangeText={setText}
      autoCorrect={false}
      autoCapitalize="none"
    />
  );
}

const styles = StyleSheet.create({
  input: {
    borderWidth: 1,
    borderColor: '#d1d5db',
    borderRadius: 8,
    backgroundColor: '#fff',
    paddingHorizontal: 12,
    paddingVertical: 8,
    fontSize: 15,
    marginHorizontal: 12,
    marginBottom: 8,
  },
});
