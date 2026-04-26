import { useState, useEffect } from 'react';
import { api } from './api/client';
import MainLayout from './components/Layout/MainLayout';
import Sidebar from './components/Sidebar/Sidebar';
import KanbanBoard from './components/Kanban/KanbanBoard';

export default function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(!!localStorage.getItem('token'));
  const [projects, setProjects] = useState([]);
  const [currentProjectId, setCurrentProjectId] = useState(null);
  const [activeView, setActiveView] = useState('graph');
  
  const [authMode, setAuthMode] = useState('login');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    const handleLogout = () => setIsAuthenticated(false);
    window.addEventListener('auth-expired', handleLogout);
    if (isAuthenticated) loadProjects();
    return () => window.removeEventListener('auth-expired', handleLogout);
  }, [isAuthenticated]);

  async function loadProjects() {
    try {
      const data = await api.get('/projects');
      setProjects(data || []);
      if (data?.length > 0 && !currentProjectId) {
        setCurrentProjectId(data[0].id);
      }
    } catch (err) { 
      console.error('Ошибка загрузки проектов:', err); 
    }
  }

  const handleAuthSubmit = async (e) => {
    e.preventDefault();
    setError('');
    try {
      const endpoint = authMode === 'register' ? '/register' : '/login';
      const data = await api.post(endpoint, { username, password });
      
      if (authMode === 'register') {
        alert('Успешно! Теперь войдите под своим логином.');
        setAuthMode('login');
        setPassword('');
      } else {
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

  const currentProject = projects.find(p => p.id === currentProjectId);

  if (isAuthenticated) {
    return (
      <div className="flex h-screen w-full">
        <Sidebar 
          projects={projects} 
          currentProjectId={currentProjectId} 
          onSelectProject={setCurrentProjectId}
          onCreateProject={async (name) => {
            await api.post('/projects', { name });
            loadProjects();
          }}
        />
        <MainLayout 
  projectName={currentProject?.name}
  activeView={activeView}
  onSwitchView={setActiveView}
  onLogout={logout}
>
  {activeView === 'kanban' ? (
    <KanbanBoard projectId={currentProjectId} />
  ) : (
    <div className="p-8 h-full">
      <div className="bg-white border-2 border-dashed border-gray-200 rounded-2xl h-full flex items-center justify-center text-gray-400">
        Тут скоро будет твой математический Граф 🕸️
      </div>
    </div>
  )}
</MainLayout>
      </div>
    );
  }

  return (
    <div className="flex h-screen items-center justify-center bg-gray-900">
      <div className="bg-white p-8 rounded-xl shadow-2xl w-96">
        <h1 className="text-3xl font-black text-center text-blue-600 mb-6">SmartSync</h1>
        <div className="flex gap-2 mb-6 border-b border-gray-200">
          <button 
            onClick={() => setAuthMode('login')} 
            className={`flex-1 py-2 font-bold transition-colors ${authMode === 'login' ? 'border-b-2 border-blue-600 text-blue-600' : 'text-gray-400'}`}
          >
            Вход
          </button>
          <button 
            onClick={() => setAuthMode('register')} 
            className={`flex-1 py-2 font-bold transition-colors ${authMode === 'register' ? 'border-b-2 border-blue-600 text-blue-600' : 'text-gray-400'}`}
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
            className="border p-3 rounded-lg bg-gray-50 outline-none focus:ring-2 focus:ring-blue-500" 
            required 
          />
          <input 
            type="password" 
            placeholder="Пароль" 
            value={password}
            onChange={e => setPassword(e.target.value)}
            className="border p-3 rounded-lg bg-gray-50 outline-none focus:ring-2 focus:ring-blue-500" 
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