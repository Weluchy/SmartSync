import { useState, useEffect } from 'react';
import { api } from './api/client';
<<<<<<< HEAD
import MainLayout from './components/Layout/MainLayout';
import Sidebar from './components/Sidebar/Sidebar';

export default function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(!!localStorage.getItem('token'));
  const [projects, setProjects] = useState([]);
  const [currentProjectId, setCurrentProjectId] = useState(null);
  const [activeView, setActiveView] = useState('graph');
  
  // Состояние формы входа (оставим как в прошлый раз)
  const [authMode, setAuthMode] = useState('login');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');

  useEffect(() => {
    if (isAuthenticated) {
      loadProjects();
    }
  }, [isAuthenticated]);

  async function loadProjects() {
    try {
      const data = await api.get('/projects');
      setProjects(data || []);
      if (data?.length > 0 && !currentProjectId) {
        setCurrentProjectId(data[0].id);
      }
    } catch (err) { console.error(err); }
  }

  async function handleCreateProject(name) {
    await api.post('/projects', { name });
    loadProjects();
  }
=======

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
>>>>>>> 52a78df0226ee1eef55feb12b6822cf095a3b258

  const logout = () => {
    localStorage.removeItem('token');
    setIsAuthenticated(false);
  };

<<<<<<< HEAD
  const currentProject = projects.find(p => p.id === currentProjectId);

  if (isAuthenticated) {
    return (
      <div className="flex h-screen w-full">
        <Sidebar 
          projects={projects} 
          currentProjectId={currentProjectId} 
          onSelectProject={setCurrentProjectId}
          onCreateProject={handleCreateProject}
        />
        <MainLayout 
          projectName={currentProject?.name}
          activeView={activeView}
          onSwitchView={setActiveView}
          onLogout={logout}
        >
          {/* Сюда мы вставим сам Граф или Канбан */}
          <div className="p-8">
            <div className="bg-white border-2 border-dashed border-gray-200 rounded-2xl h-[calc(100vh-120px)] flex items-center justify-center text-gray-400">
              Здесь будет {activeView === 'graph' ? 'Граф' : 'Канбан-доска'}
            </div>
          </div>
        </MainLayout>
=======
  // Если авторизован - показываем заглушку будущего приложения
  if (isAuthenticated) {
    return (
      <div className="flex h-screen items-center justify-center bg-gray-50 flex-col gap-4">
        <h1 className="text-3xl font-black text-green-600">Мы в системе! 🎉</h1>
        <p className="text-gray-500">React + Vite инфраструктура настроена.</p>
        <button onClick={logout} className="px-4 py-2 bg-red-100 text-red-600 font-bold rounded-lg hover:bg-red-200">Выйти</button>
>>>>>>> 52a78df0226ee1eef55feb12b6822cf095a3b258
      </div>
    );
  }

<<<<<<< HEAD
=======
  // Если нет - показываем красивую форму
>>>>>>> 52a78df0226ee1eef55feb12b6822cf095a3b258
  return (
    <div className="flex h-screen items-center justify-center bg-gray-900">
      <div className="bg-white p-8 rounded-xl shadow-2xl w-96">
        <h1 className="text-3xl font-black text-center text-blue-600 mb-6">SmartSync</h1>
<<<<<<< HEAD
        <div className="flex gap-2 mb-6 border-b border-gray-200">
          <button onClick={() => setAuthMode('login')} className={`flex-1 py-2 font-bold ${authMode === 'login' ? 'border-b-2 border-blue-600 text-blue-600' : 'text-gray-400'}`}>Вход</button>
          <button onClick={() => setAuthMode('register')} className={`flex-1 py-2 font-bold ${authMode === 'register' ? 'border-b-2 border-blue-600 text-blue-600' : 'text-gray-400'}`}>Регистрация</button>
        </div>
        {error && <div className="mb-4 p-3 bg-red-50 text-red-600 text-sm rounded-lg">{error}</div>}
        <form onSubmit={handleAuthSubmit} className="flex flex-col gap-4">
          <input type="text" placeholder="Логин" value={username} onChange={e => setUsername(e.target.value)} className="border p-3 rounded-lg bg-gray-50 outline-none focus:ring-2 focus:ring-blue-500" required />
          <input type="password" placeholder="Пароль" value={password} onChange={e => setPassword(e.target.value)} className="border p-3 rounded-lg bg-gray-50 outline-none focus:ring-2 focus:ring-blue-500" required />
          <button type="submit" className="bg-blue-600 text-white font-bold py-3 rounded-lg hover:bg-blue-700 transition">
=======
        
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
>>>>>>> 52a78df0226ee1eef55feb12b6822cf095a3b258
            {authMode === 'login' ? 'Войти' : 'Создать аккаунт'}
          </button>
        </form>
      </div>
    </div>
  );
}