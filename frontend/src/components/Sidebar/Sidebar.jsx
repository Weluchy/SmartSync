import { useState, useEffect, useCallback } from 'react';
import PropTypes from 'prop-types';
import { Plus, Users, Send, Settings2, Moon, Sun, Activity, X } from 'lucide-react';
import { api } from '../../api/client';
import { useTheme } from '../../ThemeContext';
import { toast } from 'react-hot-toast';
import ProjectSettingsModal from './ProjectSettingsModal';

const MICROSERVICES = [
  { id: 'task', label: 'Task API' },
  { id: 'auth', label: 'Auth API' },
  { id: 'math', label: 'Math Engine' },
];

export default function Sidebar({ projects, currentProjectId, onSelectProject, onCreateProject, onProjectUpdated }) {
  const { isDark, toggleTheme } = useTheme();
  const [inviteUser, setInviteUser] = useState('');
  const [members, setMembers] = useState([]);
  const [selectedRole, setSelectedRole] = useState('viewer');
  const [servicesStatus, setServicesStatus] = useState({});
  const [editingProject, setEditingProject] = useState(null);

  // Пинг микросервисов
  const pingServices = useCallback(async () => {
    const results = {};
    for (const svc of MICROSERVICES) {
      try {
        const controller = new AbortController();
        const timeout = setTimeout(() => controller.abort(), 5000);
        const GATEWAY = import.meta.env.VITE_GATEWAY_URL || "/api";
        const token = localStorage.getItem('token');
        const headers = { 
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        };
        
        let res;
        if (svc.id === 'task') {
          res = await fetch(`${GATEWAY}/search?q=health_check_ping`, { 
            headers, signal: controller.signal 
          });
        } else if (svc.id === 'auth') {
          res = await fetch(`${GATEWAY}/user/profile`, { 
            headers, signal: controller.signal 
          });
        } else if (svc.id === 'math') {
          res = await fetch(`${GATEWAY}/projects/0/graph`, { 
            headers, signal: controller.signal 
          });
        }
        clearTimeout(timeout);
        results[svc.id] = res?.status !== 502 && res?.status !== 503 ? 'healthy' : 'down';
      } catch {
        results[svc.id] = 'down';
      }
    }
    setServicesStatus(results);
  }, []);

  useEffect(() => {
    pingServices();
    const interval = setInterval(pingServices, 30000);
    return () => clearInterval(interval);
  }, [pingServices]);

  const loadMembers = useCallback(async () => {
    if (!currentProjectId) return;
    try {
      const data = await api.get(`/projects/${currentProjectId}/members`);
      setMembers(data || []);
    } catch (err) {
      console.error('Ошибка загрузки участников:', err);
    }
  }, [currentProjectId]);

const removeMember = async (userId) => {
  try {
    await api.delete(`/projects/${currentProjectId}/members/${userId}`);
    toast('Участник удалён из проекта', {
      icon: '👤',
      style: { background: '#1a1a2e', color: '#e4e4e7', border: '1px solid #f87171' }
    });
    loadMembers();
  } catch (err) { 
    toast.error(err.message, {
      style: { background: '#1a1a2e', color: '#f87171', border: '1px solid #f87171' }
    });
  }
};

 const handleInvite = async (e) => {
    e.preventDefault();
    try {
      await api.post(`/projects/${currentProjectId}/members`, { 
        username: inviteUser, 
        role: selectedRole
      });
      setInviteUser('');
      toast.success(`Приглашение отправлено ${inviteUser}`, {
        style: { background: '#1a1a2e', color: '#7ac9a7', border: '1px solid #7ac9a7' }
      });
      loadMembers();
    } catch (err) { 
      toast.error(err.message, {
        style: { background: '#1a1a2e', color: '#f87171', border: '1px solid #f87171' }
      });
    }
};

const changeRole = async (userId, newRole) => {
    try {
        await api.patch(`/projects/${currentProjectId}/members/${userId}`, { role: newRole });
        loadMembers();
    } catch (err) { 
      toast.error(err.message, {
        style: { background: '#1a1a2e', color: '#f87171', border: '1px solid #f87171' }
      });
    }
};

  useEffect(() => {
    loadMembers();
  }, [loadMembers]);

  return (
    <aside className="w-72 flex flex-col shadow-sm relative" style={{ backgroundColor: 'var(--bg-sidebar)', borderRight: '1px solid var(--border)' }}>
      <div className="p-5 border-b flex items-center justify-between" style={{ borderColor: 'var(--border)' }}>
        <h2 className="text-lg font-black text-blue-600 tracking-tight italic">SmartSync.engine</h2>
        <button onClick={toggleTheme} className="p-1.5 rounded-lg transition-colors hover:bg-gray-100" style={{ color: 'var(--text-secondary)' }}>
          {isDark ? <Sun size={16} /> : <Moon size={16} />}
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-1">
        <div className="flex items-center justify-between mb-4 px-2">
          <span className="text-xs font-bold uppercase tracking-widest" style={{ color: 'var(--text-muted)' }}>Проекты</span>
          <button 
            onClick={() => {
              const name = prompt('Название нового проекта:');
              if (name) onCreateProject(name);
            }} 
            className="text-blue-600 hover:bg-blue-50 p-1 rounded transition-colors"
          >
            <Plus size={16} />
          </button>
        </div>
        
        {projects.map(p => {
          const isOwner = p.role === 'owner' || !p.role;
          return (
            <div 
              key={p.id}
              onClick={() => onSelectProject(p.id)}
              style={{ color: currentProjectId === p.id ? '' : 'var(--text-secondary)' }}
              className={`group flex items-center justify-between p-2.5 rounded-lg cursor-pointer transition-all ${
                currentProjectId === p.id ? 'bg-blue-50 text-blue-700 shadow-sm font-bold' : 'hover:opacity-80'
              }`}
            >
              <div className="flex items-center gap-2 truncate">
                {!isOwner && <Users size={12} className="text-indigo-400 shrink-0" title="Вы приглашены" />}
                <span className="truncate text-sm font-medium">{p.name}</span>
              </div>
              
              <Settings2 size={14} 
                className="opacity-0 group-hover:opacity-100 text-gray-400 hover:text-blue-500 shrink-0" 
                onClick={(e) => { e.stopPropagation(); setEditingProject(p); }} 
              />
            </div>
          );
        })}
      </div>

      {/* Участники и приглашение */}
      {currentProjectId && (
        <div className="p-4 border-t" style={{ borderColor: 'var(--border)', backgroundColor: 'var(--bg-card-hover)' }}>
          <div className="flex items-center gap-2 mb-3 px-2" style={{ color: 'var(--text-muted)' }}>
            <Users size={14} />
            <span className="text-[10px] font-bold uppercase tracking-widest">Участники</span>
          </div>

          <div className="space-y-2 mb-4 max-h-40 overflow-y-auto px-1">
            {members.map(member => (
  <div key={member.user_id} className="flex items-center justify-between text-xs py-2 border-b border-gray-50 last:border-0 group">
    <div className="flex flex-col flex-1">
      <span className="text-gray-800 font-semibold">{member.username}</span>
      
      {member.role === 'owner' ? (
        <span className="text-[9px] uppercase font-bold text-blue-500">{member.role}</span>
      ) : (
        <select 
          value={member.role}
          onChange={(e) => changeRole(member.user_id, e.target.value)}
          className="text-[9px] bg-transparent font-bold text-gray-400 uppercase outline-none cursor-pointer hover:text-blue-600"
        >
          <option value="viewer">viewer</option>
          <option value="editor">editor</option>
          <option value="admin">admin</option>
        </select>
      )}
    </div>
    
    {member.role !== 'owner' && (
      <button 
        onClick={() => removeMember(member.user_id)}
        className="opacity-0 group-hover:opacity-100 text-red-400 hover:text-red-600 transition-all p-1"
        title="Удалить из проекта"
      >
        <X size={14} />
      </button>
    )}
  </div>
))}
          </div>

          <form onSubmit={handleInvite} className="space-y-2">
  <input 
    value={inviteUser}
    onChange={e => setInviteUser(e.target.value)}
    placeholder="Логин коллеги..."
    className="w-full text-xs border rounded-lg p-2"
  />
  <div className="flex gap-2">
    <select 
      value={selectedRole}
      onChange={e => setSelectedRole(e.target.value)}
      className="flex-1 text-[10px] border rounded-lg bg-white p-1"
    >
      <option value="viewer">Viewer (10)</option>
      <option value="editor">Editor (40)</option>
      <option value="admin">Admin (80)</option>
    </select>
    <button type="submit" className="bg-blue-600 text-white p-1.5 rounded-lg">
      <Send size={14} />
    </button>
  </div>
</form>
        </div>
      )}

      {/* Индикатор здоровья микросервисов */}
      <div className="p-3 border-t flex items-center gap-3 px-4" style={{ borderColor: 'var(--border)', backgroundColor: 'var(--bg-card-hover)' }}>
        <Activity size={12} style={{ color: 'var(--text-muted)' }} />
        {MICROSERVICES.map(svc => (
          <div key={svc.id} className="flex items-center gap-1.5" title={`${svc.label}: ${servicesStatus[svc.id] === 'healthy' ? 'Работает' : 'Недоступен'}`}>
            <span className={`w-2 h-2 rounded-full ${
              servicesStatus[svc.id] === 'healthy' ? 'bg-green-500 shadow-sm shadow-green-400' : 'bg-red-500 shadow-sm shadow-red-400'
            }`} />
            <span className="text-[8px] font-bold uppercase tracking-wider" style={{ color: 'var(--text-muted)' }}>
              {svc.label === 'Task API' ? 'T' : svc.label === 'Auth API' ? 'A' : 'M'}
            </span>
          </div>
        ))}
      </div>

      <ProjectSettingsModal 
        isOpen={!!editingProject} 
        onClose={() => setEditingProject(null)} 
        project={editingProject} 
        onProjectUpdated={onProjectUpdated}
      />
    </aside>
  );
}

Sidebar.propTypes = {
  projects: PropTypes.array.isRequired,
  currentProjectId: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
  onSelectProject: PropTypes.func.isRequired,
  onCreateProject: PropTypes.func.isRequired,
  invitations: PropTypes.array,
  onProjectUpdated: PropTypes.func
};