const GATEWAY = "http://localhost:8000";

export const api = {
  async request(endpoint, options = {}) {
    const token = localStorage.getItem('token');
  
    const headers = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    // Убираем возможный лишний слеш, если он есть в начале эндпоинта
    const cleanEndpoint = endpoint.startsWith('/') ? endpoint : `/${endpoint}`;
    
    const response = await fetch(`${GATEWAY}${cleanEndpoint}`, {
      ...options,
      headers,
    });

    if (response.status === 401) {
      window.dispatchEvent(new Event('auth-expired'));
      throw new Error('Unauthorized');
    }
    
    const data = await response.json();
    if (!response.ok) throw new Error(data.error || 'Ошибка API');
    return data;
  },
  
  get(e) { return this.request(e); },
  post(e, b) { return this.request(e, { method: 'POST', body: JSON.stringify(b) }); },
  put(e, b) { return this.request(e, { method: 'PUT', body: JSON.stringify(b) }); },
  patch(e, b) { return this.request(e, { method: 'PATCH', body: JSON.stringify(b) }); },
  delete(e) { return this.request(e, { method: 'DELETE' }); }
};