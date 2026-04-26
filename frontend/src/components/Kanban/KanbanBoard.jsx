import { useState, useEffect } from 'react';
import { api } from '../../api/client';
import { Plus, MoreVertical } from 'lucide-react';
import TaskModal from './TaskModal'; // Импортируем модалку

const COLUMNS = [
  { id: 'todo', title: 'Бэклог', color: 'bg-gray-100' },
  { id: 'in_progress', title: 'В работе', color: 'bg-blue-50' },
  { id: 'done', title: 'Готово', color: 'bg-green-50' }
];

export default function KanbanBoard({ projectId }) {
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);

  useEffect(() => {
    if (projectId) loadTasks();
  }, [projectId]);

  async function loadTasks() {
    setLoading(true);
    try {
      const data = await api.get(`/projects/${projectId}/tasks`);
      setTasks(data || []);
    } catch (err) {
      console.error('Ошибка загрузки:', err);
    } finally {
      setLoading(false);
    }
  }

  // Функция создания задачи
  async function handleCreateTask(taskData) {
    try {
      await api.post('/tasks', taskData);
      loadTasks(); // Обновляем список после создания
    } catch (err) {
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

  if (loading) return <div className="p-8 text-center text-gray-400">Загрузка...</div>;

  return (
    <div className="h-full relative">
      {/* Кнопка "Создать задачу" */}
      <button 
        onClick={() => setIsModalOpen(true)}
        className="absolute top-4 right-8 z-10 bg-blue-600 text-white px-4 py-2 rounded-lg font-bold flex items-center gap-2 hover:bg-blue-700 shadow-md transition-all"
      >
        <Plus size={20} /> Создать задачу
      </button>

      <div className="h-full p-6 pt-16 overflow-x-auto flex gap-6">
        {COLUMNS.map(col => (
          <div key={col.id} className="flex-shrink-0 w-80 flex flex-col">
            <div className="flex items-center justify-between mb-4 px-2">
              <h3 className="font-bold text-gray-700 flex items-center gap-2">
                {col.title}
                <span className="bg-gray-200 text-gray-500 text-xs px-2 py-0.5 rounded-full">
                  {tasks.filter(t => t.status === col.id).length}
                </span>
              </h3>
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
                    <span className="text-xs font-bold text-blue-500 uppercase tracking-tighter">ID-{task.id}</span>
                    <button className="opacity-0 group-hover:opacity-100 text-gray-400 hover:text-red-500 transition-all">
                      <MoreVertical size={14}/>
                    </button>
                  </div>
                  <h4 className="text-sm font-semibold text-gray-800 leading-tight">{task.title}</h4>
                  <div className="mt-3 grid grid-cols-3 gap-1">
                    <div className="text-[9px] text-center bg-green-50 text-green-700 rounded py-0.5 font-bold">O: {task.opt}ч</div>
                    <div className="text-[9px] text-center bg-blue-50 text-blue-700 rounded py-0.5 font-bold">R: {task.real}ч</div>
                    <div className="text-[9px] text-center bg-red-50 text-red-700 rounded py-0.5 font-bold">P: {task.pess}ч</div>
                  </div>
                </div>
              ))}
              
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

      {/* Модальное окно */}
      <TaskModal 
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onCreate={handleCreateTask}
        projectId={projectId}
      />
    </div>
  );
}