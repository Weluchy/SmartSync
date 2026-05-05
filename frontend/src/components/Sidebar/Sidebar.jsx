import { useState, useEffect, useCallback } from 'react'; // ИСПРАВЛЕНО: добавлены хуки
import PropTypes from 'prop-types';
import { Plus, Users, Send, Settings2 } from 'lucide-react'; 
import { api } from '../../api/client';

export default function Sidebar({ projects, currentProjectId, onSelectProject, onCreateProject, invitations = [] }) {
  const [inviteUser, setInviteUser] = useState('');
  const [members, setMembers] = useState([]);

  // Загрузка участников текущего проекта
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
  if (!confirm('Удалить участника из проекта?')) return;
  try {
    await api.delete(`/projects/${currentProjectId}/members/${userId}`);
    loadMembers(); // Обновляем список
  } catch (err) { alert('Ошибка при удалении: ' + err.message); }
};

 const handleInvite = async (e) => {
    e.preventDefault();
    if (!inviteUser) return;
    try {
      await api.post(`/projects/${currentProjectId}/members`, { username: inviteUser });
      alert(`Пользователь ${inviteUser} успешно добавлен!`);
      setInviteUser('');
      loadMembers(); // Обновляем список участников сразу после добавления
    } catch (err) {
      alert('Ошибка: ' + err.message);
    }
  };


  useEffect(() => {
    loadMembers();
  }, [loadMembers]);

 

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
              if (name) onCreateProject(name);
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

      {/* Входящие приглашения (уже было добавлено) */}
      {invitations.length > 0 && (
        <div className="p-4 border-t border-gray-100 bg-blue-50/30">
          <span className="text-[10px] font-bold text-gray-400 uppercase tracking-widest block mb-3 px-2">Приглашения</span>
          <div className="space-y-2">
            {invitations.map(inv => (
              <div key={inv.id} className="bg-white p-3 rounded-xl border border-blue-100 shadow-sm">
                <p className="text-[11px] text-blue-800 font-medium mb-2">Проект: {inv.project_name}</p>
                <button className="w-full bg-blue-600 text-white text-[10px] font-bold py-1.5 rounded-lg hover:bg-blue-700">
                  ПРИНЯТЬ
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* НОВОЕ: СПИСОК УЧАСТНИКОВ И ФОРМА ПРИГЛАШЕНИЯ */}
      {currentProjectId && (
        <div className="p-4 border-t border-gray-100 bg-gray-50/50">
          <div className="flex items-center gap-2 mb-3 px-2 text-gray-500">
            <Users size={14} />
            <span className="text-[10px] font-bold uppercase tracking-widest">Участники</span>
          </div>

          {/* Отрисовка списка людей из стейта members */}
          
          <div className="space-y-2 mb-4 max-h-40 overflow-y-auto px-1">
            {members.map(member => (
              <div key={member.user_id} className="flex items-center justify-between text-xs py-2 border-b border-gray-50 last:border-0 group">
                <div className="flex flex-col">
                  <span className="text-gray-800 font-semibold">{member.username}</span>
                  <span className={`text-[9px] uppercase font-bold ${member.role === 'owner' ? 'text-blue-500' : 'text-gray-400'}`}>
                    {member.role}
                  </span>
                </div>
                
                {/* Кнопка удаления видна только если это не владелец */}
                {member.role !== 'owner' && (
                  <button 
                    onClick={() => removeMember(member.user_id)}
                    className="opacity-0 group-hover:opacity-100 text-red-400 hover:text-red-600 transition-all p-1"
                    title="Удалить из проекта"
                  >
                    <Plus size={14} className="rotate-45" /> 
                  </button>
                )}
              </div>
            ))}
          </div>


          <form onSubmit={handleInvite} className="relative">
            <input 
              type="text" 
              placeholder="Логин коллеги..." 
              value={inviteUser}
              onChange={e => setInviteUser(e.target.value)}
              className="w-full text-xs border border-gray-200 rounded-lg pl-3 pr-8 py-2 outline-none focus:ring-2 focus:ring-blue-500 bg-white shadow-inner" 
            />
            <button type="submit" className="absolute right-2 top-1.5 text-blue-600 hover:text-blue-800 transition-colors">
              <Send size={14} />
            </button>
          </form>
        </div>
      )}
    </aside>
  );
}

Sidebar.propTypes = {
  projects: PropTypes.array.isRequired,
  currentProjectId: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
  onSelectProject: PropTypes.func.isRequired,
  onCreateProject: PropTypes.func.isRequired,
  invitations: PropTypes.array
};