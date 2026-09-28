import {render} from '@testing-library/react-native';
import React from 'react';

import App from '../App';

// Native component safe-area tidak merender children di test renderer —
// ganti dengan passthrough sederhana (insets nol).
jest.mock('react-native-safe-area-context', () => ({
  SafeAreaProvider: ({children}: {children?: React.ReactNode}) => children,
  SafeAreaView: ({children}: {children?: React.ReactNode}) => children,
  useSafeAreaInsets: () => ({top: 0, right: 0, bottom: 0, left: 0}),
}));

// Mock API layer: app smoke test tidak menembak server sungguhan.
jest.mock('../src/api/tasks', () => ({
  listTasks: jest.fn().mockResolvedValue({
    data: [],
    meta: {page: 1, limit: 10, total_items: 0, total_pages: 0},
  }),
  updateTask: jest.fn().mockResolvedValue({}),
}));

test('menampilkan layar task list sampai list kosong selesai dimuat', async () => {
  const screen = await render(<App />);
  expect(await screen.findByText('Tidak ada task.')).toBeTruthy();
});
