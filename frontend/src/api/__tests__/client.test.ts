import {buildQuery} from '../client';

// Placeholder DoD Fase FE-0: membuktikan pipeline Jest jalan tanpa mock tambahan.
// Test komponen yang sebenarnya (SearchInput debounce, fake timers) menyusul di Fase FE-5.
describe('buildQuery', () => {
  it('mengembalikan string kosong kalau tidak ada param', () => {
    expect(buildQuery({})).toBe('');
  });

  it('mengabaikan param undefined/null/string kosong', () => {
    expect(buildQuery({status: undefined, keyword: null, page: ''})).toBe('');
  });

  it('menyusun query dari param terisi', () => {
    expect(buildQuery({status: 'done', page: 2})).toBe('?status=done&page=2');
  });

  it('menyandikan spasi pada keyword', () => {
    // URLSearchParams memakai encoding form-urlencoded: spasi → "+".
    // Go (`r.URL.Query()`) di sisi BE meng-decode "+" kembali menjadi spasi.
    expect(buildQuery({keyword: 'login page'})).toBe('?keyword=login+page');
  });
});
