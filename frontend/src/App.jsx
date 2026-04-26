import { useState, useEffect } from 'react';
import { api } from './api/client';
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
      </div>
    );
  }

  return (
    <div className="flex h-screen items-center justify-center bg-gray-900">
      <div className="bg-white p-8 rounded-xl shadow-2xl w-96">
        <h1 className="text-3xl font-black text-center text-blue-600 mb-6">SmartSync</h1>
        <div className="flex gap-2 mb-6 border-b border-gray-200">
          <button onClick={() => setAuthMode('login')} className={`flex-1 py-2 font-bold ${authMode === 'login' ? 'border-b-2 border-blue-600 text-blue-600' : 'text-gray-400'}`}>Вход</button>
          <button onClick={() => setAuthMode('register')} className={`flex-1 py-2 font-bold ${authMode === 'register' ? 'border-b-2 border-blue-600 text-blue-600' : 'text-gray-400'}`}>Регистрация</button>
        </div>
        {error && <div className="mb-4 p-3 bg-red-50 text-red-600 text-sm rounded-lg">{error}</div>}
        <form onSubmit={handleAuthSubmit} className="flex flex-col gap-4">
          <input type="text" placeholder="Логин" value={username} onChange={e => setUsername(e.target.value)} className="border p-3 rounded-lg bg-gray-50 outline-none focus:ring-2 focus:ring-blue-500" required />
          <input type="password" placeholder="Пароль" value={password} onChange={e => setPassword(e.target.value)} className="border p-3 rounded-lg bg-gray-50 outline-none focus:ring-2 focus:ring-blue-500" required />
          <button type="submit" className="bg-blue-600 text-white font-bold py-3 rounded-lg hover:bg-blue-700 transition">
            {authMode === 'login' ? 'Войти' : 'Создать аккаунт'}
          </button>
        </form>
      </div>
    </div>
  );
}