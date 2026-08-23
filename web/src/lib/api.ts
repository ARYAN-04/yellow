export async function fetchAPI(url: string, method = 'GET', body: any = null, isMultipart = false) {
  const options: RequestInit = { method };
  if (body) {
    if (isMultipart) {
      options.body = body;
    } else {
      options.headers = { 'Content-Type': 'application/json' };
      options.body = JSON.stringify(body);
    }
  }

  const res = await fetch(url, options);
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    const err: any = new Error(errorData.error || `Request failed with status ${res.status}`);
    err.data = errorData;
    throw err;
  }
  if (res.status === 244 || res.status === 204) {
    return null;
  }
  return res.json();
}

export interface AdminContext {
  slug: string;
  isReadOnly: boolean;
}

export interface RoundContext extends AdminContext {
  roundId: string;
  round: any | null;
}
