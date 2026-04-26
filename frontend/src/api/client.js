const GATEWAY = "http://localhost:8000";

export const api = {
  // Базовая функция запроса с автоматическим подкидыванием токена
  async request(endpoint, options = {}) {
    const token = localStorage.getItem('token');
    const headers = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const config = {
      ...options,
      headers,
    };

    const response = await fetch(`${GATEWAY}${endpoint}`, config);
    
    // Если токен протух — автоматически разлогиниваем
    if (response.status === 401) {
      localStorage.removeItem('token');
      window.dispatchEvent(new Event('auth-expired'));
      return null;
    }

    const data = await response.json();
    if (!response.ok) {
      throw new Error(data.error || 'Произошла ошибка API');
    }

    return data;
  },

  // Удобные методы-обертки
  get(endpoint) { return this.request(endpoint); },
  post(endpoint, body) { return this.request(endpoint, { method: 'POST', body: JSON.stringify(body) }); },
  put(endpoint, body) { return this.request(endpoint, { method: 'PUT', body: JSON.stringify(body) }); },
  patch(endpoint, body) { return this.request(endpoint, { method: 'PATCH', body: JSON.stringify(body) }); },
  delete(endpoint) { return this.request(endpoint, { method: 'DELETE' }); },
};