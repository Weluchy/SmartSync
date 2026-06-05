import { useState, useEffect, useCallback } from 'react';
import PropTypes from 'prop-types';
import { motion, AnimatePresence } from 'framer-motion';
import { api } from '../../api/client';
import { Plus, Trash2, Search, Filter, Clock, CheckSquare, Square, ArrowUpDown, Flame } from 'lucide-react';
import TaskModal from './TaskModal';
import { toast } from 'react-hot-toast';

const COLUMNS = [
  { id: 'todo', title: 'Бэклог', color: 'bg-gray-100' },
  { id: 'in_progress', title: 'В работе', color: 'bg-blue-50' },
  { id: 'done', title: 'Готово', color: 'bg-green-50' }
];

const SORT_OPTIONS = [
  { id: 'created', label: 'По дате' },
  { id: 'priority', label: 'По приоритету' },
  { id: 'deadline', label: 'По дедлайну' },
  { id: 'title', label: 'По названию' },
];

export default function KanbanBoard({ projectId, onTasksChange, onViewUser }) {
  const [tasks, setTasks] = useState([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingTask, setEditingTask] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [filterStatus, setFilterStatus] = useState('all');
  const [filterAssignee, setFilterAssignee] = useState('all');
  const [filterMilestone, setFilterMilestone] = useState('all');
  const [members, setMembers] = useState([]);
  const [milestones, setMilestones] = useState([]);
  const [sortBy, setSortBy] = useState('created');
  const [selectedIds, setSelectedIds] = useState(new Set());
  const [editingTitleId, setEditingTitleId] = useState(null);
  const [editingTitle, setEditingTitle] = useState('');
  const [showCritical, setShowCritical] = useState(false);

  const loadTasks = useCallback(async () => {
    if (!projectId) return;
    try {
      const data = await api.get(`/projects/${projectId}/tasks`);
      // Эндпоинт возвращает {tasks: [...], dependencies: [...]}
      const taskList = data?.tasks || data || [];
      setTasks(taskList);
      if (onTasksChange) onTasksChange(taskList);
    } catch (err) { console.error(err); }
  }, [projectId, onTasksChange]);

  const loadMembers = useCallback(async () => {
    if (!projectId) return;
    try {
      const data = await api.get(`/projects/${projectId}/members`);
      setMembers(data || []);
    } catch (err) { console.error(err); }
  }, [projectId]);

  const loadMilestones = useCallback(async () => {
    if (!projectId) return;
    try {
      const data = await api.get(`/projects/${projectId}/milestones`);
      setMilestones(data || []);
    } catch { /* ignore */ }
  }, [projectId]);

  const handleSaveTask = async (taskData) => {
    try {
      if (editingTask) await api.put(`/tasks/${editingTask.id}`, taskData);
      else await api.post('/tasks', taskData);
      setIsModalOpen(false);
      toast.success(editingTask ? 'Задача обновлена' : 'Задача создана', {
        style: { background: '#1a1a2e', color: '#7ac9a7', border: '1px solid #7ac9a7' }
      });
      setTimeout(loadTasks, 300);
    } catch (err) { 
      toast.error(err.message, {
        style: { background: '#1a1a2e', color: '#f87171', border: '1px solid #f87171' }
      });
    }
  };

  const deleteTask = async (e, id) => {
    e.stopPropagation();
    try {
      // heal=true — автоматическое сшивание графа: 
      // зависевшие от удаляемой задачи переходят на те задачи, от которых зависела удаляемая
      await api.delete(`/tasks/${id}?heal=true`);
      toast('Задача удалена. Граф перестроен.', {
        icon: '🗑️',
        style: { background: '#1a1a2e', color: '#e4e4e7', border: '1px solid #f87171' }
      });
      setTimeout(loadTasks, 300);
    } catch (err) { 
      toast.error(err.message, {
        style: { background: '#1a1a2e', color: '#f87171', border: '1px solid #f87171' }
      });
    }
  };

  const updateTaskStatus = async (taskId, newStatus) => {
    try {
      await api.patch(`/tasks/${taskId}/status`, { status: newStatus });
      setTimeout(loadTasks, 300);
    } catch (err) { 
      toast.error(err.message, {
        style: { background: '#1a1a2e', color: '#f87171', border: '1px solid #f87171' }
      });
      loadTasks(); 
    }
  };

  // Inline-редактирование названия
  const startEditTitle = (task) => {
    setEditingTitleId(task.id);
    setEditingTitle(task.title);
  };

const saveEditTitle = async () => {
    if (!editingTitleId || !editingTitle.trim()) {
      setEditingTitleId(null);
      return;
    }
    try {
      // Находим текущую задачу, чтобы не потерять ее данные
      const taskToUpdate = tasks.find(t => t.id === editingTitleId);
      
      await api.put(`/tasks/${editingTitleId}`, { 
        ...taskToUpdate, // Отправляем все старые поля (opt, real, pess, assignee_id)
        title: editingTitle.trim() // Перезаписываем только название
      });
      
      setEditingTitleId(null);
      setTimeout(loadTasks, 300);
    } catch (err) {
      toast.error(err.message, {
        style: { background: '#1a1a2e', color: '#f87171', border: '1px solid #f87171' }
      });
      setEditingTitleId(null);
    }
  };

  // Batch-операции
  const toggleSelect = (id) => {
    setSelectedIds(prev => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const selectAll = () => {
    if (selectedIds.size === filteredTasks.length) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(filteredTasks.map(t => t.id)));
    }
  };

  const batchDelete = async () => {
    if (selectedIds.size === 0) return;
    for (const id of selectedIds) {
      try { await api.delete(`/tasks/${id}`); } catch { /* ignore single failures */ }
    }
    toast(`${selectedIds.size} задач удалено`, {
      icon: '🗑️',
      style: { background: '#1a1a2e', color: '#e4e4e7', border: '1px solid #f87171' }
    });
    setSelectedIds(new Set());
    setTimeout(loadTasks, 300);
  };

  const batchStatus = async (status) => {
    if (selectedIds.size === 0) return;
    for (const id of selectedIds) {
      try { await api.patch(`/tasks/${id}/status`, { status }); } catch { /* ignore single failures */ }
    }
    toast(`Статус ${selectedIds.size} задач изменён на "${status}"`, {
      style: { background: '#1a1a2e', color: '#60a5fa', border: '1px solid #60a5fa' }
    });
    setSelectedIds(new Set());
    setTimeout(loadTasks, 300);
  };

  useEffect(() => {
    loadTasks();
    loadMembers();
    loadMilestones();
    
    if (Notification.permission === 'default') {
      Notification.requestPermission();
    }

    const userId = Number(localStorage.getItem('userId'));
    const ws = new WebSocket('ws://localhost:8000/ws');
    ws.onopen = () => console.log('✅ WebSocket подключен');
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.project_id === Number(projectId)) {
          loadTasks();
          
          // Push-уведомление через тост, если назначили этого пользователя
          if (data.assignee_id && Number(data.assignee_id) === userId) {
            toast(`Вас назначили ответственным за задачу`, {
              icon: '📋',
              duration: 6000,
              style: { background: '#1a1a2e', color: '#c4b5fd', border: '1px solid #8b5cf6' }
            });
          }

          if (Notification.permission === 'granted' && data.action) {
            new Notification('SmartSync', { body: `Изменение в задаче: ${data.action}` });
          }
        }
      } catch (e) { console.error(e); }
    };
    return () => { if (ws.readyState === 1) ws.close(); };
  }, [loadTasks, loadMembers, loadMilestones, projectId]);

  // Фильтрация
  let filteredTasks = tasks.filter(t => {
    if (searchQuery && !t.title.toLowerCase().includes(searchQuery.toLowerCase())) return false;
    if (filterStatus !== 'all' && t.status !== filterStatus) return false;
    if (filterAssignee === 'unassigned' && t.assignee_id) return false;
    if (filterAssignee === 'assigned' && !t.assignee_id) return false;
    if (filterAssignee === 'me' && t.assignee_id !== Number(localStorage.getItem('userId'))) return false;
    if (filterAssignee !== 'all' && filterAssignee !== 'me' && filterAssignee !== 'assigned' && filterAssignee !== 'unassigned') {
      if (t.assignee_id !== Number(filterAssignee)) return false;
    }
    if (filterMilestone !== 'all') {
      if (Number(t.milestone_id) !== Number(filterMilestone)) return false;
    }
    return true;
  });

  // Сортировка
  filteredTasks = [...filteredTasks].sort((a, b) => {
    switch (sortBy) {
      case 'priority': return (b.priority_score || 0) - (a.priority_score || 0);
      case 'title': return (a.title || '').localeCompare(b.title || '');
      case 'deadline': {
        const da = a.duration_hours ? new Date(a.created_at).getTime() + a.duration_hours * 3600000 : Infinity;
        const db = b.duration_hours ? new Date(b.created_at).getTime() + b.duration_hours * 3600000 : Infinity;
        return da - db;
      }
      default: return new Date(b.created_at) - new Date(a.created_at);
    }
  });

  const maxScore = filteredTasks.length > 0 ? Math.max(...filteredTasks.map(t => t.priority_score || 0)) : 0;

  return (
    <div className="h-full w-full p-6 overflow-hidden" style={{ backgroundColor: 'var(--bg-page)' }}>
      <div className="w-full h-full flex flex-col rounded-2xl shadow-xl border overflow-hidden" style={{ backgroundColor: 'var(--bg-card)', borderColor: 'var(--border)' }}>
        {/* Верхняя панель */}
        <div className="h-[72px] min-h-[72px] px-8 border-b flex items-center gap-4" style={{ borderColor: 'var(--border)' }}>
          <div className="flex items-center gap-2 flex-1">
            <Search size={14} style={{ color: 'var(--text-muted)' }} />
            <input placeholder="Поиск задач..." value={searchQuery} onChange={e => setSearchQuery(e.target.value)}
              className="flex-1 bg-transparent text-sm outline-none" style={{ color: 'var(--text-primary)' }} />
          </div>
          <div className="flex items-center gap-2">
            <ArrowUpDown size={14} style={{ color: 'var(--text-muted)' }} />
            <select value={sortBy} onChange={e => setSortBy(e.target.value)}
              className="text-xs border rounded-lg p-1.5 outline-none" style={{ backgroundColor: 'var(--bg-input)', color: 'var(--text-secondary)', borderColor: 'var(--border)' }}>
              {SORT_OPTIONS.map(o => <option key={o.id} value={o.id}>{o.label}</option>)}
            </select>
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
            {/* Фильтр по вехам */}
            <select value={filterMilestone} onChange={e => setFilterMilestone(e.target.value)}
              className="text-xs border rounded-lg p-1.5 outline-none" style={{ backgroundColor: 'var(--bg-input)', color: 'var(--text-secondary)', borderColor: 'var(--border)' }}>
              <option value="all">Все вехи</option>
              <option value="none">Без вехи</option>
              {milestones.map(m => <option key={m.id} value={m.id}>{m.title}</option>)}
            </select>
          </div>
          {/* Тумблер критического пути */}
          <button
            onClick={() => setShowCritical(!showCritical)}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-bold transition-all ${
              showCritical ? 'bg-red-50 text-red-600 border border-red-200' : 'border border-gray-200 text-gray-500 hover:bg-gray-50'
            }`}
            title="Показать узкие места"
          >
            <Flame size={14} className={showCritical ? 'text-red-500' : ''} />
            Крит. путь
          </button>
          {/* Batch-панель */}
          {selectedIds.size > 0 && (
            <div className="flex items-center gap-2 bg-blue-50 px-3 py-1.5 rounded-lg">
              <span className="text-xs font-bold text-blue-600">{selectedIds.size}</span>
              <button onClick={() => batchStatus('todo')} className="text-[10px] bg-gray-200 px-2 py-1 rounded font-bold hover:bg-gray-300">→ Бэклог</button>
              <button onClick={() => batchStatus('in_progress')} className="text-[10px] bg-blue-200 px-2 py-1 rounded font-bold hover:bg-blue-300">→ В работе</button>
              <button onClick={() => batchStatus('done')} className="text-[10px] bg-green-200 px-2 py-1 rounded font-bold hover:bg-green-300">→ Готово</button>
              <button onClick={batchDelete} className="text-[10px] bg-red-200 px-2 py-1 rounded font-bold hover:bg-red-300 text-red-700">🗑</button>
            </div>
          )}
          <button onClick={() => { setEditingTask(null); setIsModalOpen(true); }}
            className="bg-blue-600 text-white px-5 py-2 rounded-xl text-xs font-bold flex items-center gap-2 hover:bg-blue-700 shadow-md">
            <Plus size={16} /> СОЗДАТЬ
          </button>
        </div>

        {/* Колонки */}
        <div className="flex-1 overflow-x-auto p-6">
          <div className="flex gap-6 h-full">
            {COLUMNS.map(col => (
              <div key={col.id} className="flex-shrink-0 w-80 flex flex-col h-full">
                <div className="flex items-center justify-between mb-3 px-2">
                  <h3 className="font-bold text-xs uppercase tracking-widest" style={{ color: 'var(--text-secondary)' }}>
                    {col.title} · {filteredTasks.filter(t => t.status === col.id).length}
                  </h3>
                  <button onClick={selectAll} style={{ color: 'var(--text-muted)' }} className="hover:text-blue-600 text-[10px]">
                    {selectedIds.size === filteredTasks.length ? <CheckSquare size={12} /> : <Square size={12} />}
                  </button>
                </div>
                <div onDragOver={e => e.preventDefault()} onDrop={e => updateTaskStatus(e.dataTransfer.getData('taskId'), col.id)}
                  className="flex-1 rounded-2xl p-3 space-y-3 border-2 border-dashed border-transparent overflow-y-auto"
                  style={{ backgroundColor: col.color === 'bg-gray-100' ? 'var(--kanban-bg)' : col.color === 'bg-blue-50' ? 'rgba(59,130,246,0.05)' : 'rgba(34,197,94,0.05)' }}>
                  <AnimatePresence>
                  {filteredTasks.filter(t => t.status === col.id).map(task => {
                    const isCritical = task.priority_score >= (maxScore * 0.8) && task.priority_score > 0 && task.status !== 'done';
                    const createdDate = new Date(task.created_at);
                    const isDateValid = createdDate.getFullYear() > 2000;
                    
                    // СЧИТАЕМ ВРЕМЯ НА ЛЕТУ: если бэкенд отдал 0, применяем формулу PERT
                    const duration = task.duration_hours || ((task.opt + 4 * task.real + task.pess) / 6);
                    
                    let deadlineStr = '', exactDate = '';
                    if (isDateValid && duration > 0) {
                      const deadline = new Date(createdDate.getTime() + (duration * 3600000));
                      exactDate = deadline.toLocaleString('ru-RU', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' });
                      if (task.status === 'done') {
                        deadlineStr = 'Выполнено';
                      } else {
                        const diff = deadline.getTime() - Date.now();
                        deadlineStr = diff < 0 ? 'Просрочено!' : `Ост. ${Math.floor(diff / 3600000)}ч ${Math.floor((diff % 3600000) / 60000)}м`;
                      }
                    }
                    return (
                      <motion.div key={task.id} layout initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, scale: 0.8 }} transition={{ duration: 0.2 }}
  draggable 
  onDragStart={e => { e.stopPropagation(); e.dataTransfer.setData('taskId', task.id); e.currentTarget.classList.add('dragging'); }}
  onDragEnd={e => { e.stopPropagation(); e.currentTarget.classList.remove('dragging'); }}
  onClick={(e) => { e.stopPropagation(); setEditingTask(task); setIsModalOpen(true); }}
  className={`task-card relative bg-white p-4 rounded-2xl shadow-sm border-2 transition-all cursor-pointer group ${
    showCritical && isCritical ? 'ring-2 ring-red-400 ring-offset-2' : ''
  } ${showCritical && !isCritical ? 'opacity-40' : ''}`}
  style={{ borderColor: isCritical ? '#fecaca' : 'var(--border)' }}>
                        <div className="flex justify-between items-start">
                          <div className="flex items-center gap-2">
                            <button onClick={e => { e.stopPropagation(); toggleSelect(task.id); }}
                              className="opacity-0 group-hover:opacity-100 transition-opacity" style={{ color: 'var(--text-muted)' }}>
                              {selectedIds.has(task.id) ? <CheckSquare size={12} /> : <Square size={12} />}
                            </button>
                            <span className="text-[10px] font-black uppercase" style={{ color: isCritical ? '#ef4444' : 'var(--text-muted)' }}>ID-{task.id}</span>
                            {isCritical && <Flame size={12} className="text-red-500" />}
                          </div>
                          <button onClick={e => deleteTask(e, task.id)} className="opacity-0 group-hover:opacity-100 transition-opacity" style={{ color: 'var(--text-muted)' }}><Trash2 size={14} /></button>
                        </div>

                        {/* Inline-редактирование названия */}
                        {editingTitleId === task.id ? (
                          <input autoFocus value={editingTitle} onChange={e => setEditingTitle(e.target.value)}
                            onBlur={saveEditTitle} onKeyDown={e => { if (e.key === 'Enter') saveEditTitle(); if (e.key === 'Escape') setEditingTitleId(null); }}
                            className="w-full text-sm font-bold mt-2 border rounded p-1 outline-none focus:ring-2 focus:ring-blue-500" />
                        ) : (
                          <h4 className="text-sm font-bold mt-2 cursor-pointer hover:text-blue-600 transition-colors"
                            onClick={e => { e.stopPropagation(); startEditTitle(task); }}
                            style={{ color: 'var(--text-primary)' }}>
                            {task.title}
                          </h4>
                        )}

                        {task.description && <p className="text-[11px] mt-1.5 line-clamp-2" style={{ color: 'var(--text-muted)' }}>{task.description.length > 60 ? `${task.description.substring(0, 60)}...` : task.description}</p>}
                        
                        {isDateValid && duration > 0 ? (
                          <div className={`mt-3 rounded-xl p-3 border ${
                            task.status === 'done' ? 'bg-green-50 border-green-200' : 
                            deadlineStr === 'Просрочено!' ? 'bg-red-50 border-red-200' : 
                            'bg-amber-50 border-amber-200'
                          }`}>
                            <div className="flex items-center justify-between">
                              <div className="flex items-center gap-1.5">
                                <Clock size={12} className={
                                  task.status === 'done' ? 'text-green-600' : 
                                  deadlineStr === 'Просрочено!' ? 'text-red-600' : 
                                  'text-amber-600'
                                } />
                                <span className={`text-[11px] font-black ${
                                  task.status === 'done' ? 'text-green-700' : 
                                  deadlineStr === 'Просрочено!' ? 'text-red-700' : 
                                  'text-amber-700'
                                }`}>
                                  {task.status === 'done' ? '✅ Выполнено' : deadlineStr}
                                </span>
                              </div>
                              <span className={`text-[9px] font-bold ${
                                task.status === 'done' ? 'text-green-500' : 
                                deadlineStr === 'Просрочено!' ? 'text-red-500' : 
                                'text-amber-500'
                              }`}>
                                {duration.toFixed(1)} ч
                              </span>
                            </div>
                            {task.status !== 'done' && (
                              <div className="text-[10px] mt-1 font-medium" style={{ color: 'var(--text-muted)' }}>
                                Дедлайн: {exactDate}
                              </div>
                            )}
                          </div>
                        ) : (
                          <div className="mt-2 text-[10px]" style={{ color: 'var(--text-muted)' }}>
                            <div className="flex items-center gap-1"><Clock size={10} /> Без срока</div>
                          </div>
                        )}

                        {/* Дата создания */}
                        <div className="text-[9px] mt-1" style={{ color: 'var(--text-muted)' }}>
                          Создано: {createdDate.toLocaleDateString('ru-RU')}
                        </div>

                        <div className="flex items-center justify-between mt-3 pt-2 border-t" style={{ borderColor: 'var(--border)' }}>
                          <span className="text-[10px] font-bold" style={{ color: 'var(--text-secondary)' }}>{duration.toFixed(1)} ч.</span>
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

        <TaskModal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)} onSave={handleSaveTask} initialData={editingTask} projectId={projectId} milestones={milestones} onViewUser={onViewUser} />  
      </div>
    </div>
  );
}

KanbanBoard.propTypes = { projectId: PropTypes.oneOfType([PropTypes.number, PropTypes.string]), onTasksChange: PropTypes.func, onViewUser: PropTypes.func };