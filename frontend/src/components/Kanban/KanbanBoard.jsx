import { useState, useEffect, useCallback } from 'react';
import PropTypes from 'prop-types';
import { api } from '../../api/client';
import { Plus } from 'lucide-react';
import TaskModal from './TaskModal';

const COLUMNS = [
  { id: 'todo', title: 'Бэклог', color: 'bg-gray-100' },
  { id: 'in_progress', title: 'В работе', color: 'bg-blue-50' },
  { id: 'done', title: 'Готово', color: 'bg-green-50' }
];

export default function KanbanBoard({ projectId }) {
  // --- ВСЕ ХУКИ ДОЛЖНЫ БЫТЬ ТУТ (ВНУТРИ ФУНКЦИИ) ---
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const loadTasks = useCallback(async () => {
    if (!projectId) return;
    setLoading(true);
    try {
      const data = await api.get(`/projects/${projectId}/tasks`);
      setTasks(data || []);
    } catch (err) {
      console.error('Ошибка загрузки задач:', err);
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => {
    loadTasks();
  }, [loadTasks]);

  async function handleCreateTask(taskData) {
    try {
      await api.post('/tasks', taskData);
      loadTasks();
    } catch {
      alert('Ошибка при создании задачи');
    }
  }

  async function updateTaskStatus(taskId, newStatus) {
    try {
      await api.patch(`/tasks/${taskId}/status`, { status: newStatus });
      setTasks(prev => prev.map(t => 
        t.id === parseInt(taskId) ? { ...t, status: newStatus } : t
      ));
    } catch (err) {
      console.error(err);
    }
  }
  // --- КОНЕЦ ЛОГИКИ ХУКОВ ---

  if (loading) return <div className="p-8 text-center text-gray-400 font-medium">Загрузка задач...</div>;

  return (
    <div className="h-full relative bg-gray-50">
      <button 
        onClick={() => setIsModalOpen(true)}
        className="absolute top-4 right-8 z-10 bg-blue-600 text-white px-5 py-2.5 rounded-xl font-bold flex items-center gap-2 hover:bg-blue-700 shadow-lg shadow-blue-200 transition-all"
      >
        <Plus size={20} /> Создать задачу
      </button>

      <div className="h-full p-6 pt-20 overflow-x-auto flex gap-6">
        {COLUMNS.map(col => (
          <div key={col.id} className="flex-shrink-0 w-80 flex flex-col">
            <h3 className="font-bold text-gray-500 uppercase text-[11px] tracking-widest mb-4 px-2">
              {col.title} · {tasks.filter(t => t.status === col.id).length}
            </h3>
            <div 
              onDragOver={(e) => e.preventDefault()}
              onDrop={(e) => updateTaskStatus(e.dataTransfer.getData('taskId'), col.id)}
              className={`flex-1 rounded-2xl p-2 space-y-3 border-2 border-dashed border-transparent transition-colors ${col.color} min-h-[200px]`}
            >
              {tasks.filter(t => t.status === col.id).map(task => (
                <div 
                  key={task.id}
                  draggable
                  onDragStart={(e) => e.dataTransfer.setData('taskId', task.id)}
                  className="bg-white p-4 rounded-xl shadow-sm border border-gray-200 cursor-grab active:cursor-grabbing hover:shadow-md transition-all group"
                >
                  <span className="text-[10px] font-black text-blue-500 uppercase">ID-{task.id}</span>
                  <h4 className="text-sm font-bold text-gray-800 mt-1">{task.title}</h4>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>

      <TaskModal 
        isOpen={isModalOpen} 
        onClose={() => setIsModalOpen(false)} 
        onCreate={handleCreateTask} 
        projectId={projectId} 
      />
    </div>
  );
}

KanbanBoard.propTypes = {
  projectId: PropTypes.oneOfType([PropTypes.number, PropTypes.string])
};