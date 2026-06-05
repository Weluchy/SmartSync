import { useState, useEffect, useCallback, lazy, Suspense } from 'react';
import { api } from './api/client';
import Sidebar from './components/Sidebar/Sidebar';
import { Toaster, toast } from 'react-hot-toast';

// Lazy-loaded components — грузятся только когда нужны
const MainLayout = lazy(() => import('./components/Layout/MainLayout'));
const KanbanBoard = lazy(() => import('./components/Kanban/KanbanBoard'));
const TaskGraph = lazy(() => import('./components/Graph/TaskGraph'));
const UserProfile = lazy(() => import('./components/Profile/UserProfile'));
const Dashboard = lazy(() => import('./components/Dashboard/Dashboard'));
const UserProfilePage = lazy(() => import('./components/Profile/UserProfilePage'));

// Плейсхолдер загрузки
const PageLoader = () => (
  <div className="flex-1 flex items-center justify-center bg-gray-50">
    <div className="text-gray-400 text-sm font-medium animate-pulse">Загрузка...</div>
  </div>
);

export default function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(!!localStorage.getItem('token'));
  const [projects, setProjects] = useState([]);
  const [currentProjectId, setCurrentProjectId] = useState(null);
  const [activeView, setActiveView] = useState('graph');
  const [invitations, setInvitations] = useState([]);
  const [viewUserId, setViewUserId] = useState(null);
  
  const [authMode, setAuthMode] = useState('login');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

const loadInvitations = useCallback(async () => {
    try {
      const data = await api.get('/invitations/my');
      setInvitations(data || []);
    } catch (err) { console.error(err); }
  }, []);

  const loadProjects = useCallback(async () => {
    try {
      const data = await api.get('/projects');
      setProjects(data || []);
      if (data?.length > 0 && !currentProjectId) {
        setCurrentProjectId(data[0].id);
      }
    } catch (err) { console.error('Ошибка загрузки проектов:', err); }
  }, [currentProjectId]);

  useEffect(() => {
    if (isAuthenticated) {
      loadProjects();
      loadInvitations();
      
      const interval = setInterval(() => {
        loadProjects();
        loadInvitations();
      }, 10000); 

      return () => clearInterval(interval);
    }
  }, [isAuthenticated, loadProjects, loadInvitations]);

  const handleAuthSubmit = async (e) => {
    e.preventDefault();
    setError('');
    try {
      const endpoint = authMode === 'register' ? '/register' : '/login';
      const data = await api.post(endpoint, { username, password });
      
      if (authMode === 'register') {
        toast.success('Аккаунт создан! Теперь войдите под своим логином.', {
          style: { background: '#1a1a2e', color: '#7ac9a7', border: '1px solid #7ac9a7' }
        });
        setAuthMode('login');
        setPassword('');
      } else {
        localStorage.setItem('token', data.token);
        toast.success('Успешный вход!', {
          style: { background: '#1a1a2e', color: '#7ac9a7', border: '1px solid #7ac9a7' }
        });
        setIsAuthenticated(true);
      }
    } catch (err) { 
      setError(err.message); 
      toast.error(err.message, {
        style: { background: '#1a1a2e', color: '#f87171', border: '1px solid #f87171' }
      });
    }
  };

  const logout = () => {
    localStorage.removeItem('token');
    // Очищаем все сохранённые позиции графа при выходе
    const keysToRemove = [];
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key && key.startsWith('smartsync_graph_positions_')) {
        keysToRemove.push(key);
      }
    }
    keysToRemove.forEach(k => localStorage.removeItem(k));
    toast('Вы вышли из системы', {
      icon: '👋',
      style: { background: '#1a1a2e', color: '#e4e4e7', border: '1px solid #6366f1' }
    });
    setIsAuthenticated(false);
  };

  const currentProject = projects.find(p => p.id === currentProjectId);

  if (isAuthenticated) {
    return (
      <div className="flex h-screen w-full">
        <Toaster
          position="top-right"
          toastOptions={{
            duration: 4000,
            style: {
              borderRadius: '12px',
              padding: '12px 16px',
              fontSize: '13px',
              fontWeight: 500,
            },
          }}
        />
        <Sidebar 
  projects={projects} 
  currentProjectId={currentProjectId} 
  onSelectProject={setCurrentProjectId}
  onCreateProject={async (name) => {
    await api.post('/projects', { name });
    loadProjects();
  }}
  onProjectUpdated={loadProjects}
/>
        <Suspense fallback={<PageLoader />}>
          <MainLayout 
  projectName={currentProject?.name}
  activeView={activeView}
  onSwitchView={(view) => {
    setActiveView(view);
    setViewUserId(null);
  }}
  onLogout={logout}
  tasks={[]}
  invitations={invitations}
  onSelectProject={setCurrentProjectId}
>
   {viewUserId ? (
    <UserProfilePage projectId={currentProjectId} userId={viewUserId} onBack={() => setViewUserId(null)} />
  ) : activeView === 'kanban' ? (
    <KanbanBoard projectId={currentProjectId} onTasksChange={(t) => console.log('tasks loaded:', t?.length)} onViewUser={(uid) => setViewUserId(uid)} />
  ) : activeView === 'profile' ? (
    <UserProfile />
  ) : activeView === 'dashboard' ? (
    <Dashboard projectId={currentProjectId} />
  ) : (
    <TaskGraph projectId={currentProjectId} />
  )}
</MainLayout>
        </Suspense>
      </div>
    );
  }

  return (
    <div className="flex h-screen items-center justify-center bg-gray-900">
      <Toaster
        position="top-right"
        toastOptions={{
          duration: 4000,
          style: {
            borderRadius: '12px',
            padding: '12px 16px',
            fontSize: '13px',
            fontWeight: 500,
          },
        }}
      />
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

<div className="flex flex-col gap-4">
  {/* Поле логина */}
  <input 
    type="text" 
    placeholder="Логин" 
    value={username}
    onChange={e => setUsername(e.target.value)}
    className="border p-3 rounded-lg bg-gray-50 outline-none focus:ring-2 focus:ring-blue-500" 
    autoComplete="off"
    data-lpignore="true"
  />
  {/* Поле пароля (замаскировано под обычный текст через CSS, чтобы ослепить расширения) */}
  <input 
    type="text" 
    placeholder="Пароль" 
    value={password}
    onChange={e => setPassword(e.target.value)}
    className="border p-3 rounded-lg bg-gray-50 outline-none focus:ring-2 focus:ring-blue-500" 
    style={{ WebkitTextSecurity: 'disc' }}
    autoComplete="off"
    data-lpignore="true"
  />
  <button 
    type="button" 
    onClick={handleAuthSubmit} 
    className="bg-blue-600 text-white font-bold py-3 rounded-lg hover:bg-blue-700 transition shadow-lg mt-2"
  >
    {authMode === 'login' ? 'Войти' : 'Создать аккаунт'}
  </button>
</div>
      </div>
    </div>
  );
}