const GATEWAY = "http://localhost:8000";

export const api = {
  async request(endpoint, options = {}) {
    const token = localStorage.getItem('token');
    const headers = { 'Content-Type': 'application/json', ...options.headers };
    if (token) headers['Authorization'] = `Bearer ${token}`;

    const response = await fetch(`${GATEWAY}${endpoint}`, { ...options, headers });
    if (response.status === 401) {
      localStorage.removeItem('token');
      window.dispatchEvent(new Event('auth-expired'));
      return null;
    }
    const data = await response.json();
    if (!response.ok) throw new Error(data.error || 'Ошибка API');
    return data;
  },
  get(e) { return this.request(e); },
  post(e, b) { return this.request(e, { method: 'POST', body: JSON.stringify(b) }); }
};