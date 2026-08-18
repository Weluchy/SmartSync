// Адрес API-шлюза: из переменной окружения или по умолчанию
const GATEWAY = import.meta.env.VITE_GATEWAY_URL || "/api";
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

    const cleanEndpoint = endpoint.startsWith('/') ? endpoint : `/${endpoint}`;
    
    let response;
    try {
      response = await fetch(`${GATEWAY}${cleanEndpoint}`, {
        ...options,
        headers,
      });
    } catch {
      throw new Error('Сервер недоступен. Проверьте подключение к сети.');
    }

    if (response.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('smartsync_user_positions');
      window.dispatchEvent(new Event('auth-expired'));
      throw new Error('Сессия истекла. Войдите снова.');
    }
    
    // Для 204 No Content (например, DELETE)
    if (response.status === 204) {
      return null;
    }

    let data;
    try {
      data = await response.json();
    } catch {
      throw new Error('Ошибка формата ответа от сервера');
    }
    
    if (!response.ok) throw new Error(data.error || 'Ошибка API');
    return data;
  },
  
  get(e) { return this.request(e); },
  post(e, b) { return this.request(e, { method: 'POST', body: JSON.stringify(b) }); },
  put(e, b) { return this.request(e, { method: 'PUT', body: JSON.stringify(b) }); },
  patch(e, b) { return this.request(e, { method: 'PATCH', body: JSON.stringify(b) }); },
  delete(e) { return this.request(e, { method: 'DELETE' }); }
};