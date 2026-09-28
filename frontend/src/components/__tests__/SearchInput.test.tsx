import {fireEvent, render} from '@testing-library/react-native';

import {SearchInput} from '../SearchInput';

// Task 5 (component test): ketik → debounce 400ms → callback terpanggil
// SEKALI dengan nilai akhir, bukan per huruf (fake timers).
// fireEvent di RNTL 14 async (return Promise) — selalu di-await.
// Advance timer murni memicu callback (tanpa setState), jadi tidak butuh act().
describe('SearchInput', () => {
  afterEach(() => {
    jest.useRealTimers();
  });

  it('memanggil onChangeDebounced sekali setelah 400ms ketikan terakhir', async () => {
    const onChangeDebounced = jest.fn();
    const screen = await render(<SearchInput value="" onChangeDebounced={onChangeDebounced} />);
    const input = screen.getByPlaceholderText('Cari judul task...');

    jest.useFakeTimers(); // render selesai dengan timer asli; sekarang bekukan waktu
    await fireEvent.changeText(input, 'log');
    jest.advanceTimersByTime(300); // belum mencapai 400ms
    expect(onChangeDebounced).not.toHaveBeenCalled();

    await fireEvent.changeText(input, 'login'); // reset timer debounce
    jest.advanceTimersByTime(399);
    expect(onChangeDebounced).not.toHaveBeenCalled();

    jest.advanceTimersByTime(1);
    expect(onChangeDebounced).toHaveBeenCalledTimes(1);
    expect(onChangeDebounced).toHaveBeenCalledWith('login');
  });

  it('membatalkan callback yang tertunda saat unmount', async () => {
    const onChangeDebounced = jest.fn();
    const screen = await render(<SearchInput value="" onChangeDebounced={onChangeDebounced} />);

    await fireEvent.changeText(screen.getByPlaceholderText('Cari judul task...'), 'abc');
    await screen.unmount(); // timer debounce ikut dibersihkan effect cleanup
    jest.useFakeTimers();
    jest.advanceTimersByTime(1000);
    expect(onChangeDebounced).not.toHaveBeenCalled();
  });
});
