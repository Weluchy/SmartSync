import { useState, useEffect, useCallback } from 'react';
import PropTypes from 'prop-types';
import { motion, AnimatePresence } from 'framer-motion';
import { api } from '../../api/client';
import { Plus, Trash2, Search, Filter, Clock } from 'lucide-react';
import TaskModal from './TaskModal';

const COLUMNS = [
  { id: 'todo', title: 'Бэклог', color: 'bg-gray-100' },
  { id: 'in_progress', title: 'В работе', color: 'bg-blue-50' },
  { id: 'done', title: 'Готово', color: 'bg-green-50' }
];

export default function KanbanBoard({ projectId, onTasksChange, onViewUser }) {
  const [tasks, setTasks] = useState([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingTask, setEditingTask] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [filterStatus, setFilterStatus] = useState('all');
  const [filterAssignee, setFilterAssignee] = useState('all');
  const [members, setMembers] = useState([]);

  const loadTasks = useCallback(async () => {
    if (!projectId) return;
    try {
      const data = await api.get(`/projects/${projectId}/tasks`);
      setTasks(data || []);
      if (onTasksChange) onTasksChange(data || []);
    } catch (err) { console.error(err); }
  }, [projectId, onTasksChange]);

  const loadMembers = useCallback(async () => {
    if (!projectId) return;
    try {
      const data = await api.get(`/projects/${projectId}/members`);
      setMembers(data || []);
    } catch (err) { console.error(err); }
  }, [projectId]);

  const handleSaveTask = async (taskData) => {
    try {
      if (editingTask) await api.put(`/tasks/${editingTask.id}`, taskData);
      else await api.post('/tasks', taskData);
      setIsModalOpen(false);
      setTimeout(loadTasks, 300);
    } catch (err) { alert(err.message); }
  };

  const deleteTask = async (e, id) => {
    e.stopPropagation();
    if (confirm('Удалить задачу?')) {
      try {
        await api.delete(`/tasks/${id}`);
        setTimeout(loadTasks, 300);
      } catch (err) { alert(err.message); }
    }
  };

  const updateTaskStatus = async (taskId, newStatus) => {
    try {
      await api.patch(`/tasks/${taskId}/status`, { status: newStatus });
      setTimeout(loadTasks, 300);
    } catch (err) { 
      alert(err.message); 
      loadTasks(); 
    }
  };

  useEffect(() => {
    loadTasks();
    loadMembers();
    const ws = new WebSocket('ws://localhost:8000/ws');
    ws.onopen = () => console.log('✅ WebSocket подключен');
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.project_id === Number(projectId)) loadTasks();
      } catch (e) { console.error(e); }
    };
    ws.onclose = () => console.log('❌ WebSocket отключен');
    return () => { if (ws.readyState === 1) ws.close(); };
  }, [loadTasks, loadMembers, projectId]);

  const filteredTasks = tasks.filter(t => {
    if (searchQuery && !t.title.toLowerCase().includes(searchQuery.toLowerCase())) return false;
    if (filterStatus !== 'all' && t.status !== filterStatus) return false;
    if (filterAssignee === 'unassigned' && t.assignee_id) return false;
    if (filterAssignee === 'assigned' && !t.assignee_id) return false;
    if (filterAssignee === 'me' && t.assignee_id !== Number(localStorage.getItem('userId'))) return false;
    if (filterAssignee !== 'all' && filterAssignee !== 'me' && filterAssignee !== 'assigned' && filterAssignee !== 'unassigned') {
      if (t.assignee_id !== Number(filterAssignee)) return false;
    }
    return true;
  });

  const maxScore = filteredTasks.length > 0 ? Math.max(...filteredTasks.map(t => t.priority_score || 0)) : 0;

  return (
    <div className="h-full w-full p-6 overflow-hidden" style={{ backgroundColor: 'var(--bg-page)' }}>
      <div className="w-full h-full flex flex-col rounded-2xl shadow-xl border overflow-hidden" style={{ backgroundColor: 'var(--bg-card)', borderColor: 'var(--border)' }}>
        <div className="h-[72px] min-h-[72px] px-8 border-b flex items-center gap-4" style={{ borderColor: 'var(--border)' }}>
          <div className="flex items-center gap-2 flex-1">
            <Search size={14} style={{ color: 'var(--text-muted)' }} />
            <input placeholder="Поиск задач..." value={searchQuery} onChange={e => setSearchQuery(e.target.value)}
              className="flex-1 bg-transparent text-sm outline-none" style={{ color: 'var(--text-primary)' }} />
          </div>
          <div className="flex items-center gap-2">
            <Filter size={14} style={{ color: 'var(--text-muted)' }} />
            <select value={filterStatus} onChange={e => setFilterStatus(e.target.value)}
              className="text-xs border rounded-lg p-1.5 outline-none" style={{ backgroundColor: 'var(--bg-input)', color: 'var(--text-secondary)', borderColor: 'var(--border)' }}>
              <option value="all">Все статусы</option>
              <option value="todo">Бэклог</option>
              <option value="in_progress">В работе</option>
              <option value="done">Готово</option>
            </select>
            <select value={filterAssignee} onChange={e => setFilterAssignee(e.target.value)}
              className="text-xs border rounded-lg p-1.5 outline-none" style={{ backgroundColor: 'var(--bg-input)', color: 'var(--text-secondary)', borderColor: 'var(--border)' }}>
              <option value="all">Все задачи</option>
              <option value="me">Мои задачи</option>
              <option value="assigned">Назначенные</option>
              <option value="unassigned">Не назначены</option>
              {members.map(m => <option key={m.user_id} value={m.user_id}>{m.username}</option>)}
            </select>
          </div>
          <button onClick={() => { setEditingTask(null); setIsModalOpen(true); }}
            className="bg-blue-600 text-white px-5 py-2 rounded-xl text-xs font-bold flex items-center gap-2 hover:bg-blue-700 shadow-md">
            <Plus size={16} /> СОЗДАТЬ
          </button>
        </div>

        <div className="flex-1 overflow-x-auto p-6">
          <div className="flex gap-6 h-full">
            {COLUMNS.map(col => (
              <div key={col.id} className="flex-shrink-0 w-80 flex flex-col h-full">
                <h3 className="font-bold text-xs uppercase tracking-widest mb-3 px-2" style={{ color: 'var(--text-secondary)' }}>
                  {col.title} · {filteredTasks.filter(t => t.status === col.id).length}
                </h3>
                <div onDragOver={e => e.preventDefault()} onDrop={e => updateTaskStatus(e.dataTransfer.getData('taskId'), col.id)}
                  className="flex-1 rounded-2xl p-3 space-y-3 border-2 border-dashed border-transparent overflow-y-auto"
                  style={{ backgroundColor: col.color === 'bg-gray-100' ? 'var(--kanban-bg)' : col.color === 'bg-blue-50' ? 'rgba(59,130,246,0.05)' : 'rgba(34,197,94,0.05)' }}>
                  <AnimatePresence>
                  {filteredTasks.filter(t => t.status === col.id).map(task => {
                    const isCritical = task.priority_score >= (maxScore * 0.8) && task.priority_score > 0 && task.status !== 'done';
                    // Дедлайн из PERT: created_at + duration_hours
                    let deadlineStr = '';
                    if (task.created_at && task.duration_hours > 0) {
                      const deadline = new Date(new Date(task.created_at).getTime() + task.duration_hours * 3600000);
                      const now = Date.now();
                      const diff = deadline.getTime() - now;
                      const hoursLeft = Math.floor(diff / 3600000);
                      const minsLeft = Math.floor((diff % 3600000) / 60000);
                      deadlineStr = diff < 0 ? 'Просрочено!' : `Осталось ${hoursLeft}ч ${minsLeft}м`;
                    }
                    return (
                      <motion.div key={task.id} layout initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, scale: 0.8 }} transition={{ duration: 0.2 }}
                        draggable onDragStart={e => { e.dataTransfer.setData('taskId', task.id); e.currentTarget.classList.add('dragging'); }}
                        onDragEnd={e => e.currentTarget.classList.remove('dragging')}
                        onClick={() => { setEditingTask(task); setIsModalOpen(true); }}
                        className="task-card bg-white p-4 rounded-2xl shadow-sm border-2 transition-all cursor-pointer group"
                        style={{ borderColor: isCritical ? '#fecaca' : 'var(--border)' }}>
                        <div className="flex justify-between items-start">
                          <span className="text-[10px] font-black uppercase" style={{ color: isCritical ? '#ef4444' : 'var(--text-muted)' }}>ID-{task.id}</span>
                          <button onClick={e => deleteTask(e, task.id)} className="opacity-0 group-hover:opacity-100 transition-opacity" style={{ color: 'var(--text-muted)' }}><Trash2 size={14} /></button>
                        </div>
                        <h4 className="text-sm font-bold mt-2" style={{ color: 'var(--text-primary)' }}>{task.title}</h4>
                        {task.description && <p className="text-[11px] mt-1.5 line-clamp-2" style={{ color: 'var(--text-muted)' }}>{task.description.length > 60 ? `${task.description.substring(0, 60)}...` : task.description}</p>}
                        
                        {/* Deadline */}
                        {deadlineStr && (
                          <div className="flex items-center gap-1 mt-2 text-[10px] font-bold" style={{ color: deadlineStr === 'Просрочено!' ? '#ef4444' : '#f59e0b' }}>
                            <Clock size={10} /> {deadlineStr}
                          </div>
                        )}

                        <div className="flex items-center justify-between mt-3 pt-2 border-t" style={{ borderColor: 'var(--border)' }}>
                          <span className="text-[10px] font-bold" style={{ color: 'var(--text-secondary)' }}>{task.duration_hours ? `${task.duration_hours.toFixed(1)} ч.` : '0 ч.'}</span>
                          <div className="flex items-center gap-1.5">
                            <div className="w-5 h-5 bg-blue-500 rounded-full flex items-center justify-center text-[8px] text-white font-bold cursor-pointer hover:opacity-80 transition-opacity" 
                              onClick={e => { e.stopPropagation(); onViewUser?.(task.user_id); }} title="Профиль автора">
                              {task.created_by_name?.charAt(0) || '?'}
                            </div>
                            {task.assignee_id && (
                              <span className="text-[9px] font-bold px-1.5 py-0.5 rounded cursor-pointer hover:opacity-80 transition-opacity" 
                                style={{ backgroundColor: 'var(--bg-card-hover)', color: 'var(--text-secondary)' }}
                                onClick={e => { e.stopPropagation(); onViewUser?.(task.assignee_id); }} title="Профиль исполнителя">
                                {task.assignee_name?.split(' ')[0] || 'Исп.'}
                              </span>
                            )}
                          </div>
                        </div>
                      </motion.div>
                    );
                  })}
                  </AnimatePresence>
                </div>
              </div>
            ))}
          </div>
        </div>

        <TaskModal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)} onSave={handleSaveTask} initialData={editingTask} projectId={projectId} />  
      </div>
    </div>
  );
}

KanbanBoard.propTypes = { projectId: PropTypes.oneOfType([PropTypes.number, PropTypes.string]), onTasksChange: PropTypes.func, onViewUser: PropTypes.func };
