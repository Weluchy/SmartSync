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
    // 1. Создаем временный объект задачи для мгновенного отображения
    const tempTask = {
      ...taskData,
      id: Date.now(), // Временный ID
      status: 'todo',
      priority_score: 0,
      loading: true // Флаг, чтобы визуально отличить "сохраняемую" задачу
    };

    // 2. Обновляем состояние локально (Optimistic Update)
    setTasks(prev => [...prev, tempTask]);

    try {
      const savedTask = await api.post('/tasks', taskData);
      
      // 3. Когда сервер ответил, заменяем временную задачу на реальную
      setTasks(prev => prev.map(t => t.id === tempTask.id ? savedTask : t));
      
      // 4. Через секунду обновляем всё, чтобы подтянулись приоритеты от движка
      setTimeout(loadTasks, 1000); 
    } catch (err) {
      // Если ошибка — удаляем временную задачу
      setTasks(prev => prev.filter(t => t.id !== tempTask.id));
      alert('Ошибка при создании задачи: ' + err.message);
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
                  <div className="mt-3 grid grid-cols-3 gap-1">
                    <div className="text-[9px] text-center bg-green-50 text-green-700 rounded py-0.5 font-bold">O: {task.opt}ч</div>
                    <div className="text-[9px] text-center bg-blue-50 text-blue-700 rounded py-0.5 font-bold">R: {task.real}ч</div>
                    <div className="text-[9px] text-center bg-red-50 text-red-700 rounded py-0.5 font-bold">P: {task.pess}ч</div>
                  </div>
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