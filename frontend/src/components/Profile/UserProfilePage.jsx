import { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { api } from '../../api/client';
import { ArrowLeft, User, Clock, ListTodo } from 'lucide-react';

export default function UserProfilePage({ projectId, userId, onBack }) {
  const [user, setUser] = useState(null);
  const [tasks, setTasks] = useState([]);

  useEffect(() => {
    if (!userId) return;
    // Получаем данные пользователя (используем mock через bulk, т.к. GET /users/:id проксирован)
    api.get(`/users/${userId}`).then(setUser).catch(() => setUser({ username: `ID ${userId}` }));
    
    // Загружаем задачи проекта и фильтруем по пользователю
    if (projectId) {
      api.get(`/projects/${projectId}/tasks`).then(data => {
        setTasks((data || []).filter(t => 
          Number(t.user_id) === Number(userId) || Number(t.assignee_id) === Number(userId)
        ));
      }).catch(console.error);
    }
  }, [userId, projectId]);

  const userIdNum = Number(localStorage.getItem('userId'));
  const isMe = userIdNum === Number(userId);

  return (
    <div className="h-full w-full p-6 overflow-y-auto" style={{ backgroundColor: 'var(--bg-page)' }}>
      <div className="max-w-3xl mx-auto space-y-6">

        {/* Кнопка назад */}
        <button onClick={onBack} className="flex items-center gap-2 text-sm font-bold mb-2" style={{ color: 'var(--text-secondary)' }}>
          <ArrowLeft size={16} /> Назад
        </button>

        {/* Карточка профиля */}
        <div className="rounded-2xl p-8 border shadow-xl" style={{ backgroundColor: 'var(--bg-card)', borderColor: 'var(--border)' }}>
          <div className="flex items-center gap-6">
            <div className="w-20 h-20 bg-blue-100 text-blue-600 rounded-full flex items-center justify-center shadow-inner">
              <User size={40} />
            </div>
            <div>
              <h2 className="text-xl font-bold" style={{ color: 'var(--text-primary)' }}>
                {user?.username || 'Загрузка...'}
                {isMe && <span className="ml-2 text-xs bg-blue-100 text-blue-700 px-2 py-0.5 rounded-full">Это вы</span>}
              </h2>
              <p className="text-sm" style={{ color: 'var(--text-secondary)' }}>ID пользователя: {userId}</p>
            </div>
          </div>
        </div>

        {/* Статистика */}
        <div className="grid grid-cols-3 gap-4">
          <div className="rounded-xl p-5 border" style={{ backgroundColor: 'var(--bg-card)', borderColor: 'var(--border)' }}>
            <div className="flex items-center gap-2 mb-2">
              <ListTodo size={16} style={{ color: 'var(--text-muted)' }} />
              <span className="text-[10px] font-bold uppercase" style={{ color: 'var(--text-muted)' }}>Задачи</span>
            </div>
            <p className="text-2xl font-black" style={{ color: 'var(--text-primary)' }}>{tasks.length}</p>
          </div>
          <div className="rounded-xl p-5 border" style={{ backgroundColor: 'var(--bg-card)', borderColor: 'var(--border)' }}>
            <div className="flex items-center gap-2 mb-2">
              <Clock size={16} style={{ color: 'var(--text-muted)' }} />
              <span className="text-[10px] font-bold uppercase" style={{ color: 'var(--text-muted)' }}>В работе</span>
            </div>
            <p className="text-2xl font-black text-blue-500">{tasks.filter(t => t.status === 'in_progress').length}</p>
          </div>
          <div className="rounded-xl p-5 border" style={{ backgroundColor: 'var(--bg-card)', borderColor: 'var(--border)' }}>
            <div className="flex items-center gap-2 mb-2">
              <ListTodo size={16} style={{ color: 'var(--text-muted)' }} />
              <span className="text-[10px] font-bold uppercase" style={{ color: 'var(--text-muted)' }}>Готово</span>
            </div>
            <p className="text-2xl font-black text-green-500">{tasks.filter(t => t.status === 'done').length}</p>
          </div>
        </div>

        {/* Список задач пользователя */}
        {tasks.length > 0 && (
          <div className="rounded-xl border overflow-hidden" style={{ backgroundColor: 'var(--bg-card)', borderColor: 'var(--border)' }}>
            <div className="p-4 border-b" style={{ borderColor: 'var(--border)' }}>
              <h3 className="text-sm font-bold" style={{ color: 'var(--text-primary)' }}>Задачи пользователя</h3>
            </div>
            <div className="divide-y" style={{ borderColor: 'var(--border)' }}>
              {tasks.map(task => (
                <div key={task.id} className="p-4 flex justify-between items-center" style={{ backgroundColor: 'var(--bg-card-hover)' }}>
                  <div>
                    <p className="text-sm font-bold" style={{ color: 'var(--text-primary)' }}>#{task.id} {task.title}</p>
                    <p className="text-[10px]" style={{ color: 'var(--text-muted)' }}>{task.duration_hours?.toFixed(1) || '0'}ч</p>
                  </div>
                  <span className={`text-[10px] font-bold px-2 py-1 rounded ${
                    task.status === 'done' ? 'bg-green-100 text-green-700' : 
                    task.status === 'in_progress' ? 'bg-blue-100 text-blue-700' : 
                    'bg-gray-100 text-gray-600'
                  }`}>{task.status}</span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

UserProfilePage.propTypes = {
  projectId: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
  userId: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
  onBack: PropTypes.func
};
