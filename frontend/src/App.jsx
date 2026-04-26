import { useState, useEffect } from 'react';
import { api } from './api/client';

export default function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(!!localStorage.getItem('token'));
  const [authMode, setAuthMode] = useState('login'); // 'login' или 'register'
  
  // Форма
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    // Слушаем событие протухания токена из нашего API клиента
    const handleLogout = () => setIsAuthenticated(false);
    window.addEventListener('auth-expired', handleLogout);
    return () => window.removeEventListener('auth-expired', handleLogout);
  }, []);

  const handleAuthSubmit = async (e) => {
    e.preventDefault();
    setError('');
    try {
      if (authMode === 'register') {
        await api.post('/register', { username, password });
        alert('Регистрация успешна! Теперь войдите.');
        setAuthMode('login');
        setPassword('');
      } else {
        const data = await api.post('/login', { username, password });
        localStorage.setItem('token', data.token);
        setIsAuthenticated(true);
      }
    } catch (err) {
      setError(err.message);
    }
  };

  const logout = () => {
    localStorage.removeItem('token');
    setIsAuthenticated(false);
  };

  // Если авторизован - показываем заглушку будущего приложения
  if (isAuthenticated) {
    return (
      <div className="flex h-screen items-center justify-center bg-gray-50 flex-col gap-4">
        <h1 className="text-3xl font-black text-green-600">Мы в системе! 🎉</h1>
        <p className="text-gray-500">React + Vite инфраструктура настроена.</p>
        <button onClick={logout} className="px-4 py-2 bg-red-100 text-red-600 font-bold rounded-lg hover:bg-red-200">Выйти</button>
      </div>
    );
  }

  // Если нет - показываем красивую форму
  return (
    <div className="flex h-screen items-center justify-center bg-gray-900">
      <div className="bg-white p-8 rounded-xl shadow-2xl w-96">
        <h1 className="text-3xl font-black text-center text-blue-600 mb-6">SmartSync</h1>
        
        <div className="flex gap-2 mb-6 border-b border-gray-200">
          <button 
            onClick={() => { setAuthMode('login'); setError(''); }} 
            className={`flex-1 py-2 font-bold transition-colors ${authMode === 'login' ? 'border-b-2 border-blue-600 text-blue-600' : 'text-gray-400 hover:text-gray-600'}`}
          >
            Вход
          </button>
          <button 
            onClick={() => { setAuthMode('register'); setError(''); }} 
            className={`flex-1 py-2 font-bold transition-colors ${authMode === 'register' ? 'border-b-2 border-blue-600 text-blue-600' : 'text-gray-400 hover:text-gray-600'}`}
          >
            Регистрация
          </button>
        </div>

        {error && <div className="mb-4 p-3 bg-red-50 text-red-600 text-sm rounded-lg font-medium">{error}</div>}

        <form onSubmit={handleAuthSubmit} className="flex flex-col gap-4">
          <input 
            type="text" 
            placeholder="Логин" 
            value={username}
            onChange={e => setUsername(e.target.value)}
            className="border p-3 rounded-lg bg-gray-50 focus:ring-2 focus:ring-blue-500 outline-none" 
            required 
          />
          <input 
            type="password" 
            placeholder="Пароль" 
            value={password}
            onChange={e => setPassword(e.target.value)}
            className="border p-3 rounded-lg bg-gray-50 focus:ring-2 focus:ring-blue-500 outline-none" 
            required 
          />
          <button type="submit" className="bg-blue-600 text-white font-bold py-3 rounded-lg hover:bg-blue-700 transition shadow-lg mt-2">
            {authMode === 'login' ? 'Войти' : 'Создать аккаунт'}
          </button>
        </form>
      </div>
    </div>
  );
}