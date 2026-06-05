import { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { X, Users, Edit3, AlertTriangle, Send } from 'lucide-react';
import { api } from '../../api/client';
import { toast } from 'react-hot-toast';

export default function ProjectSettingsModal({ isOpen, onClose, project, onProjectUpdated }) {
  const [activeTab, setActiveTab] = useState('rename');
  const [newName, setNewName] = useState('');
  const [members, setMembers] = useState([]);
  const [inviteUser, setInviteUser] = useState('');
  const [selectedRole, setSelectedRole] = useState('viewer');
  const [deleteConfirm, setDeleteConfirm] = useState('');

  useEffect(() => {
    // ИСПРАВЛЕНИЕ: Теперь мы берем данные напрямую из объекта project
    if (isOpen && project) {
      setNewName(project.name || '');
      setDeleteConfirm('');
      api.get(`/projects/${project.id}/members`).then(data => setMembers(data || [])).catch(() => {});
    }
  }, [isOpen, project]);

  if (!isOpen || !project) return null;

  const handleRename = async () => {
    if (!newName.trim()) return;
    try {
      await api.put(`/projects/${project.id}`, { name: newName.trim() });
      toast.success('Проект переименован', {
        style: { background: '#1a1a2e', color: '#7ac9a7', border: '1px solid #7ac9a7' }
      });
      onProjectUpdated?.();
      window.location.reload(); // Чтобы обновилось в боковой панели
    } catch (err) {
      toast.error(err.message, {
        style: { background: '#1a1a2e', color: '#f87171', border: '1px solid #f87171' }
      });
    }
  };

  const handleInvite = async (e) => {
    e.preventDefault();
    try {
      await api.post(`/projects/${project.id}/members`, { 
        username: inviteUser, 
        role: selectedRole
      });
      setInviteUser('');
      toast.success(`Приглашение отправлено ${inviteUser}`, {
        style: { background: '#1a1a2e', color: '#7ac9a7', border: '1px solid #7ac9a7' }
      });
      const data = await api.get(`/projects/${project.id}/members`);
      setMembers(data || []);
    } catch (err) { 
      toast.error(err.message, {
        style: { background: '#1a1a2e', color: '#f87171', border: '1px solid #f87171' }
      });
    }
  };

  const removeMember = async (userId) => {
    try {
      await api.delete(`/projects/${project.id}/members/${userId}`);
      toast('Участник удалён', {
        icon: '👤',
        style: { background: '#1a1a2e', color: '#e4e4e7', border: '1px solid #f87171' }
      });
      const data = await api.get(`/projects/${project.id}/members`);
      setMembers(data || []);
    } catch (err) { 
      toast.error(err.message, {
        style: { background: '#1a1a2e', color: '#f87171', border: '1px solid #f87171' }
      });
    }
  };

  const changeRole = async (userId, newRole) => {
    try {
      await api.patch(`/projects/${project.id}/members/${userId}`, { role: newRole });
      const data = await api.get(`/projects/${project.id}/members`);
      setMembers(data || []);
    } catch (err) { 
      toast.error(err.message, {
        style: { background: '#1a1a2e', color: '#f87171', border: '1px solid #f87171' }
      });
    }
  };

  const handleDeleteProject = async () => {
    if (deleteConfirm !== project.name) return; // Защита от случайного удаления
    try {
      await api.delete(`/projects/${project.id}`);
      toast('Проект удалён', {
        icon: '🗑️',
        style: { background: '#1a1a2e', color: '#e4e4e7', border: '1px solid #f87171' }
      });
      onClose();
      onProjectUpdated?.();
      window.location.reload();
    } catch (err) {
      toast.error(err.message, {
        style: { background: '#1a1a2e', color: '#f87171', border: '1px solid #f87171' }
      });
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-2xl shadow-2xl w-full max-w-lg max-h-[80vh] flex flex-col">
        {/* Header */}
        <div className="flex justify-between items-center p-6 border-b shrink-0">
          <h2 className="text-lg font-bold text-gray-800">Настройки проекта</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 transition-colors"><X size={24} /></button>
        </div>

        {/* Tabs */}
        <div className="flex border-b px-6 gap-6 shrink-0 text-sm font-bold text-gray-500">
          <button onClick={() => setActiveTab('rename')} className={`py-3 border-b-2 transition-colors flex items-center gap-2 ${activeTab === 'rename' ? 'border-blue-600 text-blue-600' : 'border-transparent hover:text-gray-800'}`}>
            <Edit3 size={14} /> Переименовать
          </button>
          <button onClick={() => setActiveTab('members')} className={`py-3 border-b-2 transition-colors flex items-center gap-2 ${activeTab === 'members' ? 'border-blue-600 text-blue-600' : 'border-transparent hover:text-gray-800'}`}>
            <Users size={14} /> Участники
          </button>
          <button onClick={() => setActiveTab('danger')} className={`py-3 border-b-2 transition-colors flex items-center gap-2 ${activeTab === 'danger' ? 'border-red-600 text-red-600' : 'border-transparent hover:text-gray-800'}`}>
            <AlertTriangle size={14} /> Опасная зона
          </button>
        </div>

        {/* Body */}
        <div className="flex-1 overflow-y-auto p-6">
          {activeTab === 'rename' && (
            <div className="space-y-4">
              <div>
                <label className="block text-xs font-bold text-gray-700 uppercase mb-1">Новое название проекта</label>
                <input 
                  value={newName} 
                  onChange={e => setNewName(e.target.value)}
                  className="w-full border rounded-lg p-2.5 bg-white outline-none focus:ring-2 focus:ring-blue-500 shadow-sm"
                  placeholder="Введите новое название..."
                />
              </div>
              <button 
                onClick={handleRename}
                className="w-full bg-blue-600 text-white font-bold py-2.5 rounded-xl hover:bg-blue-700 transition-colors"
              >
                Сохранить
              </button>
            </div>
          )}

          {activeTab === 'members' && (
            <div className="space-y-4">
              <div className="space-y-2 max-h-48 overflow-y-auto">
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
                      <button onClick={() => removeMember(member.user_id)}
                        className="opacity-0 group-hover:opacity-100 text-red-400 hover:text-red-600 transition-all p-1"
                        title="Удалить из проекта"
                      >
                        <X size={14} />
                      </button>
                    )}
                  </div>
                ))}
              </div>

              <form onSubmit={handleInvite} className="space-y-2 pt-3 border-t">
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
                    <option value="viewer">Viewer</option>
                    <option value="editor">Editor</option>
                    <option value="admin">Admin</option>
                  </select>
                  <button type="submit" className="bg-blue-600 text-white p-1.5 rounded-lg">
                    <Send size={14} />
                  </button>
                </div>
              </form>
            </div>
          )}

          {activeTab === 'danger' && (
            <div className="space-y-4">
              <div className="bg-red-50 border border-red-200 rounded-xl p-4">
                <h4 className="text-sm font-bold text-red-700 flex items-center gap-2">
                  <AlertTriangle size={16} /> Удаление проекта
                </h4>
                <p className="text-xs text-red-600 mt-2 mb-4">
                  Это действие удалит все задачи, комментарии и данные проекта без возможности восстановления.
                </p>
                <label className="block text-[10px] font-bold text-red-700 uppercase mb-1">
                  Введите <b>{project.name}</b> для подтверждения:
                </label>
                <input
                  value={deleteConfirm}
                  onChange={e => setDeleteConfirm(e.target.value)}
                  placeholder={project.name}
                  className="w-full border border-red-300 rounded-lg p-2 outline-none focus:ring-2 focus:ring-red-500 text-sm bg-white mb-4"
                />
                <button 
                  onClick={handleDeleteProject}
                  disabled={deleteConfirm !== project.name}
                  className={`w-full font-bold py-2.5 rounded-xl transition-colors ${deleteConfirm === project.name ? 'bg-red-600 text-white hover:bg-red-700' : 'bg-red-200 text-red-400 cursor-not-allowed'}`}
                >
                  Удалить проект навсегда
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

ProjectSettingsModal.propTypes = {
  isOpen: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  project: PropTypes.object,
  onProjectUpdated: PropTypes.func
};