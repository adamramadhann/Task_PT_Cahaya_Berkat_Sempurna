import {useEffect, useState} from 'react';
import {
  ActivityIndicator,
  KeyboardAvoidingView,
  Modal,
  Platform,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View,
} from 'react-native';

import {ApiError} from '../api/client';
import {updateTask} from '../api/tasks';
import type {Status, Task} from '../types/task';
import {StatusFilter} from './StatusFilter';

interface EditTaskModalProps {
  /** Task yang sedang diedit; null = modal tertutup. */
  task: Task | null;
  onClose: () => void;
  /** Dipanggil setelah update sukses — parent me-refetch list (bug fix "refresh list after update"). */
  onSaved: () => void;
}

/**
 * Modal edit (Modal bawaan RN, Keputusan #8). Form terisi langsung dari list
 * item — tidak ada GET by id. Error server (409/400) tampil di dalam modal,
 * modal tidak tertutup sampai simpan berhasil.
 */
export function EditTaskModal({task, onClose, onSaved}: EditTaskModalProps) {
  const [title, setTitle] = useState('');
  const [status, setStatus] = useState<Status>('todo');
  const [assignee, setAssignee] = useState('');
  const [dueDate, setDueDate] = useState('');
  const [validationError, setValidationError] = useState<string | null>(null);
  const [serverError, setServerError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  // Prefill ulang setiap modal dibuka untuk task tertentu.
  useEffect(() => {
    if (task !== null) {
      setTitle(task.title);
      setStatus(task.status);
      setAssignee(task.assignee ?? '');
      setDueDate(task.due_date ?? '');
      setValidationError(null);
      setServerError(null);
      setSaving(false);
    }
  }, [task]);

  if (task === null) {
    return null;
  }

  const save = async () => {
    if (title.trim() === '') {
      setValidationError('Title wajib diisi.');
      return;
    }
    setValidationError(null);
    setServerError(null);
    setSaving(true);
    try {
      await updateTask(task.id, {
        // PUT = full update: field yang tidak diedit di modal tetap dipertahankan.
        title: title.trim(),
        status,
        description: task.description,
        assignee: assignee.trim() === '' ? null : assignee.trim(),
        due_date: dueDate.trim() === '' ? null : dueDate.trim(),
      });
      onSaved();
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.code === 'DUPLICATE_TITLE') {
          setServerError('Title sudah dipakai task lain.');
        } else {
          const fieldMessage = err.fields
            ? Object.values(err.fields).join(' ')
            : err.message;
          setServerError(fieldMessage);
        }
      } else {
        setServerError('Terjadi kesalahan tak terduga.');
      }
    } finally {
      setSaving(false);
    }
  };

  return (
    <Modal animationType="slide" onRequestClose={onClose}>
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        style={styles.overlay}>
        <View style={styles.sheet}>
          <Text style={styles.heading}>Edit Task #{task.id}</Text>
          {serverError !== null && <Text style={styles.errorText}>{serverError}</Text>}
          <Text style={styles.fieldLabel}>Title</Text>
          <TextInput
            style={styles.input}
            value={title}
            onChangeText={setTitle}
            placeholder="Title task"
          />
          {validationError !== null && <Text style={styles.errorText}>{validationError}</Text>}
          <Text style={styles.fieldLabel}>Status</Text>
          <StatusFilter value={status} onChange={setStatus} withAllOption={false} />
          <Text style={styles.fieldLabel}>Assignee</Text>
          <TextInput
            style={styles.input}
            value={assignee}
            onChangeText={setAssignee}
            placeholder="Opsional"
          />
          <Text style={styles.fieldLabel}>Due date (YYYY-MM-DD)</Text>
          <TextInput
            style={styles.input}
            value={dueDate}
            onChangeText={setDueDate}
            placeholder="Opsional"
            autoCapitalize="none"
          />
          <View style={styles.buttonRow}>
            <TouchableOpacity
              style={[styles.button, styles.buttonSecondary]}
              onPress={onClose}
              disabled={saving}>
              <Text style={styles.buttonSecondaryText}>Batal</Text>
            </TouchableOpacity>
            <TouchableOpacity
              style={[styles.button, styles.buttonPrimary, saving && styles.buttonDisabled]}
              onPress={save}
              disabled={saving}>
              {saving ? (
                <ActivityIndicator color="#fff" />
              ) : (
                <Text style={styles.buttonPrimaryText}>Simpan</Text>
              )}
            </TouchableOpacity>
          </View>
        </View>
      </KeyboardAvoidingView>
    </Modal>
  );
}

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: '#f9fafb',
  },
  sheet: {
    flex: 1,
    padding: 16,
  },
  heading: {
    fontSize: 20,
    fontWeight: '700',
    color: '#111827',
    marginBottom: 12,
  },
  fieldLabel: {
    fontSize: 13,
    fontWeight: '600',
    color: '#374151',
    marginBottom: 4,
  },
  input: {
    borderWidth: 1,
    borderColor: '#d1d5db',
    borderRadius: 8,
    backgroundColor: '#fff',
    paddingHorizontal: 12,
    paddingVertical: 8,
    fontSize: 15,
    marginBottom: 12,
  },
  errorText: {
    color: '#b91c1c',
    backgroundColor: '#fee2e2',
    borderRadius: 8,
    padding: 8,
    marginBottom: 8,
  },
  buttonRow: {
    flexDirection: 'row',
    justifyContent: 'flex-end',
    gap: 8,
    marginTop: 8,
  },
  button: {
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 10,
    minWidth: 96,
    alignItems: 'center',
  },
  buttonPrimary: {
    backgroundColor: '#2563eb',
  },
  buttonSecondary: {
    backgroundColor: '#e5e7eb',
  },
  buttonDisabled: {
    opacity: 0.6,
  },
  buttonPrimaryText: {
    color: '#fff',
    fontWeight: '600',
  },
  buttonSecondaryText: {
    color: '#374151',
    fontWeight: '600',
  },
});
