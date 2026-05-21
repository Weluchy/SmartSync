import { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { X } from 'lucide-react';
import { api } from '../../api/client'; 

export default function TaskModal({ isOpen, onClose, onSave, projectId, initialData }) {
  const [formData, setFormData] = useState({
    title: '',
    opt: 1,
    real: 2,
    pess: 3,
    status: 'todo'
  });
  const [members, setMembers] = useState([]);
  
  // Добавляем состояние для хранения логов (истории)
  const [logs, setLogs] = useState([]);

  useEffect(() => {
    if (isOpen && projectId) {
      api.get(`/projects/${projectId}/members`).then(data => setMembers(data || []));
    }
    
    if (initialData) {
      setFormData({ ...initialData, assignee_id: initialData.assignee_id || '' });
      
      // ИССЛЕДОВАНИЕ: Загружаем логи только если мы открыли существующую задачу
      if (initialData.id) {
        api.get(`/logs/${initialData.id}`)
          .then(res => setLogs(res.data || []))
          .catch(err => console.error("Ошибка загрузки логов:", err));
      }
    } else {
      setFormData({ title: '', opt: 1, real: 2, pess: 3, status: 'todo', assignee_id: '' });
      setLogs([]); // Очищаем логи при создании новой задачи
    }
  }, [initialData, isOpen, projectId]); 

  if (!isOpen) return null;

  const handleSubmit = (e) => {
    e.preventDefault();
    onSave({
      ...formData,
      assignee_id: formData.assignee_id ? parseInt(formData.assignee_id) : null,
      opt: parseFloat(formData.opt),
      real: parseFloat(formData.real),
      pess: parseFloat(formData.pess),
      project_id: parseInt(projectId, 10) // <-- ИСПРАВЛЕНО: Явно делаем числом
    });
  };

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-2xl shadow-2xl w-full max-w-md max-h-[90vh] overflow-y-auto">
        <div className="flex justify-between items-center p-6 border-b sticky top-0 bg-white z-10">
          <h2 className="text-xl font-bold text-gray-800">
            {initialData ? 'Редактировать задачу' : 'Новая задача'}
          </h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 transition-colors">
            <X size={24} />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-bold text-gray-700 mb-1">Название задачи</label>
            <input
              required
              className="w-full border rounded-lg p-2.5 bg-gray-50 focus:ring-2 focus:ring-blue-500 outline-none"
              value={formData.title}
              onChange={e => setFormData({...formData, title: e.target.value})}
            />
          </div>

          <div className="grid grid-cols-3 gap-3">
            <div>
              <label className="block text-[10px] font-black text-green-600 uppercase mb-1">Оптим.</label>
              <input
                type="number"
                min="0"
                step="0.1"
                required
                className="w-full border rounded-lg p-2 bg-gray-50"
                value={formData.opt}
                onChange={e => setFormData({...formData, opt: e.target.value})}
              />
            </div>
            <div>
              <label className="block text-[10px] font-black text-blue-600 uppercase mb-1">Реал.</label>
              <input
                type="number"
                min="0"
                step="0.1"
                required
                className="w-full border rounded-lg p-2 bg-gray-50"
                value={formData.real}
                onChange={e => setFormData({...formData, real: e.target.value})}
              />
            </div>
            <div>
              <label className="block text-[10px] font-black text-red-600 uppercase mb-1">Пессим.</label>
              <input
                type="number"
                min="0"
                step="0.1"
                required
                className="w-full border rounded-lg p-2 bg-gray-50"
                value={formData.pess}
                onChange={e => setFormData({...formData, pess: e.target.value})}
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-bold text-gray-700 mb-1">Исполнитель</label>
            <select
              className="w-full border rounded-lg p-2.5 bg-gray-50 outline-none"
              value={formData.assignee_id || ''}
              onChange={e => setFormData({...formData, assignee_id: e.target.value})}
            >
              <option value="">Не назначен</option>
              {members.map(m => (
                <option key={m.user_id} value={m.user_id}>{m.username}</option>
              ))}
            </select>
          </div>

          {/* ИССЛЕДОВАНИЕ: Блок с логами. Показываем только при редактировании (initialData) */}
          {initialData && (
            <div className="pt-4 border-t mt-4">
              <h4 className="font-bold text-sm text-gray-700 mb-2">История (Аудит)</h4>
              <div className="bg-gray-50 p-3 rounded h-32 overflow-y-auto border">
                {logs.length > 0 ? (
                  <ul className="space-y-2 text-xs">
                    {logs.map((log, i) => (
                      <li key={i} className="border-b border-gray-200 pb-2 last:border-0">
                        <div className="flex justify-between font-medium text-blue-700">
                          <span>{log.action === 'updated' ? 'Обновление' : 'Создание'}</span>
                          <span className="text-gray-400 font-normal">
                            {new Date(log.timestamp).toLocaleString('ru-RU')}
                          </span>
                        </div>
                        {log.payload && log.payload.status && (
                          <p className="text-gray-600 mt-0.5">Новый статус: <span className="font-semibold">{log.payload.status}</span></p>
                        )}
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-gray-500 text-xs italic text-center mt-2">История пуста</p>
                )}
              </div>
            </div>
          )}

          <div className="pt-4 flex gap-3">
            <button type="button" onClick={onClose} className="flex-1 py-3 border rounded-xl font-bold hover:bg-gray-50 transition-colors">Отмена</button>
            <button type="submit" className="flex-1 py-3 bg-blue-600 text-white rounded-xl font-bold hover:bg-blue-700 transition-colors">
              {initialData ? 'Сохранить' : 'Создать'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

TaskModal.propTypes = {
  isOpen: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  onSave: PropTypes.func.isRequired,
  projectId: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
  initialData: PropTypes.object
};