import { useState, useEffect, useCallback } from 'react';
import PropTypes from 'prop-types';
import { api } from '../../api/client';
import { Plus, Trash2, AlertCircle } from 'lucide-react'; // Иконки используются ниже
import TaskModal from './TaskModal';

const COLUMNS = [
  { id: 'todo', title: 'Бэклог', color: 'bg-gray-100' },
  { id: 'in_progress', title: 'В работе', color: 'bg-blue-50' },
  { id: 'done', title: 'Готово', color: 'bg-green-50' }
];

export default function KanbanBoard({ projectId }) {
  const [tasks, setTasks] = useState([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingTask, setEditingTask] = useState(null);

  const loadTasks = useCallback(async () => {
    if (!projectId) return;
    try {
      const data = await api.get(`/projects/${projectId}/tasks`);
      setTasks(data || []);
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

  // ФУНКЦИЯ ИСПОЛЬЗУЕТСЯ в карточке ниже
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
      // Выводим реальную ошибку из бэкенда (например, "только исполнитель может...")
      alert(err.message); 
      loadTasks(); // Сбрасываем положение карточки
    }
  };

  useEffect(() => {
  loadTasks(); // Загрузка при открытии

  // ПРИКАЗ: Опрашивать бэкенд каждые 5 секунд, чтобы видеть перемещения задач коллегами
  const interval = setInterval(() => {
    loadTasks();
  }, 5000);

  return () => clearInterval(interval);
}, [loadTasks]);
  // maxScore ИСПОЛЬЗУЕТСЯ для расчета критического пути
  const maxScore = tasks.length > 0 ? Math.max(...tasks.map(t => t.priority_score || 0)) : 0;

  return (
    <div className="h-full w-full bg-gray-50 p-6 overflow-hidden">
      {/* ПРИКАЗ: Возвращаем bg-white вместо красного */}
      <div className="w-full h-full flex flex-col bg-white rounded-2xl shadow-xl border overflow-hidden transition-all">
        
        {/* Хедер 72px */}
        <div className="h-[72px] min-h-[72px] px-8 border-b flex justify-between items-center bg-white z-10">
          <div>
            <h3 className="font-bold text-gray-700 uppercase text-xs tracking-widest">Канбан-доска</h3>
            <p className="text-[10px] text-gray-400">Управляйте задачами и их статусами</p>
          </div>
          <button 
            onClick={() => { setEditingTask(null); setIsModalOpen(true); }}
            className="bg-blue-600 text-white px-6 py-2.5 rounded-xl text-xs font-bold flex items-center gap-2 hover:bg-blue-700 transition-all shadow-md"
          >
            <Plus size={16} /> СОЗДАТЬ ЗАДАЧУ
          </button>
        </div>

        <div className="flex-1 overflow-x-auto p-8">
          <div className="flex gap-8 h-full">
            {COLUMNS.map(col => (
              <div key={col.id} className="flex-shrink-0 w-80 flex flex-col h-full">
                <h3 className="font-bold text-gray-500 uppercase text-[10px] tracking-widest mb-4 px-2">
                  {col.title} · {tasks.filter(t => t.status === col.id).length}
                </h3>
                <div 
                  onDragOver={e => e.preventDefault()}
                  onDrop={e => updateTaskStatus(e.dataTransfer.getData('taskId'), col.id)}
                  className={`flex-1 rounded-2xl p-3 space-y-4 border-2 border-dashed border-transparent ${col.color} overflow-y-auto`}
                >
                  {tasks.filter(t => t.status === col.id).map(task => {
                    const isCritical = task.priority_score >= (maxScore * 0.8) && task.priority_score > 0 && task.status !== 'done';
                    
                    return (
                      <div 
                        key={task.id}
                        draggable
                        onDragStart={e => e.dataTransfer.setData('taskId', task.id)}
                        onClick={() => { setEditingTask(task); setIsModalOpen(true); }}
                        className={`bg-white p-5 rounded-2xl shadow-sm border-2 transition-all hover:shadow-lg group cursor-pointer ${
                          isCritical ? 'border-red-200 bg-red-50/30' : 'border-transparent hover:border-gray-200'
                        }`}
                      >
                        <div className="flex justify-between items-start">
                          <span className={`text-[10px] font-black uppercase ${isCritical ? 'text-red-500' : 'text-blue-500'}`}>
                            ID-{task.id}
                          </span>
                          <button 
                            onClick={e => deleteTask(e, task.id)} 
                            className="opacity-0 group-hover:opacity-100 text-gray-400 hover:text-red-500 transition-opacity"
                          >
                            <Trash2 size={14} />
                          </button>
                        </div>
                        <h4 className="text-sm font-bold text-gray-800 mt-2">{task.title}</h4>
<div className="flex items-center gap-2 mt-2">
  <div className="w-4 h-4 bg-blue-500 rounded-full flex items-center justify-center text-[8px] text-white">
    {task.created_by_name?.charAt(0) || '?'}
  </div>
  <span className="text-[10px] text-gray-400">
    {task.created_by_name || 'Загрузка...'}
  </span>
</div>



{/* НОВОЕ: Отображение автора */}
<p className="text-[9px] text-gray-400 mt-1 uppercase font-medium">
  Автор: {task.created_by_name || 'Загрузка...'}
</p>

{/* ВСТАВЬ КОД ИСПОЛНИТЕЛЯ СЮДА */}
{task.assignee_id ? (
   <p className="text-[10px] text-gray-500 mt-1 font-medium bg-gray-100 p-1.5 rounded w-fit">
      Исп: <span className="font-bold text-gray-700">{task.assignee_name}</span>
   </p>
) : (
   <p className="text-[10px] text-gray-400 mt-1 font-medium italic">Не назначен</p>
)}
                        
{isCritical && (
                          <div className="flex items-center gap-1 text-[9px] text-red-600 font-bold mt-3 bg-red-100 w-fit px-2.5 py-1 rounded-full">
                            <AlertCircle size={10} /> КРИТИЧЕСКИЙ ВЕС
                          </div>
                        )}

                        <div className="mt-4 pt-4 border-t border-gray-50 flex items-center justify-between">
                          <div className="text-[10px] bg-gray-100 text-gray-600 px-2.5 py-1.5 rounded-lg font-black">
                            {task.priority_score.toFixed(1)}ч
                          </div>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            ))}
          </div>
        </div>

        <TaskModal 
          isOpen={isModalOpen} 
          onClose={() => setIsModalOpen(false)} 
          onSave={handleSaveTask} 
          initialData={editingTask} 
          projectId={projectId} 
        />
      </div>
    </div>
  );
}


KanbanBoard.propTypes = {
  projectId: PropTypes.oneOfType([PropTypes.number, PropTypes.string])
};