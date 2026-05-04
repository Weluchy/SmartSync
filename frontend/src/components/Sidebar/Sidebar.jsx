import { useState } from 'react';
import PropTypes from 'prop-types'; // Добавили импорт
import { Plus, Users, Send, Settings2 } from 'lucide-react';
import { api } from '../../api/client';

export default function Sidebar({ projects, currentProjectId, onSelectProject, onCreateProject }) {
  const [inviteUser, setInviteUser] = useState('');

  const handleInvite = async (e) => {
    e.preventDefault();
    if (!inviteUser) return;
    try {
      await api.post(`/projects/${currentProjectId}/members`, { username: inviteUser }); //[cite: 9]
      alert(`Пользователь ${inviteUser} успешно добавлен в проект!`);
      setInviteUser('');
    } catch (err) {
      alert('Ошибка: ' + err.message);
    }
  };

  return (
    <aside className="w-72 bg-white border-r border-gray-200 flex flex-col shadow-sm">
      <div className="p-6 border-b border-gray-100">
        <h2 className="text-xl font-black text-blue-600 tracking-tight italic">SmartSync.engine</h2>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-1">
        <div className="flex items-center justify-between mb-4 px-2">
          <span className="text-xs font-bold text-gray-400 uppercase tracking-widest">Проекты</span>
          <button 
            onClick={() => {
              const name = prompt('Название нового проекта:');
              if (name) onCreateProject(name); //[cite: 14]
            }} 
            className="text-blue-600 hover:bg-blue-50 p-1 rounded transition-colors"
          >
            <Plus size={16} />
          </button>
        </div>
        
        {projects.map(p => (
          <div 
            key={p.id}
            onClick={() => onSelectProject(p.id)}
            className={`group flex items-center justify-between p-2.5 rounded-lg cursor-pointer transition-all ${
              currentProjectId === p.id ? 'bg-blue-50 text-blue-700 shadow-sm font-bold' : 'text-gray-600 hover:bg-gray-50'
            }`}
          >
            <span className="truncate text-sm font-medium">{p.name}</span>
            <Settings2 size={14} className="opacity-0 group-hover:opacity-100 text-gray-400" />
          </div>
        ))}
      </div>

      {currentProjectId && (
        <div className="p-4 border-t border-gray-100 bg-gray-50/50">
          <div className="flex items-center gap-2 mb-3 px-2 text-gray-500">
            <Users size={14} />
            <span className="text-[10px] font-bold uppercase tracking-widest">Команда проекта</span>
          </div>
          <form onSubmit={handleInvite} className="relative">
            <input 
              type="text" 
              placeholder="Логин коллеги..." 
              value={inviteUser}
              onChange={e => setInviteUser(e.target.value)}
              className="w-full text-xs border rounded-lg pl-3 pr-8 py-2 outline-none focus:ring-2 focus:ring-blue-500 bg-white" 
            />
            <button type="submit" className="absolute right-2 top-1.5 text-blue-600 hover:text-blue-800">
              <Send size={14} />
            </button>
          </form>
        </div>
      )}
    </aside>
  );
}

// РЕШЕНИЕ: Добавляем валидацию всех пропсов
Sidebar.propTypes = {
  projects: PropTypes.array.isRequired,
  currentProjectId: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
  onSelectProject: PropTypes.func.isRequired,
  onCreateProject: PropTypes.func.isRequired,
};