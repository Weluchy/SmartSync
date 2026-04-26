import { useState, useEffect } from 'react';
import { api } from '../../api/client';
import { Plus, MoreVertical } from 'lucide-react';

const COLUMNS = [
  { id: 'todo', title: 'Бэклог', color: 'bg-gray-100' },
  { id: 'in_progress', title: 'В работе', color: 'bg-blue-50' },
  { id: 'done', title: 'Готово', color: 'bg-green-50' }
];

export default function KanbanBoard({ projectId }) {
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (projectId) loadTasks();
  }, [projectId]);

  async function loadTasks() {
    setLoading(true);
    try {
      // Подставляем твой эндпоинт из бэкенда
      const data = await api.get(`/projects/${projectId}/tasks`);
      setTasks(data || []);
    } catch (err) {
      console.error('Ошибка загрузки задач:', err);
    } finally {
      setLoading(false);
    }
  }

  async function updateTaskStatus(taskId, newStatus) {
  try {
    // Добавляем /status в конец URL
    await api.patch(`/tasks/${taskId}/status`, { status: newStatus });
    
    setTasks(prev => prev.map(t => 
      t.id === parseInt(taskId) ? { ...t, status: newStatus } : t
    ));
  } catch (err) {
    console.error('Ошибка обновления:', err);
    alert('Не удалось переместить задачу');
  }
}

  if (loading) return <div className="p-8 text-center text-gray-400">Загрузка задач...</div>;

  return (
    <div className="h-full p-6 overflow-x-auto flex gap-6">
      {COLUMNS.map(col => (
        <div key={col.id} className="flex-shrink-0 w-80 flex flex-col">
          <div className="flex items-center justify-between mb-4 px-2">
            <h3 className="font-bold text-gray-700 flex items-center gap-2">
              {col.title}
              <span className="bg-gray-200 text-gray-500 text-xs px-2 py-0.5 rounded-full">
                {tasks.filter(t => t.status === col.id).length}
              </span>
            </h3>
            <button className="text-gray-400 hover:text-gray-600"><Plus size={18} /></button>
          </div>

          <div className={`flex-1 rounded-xl p-2 space-y-3 border-2 border-transparent transition-colors ${col.color}`}>
            {tasks.filter(t => t.status === col.id).map(task => (
              <div 
                key={task.id}
                draggable
                onDragStart={(e) => e.dataTransfer.setData('taskId', task.id)}
                className="bg-white p-4 rounded-lg shadow-sm border border-gray-200 cursor-grab active:cursor-grabbing hover:shadow-md transition-shadow group"
              >
                <div className="flex justify-between items-start mb-2">
                  <span className="text-xs font-bold text-blue-500 uppercase tracking-tighter">Task-{task.id}</span>
                  <button className="opacity-0 group-hover:opacity-100 text-gray-400"><MoreVertical size={14}/></button>
                </div>
                <h4 className="text-sm font-semibold text-gray-800 leading-tight">{task.title}</h4>
                {task.priority && (
                  <div className="mt-3 flex items-center gap-2">
                    <div className={`w-2 h-2 rounded-full ${task.priority > 5 ? 'bg-red-500' : 'bg-green-500'}`} />
                    <span className="text-[10px] text-gray-400 font-bold uppercase">Приоритет: {task.priority}</span>
                  </div>
                )}
              </div>
            ))}
            
            {/* Зона для дропа */}
            <div 
              onDragOver={(e) => e.preventDefault()}
              onDrop={(e) => {
                const taskId = e.dataTransfer.getData('taskId');
                updateTaskStatus(taskId, col.id);
              }}
              className="h-20"
            />
          </div>
        </div>
      ))}
    </div>
  );
}