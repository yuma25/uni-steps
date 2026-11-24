/**
 * 非同期処理を Go 言語風の [結果, エラー] 形式で安全に処理するためのラッパーである．
 */
export async function handle<T>(promise: Promise<T>): Promise<[T, null] | [null, Error]> {
  try {
    const data = await promise;
    return [data, null];
  } catch (err: unknown) {
    if (err instanceof Error) return [null, err];
    return [null, new Error(String(err))];
  }
}

/**
 * ISO8601 形式の文字列を日本語の日時表示形式にフォーマットする．
 */
export function formatDate(isoString: string): string {
  const date = new Date(isoString);
  if (date.getFullYear() <= 1) return '期限未定';
  return date.toLocaleString('ja-JP', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

/**
 * datetime-local インプット用のフォーマット文字列（YYYY-MM-DDTHH:mm）を生成する．
 */
export function toLocalISOString(date: Date): string {
  if (date.getFullYear() <= 1) return '';
  return new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 16);
}
