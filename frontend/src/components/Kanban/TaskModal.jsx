import { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { api } from '../../api/client';
import { X } from 'lucide-react';

export default function TaskModal({ isOpen, onClose, onSave, initialData, projectId }) {
  const [title, setTitle] = useState('');
  const [opt, setOpt] = useState(1);
  const [real, setReal] = useState(2);
  const [pess, setPess] = useState(4);
  const [assigneeId, setAssigneeId] = useState('');
  const [members, setMembers] = useState([]);
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!isOpen) return;

    if (initialData) {
      setTitle(initialData.title || '');
      setOpt(initialData.opt || 1);
      setReal(initialData.real || 2);
      setPess(initialData.pess || 4);
      setAssigneeId(initialData.assignee_id ?? '');

      // Загружаем логи для существующей задачи
      api.get(`/logs/${initialData.id}`)
        .then(res => setLogs(res || []))
        .catch(() => setLogs([]));
    } else {
      setTitle('');
      setOpt(1);
      setReal(2);
      setPess(4);
      setAssigneeId('');
      setLogs([]);
    }

    // Загружаем участников проекта для выбора исполнителя
    if (projectId) {
      api.get(`/projects/${projectId}/members`)
        .then(data => setMembers(data || []))
        .catch(() => setMembers([]));
    }
  }, [isOpen, initialData, projectId]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!title.trim()) return;

    setLoading(true);
    try {
      await onSave({
        title: title.trim(),
        opt: Number(opt),
        real: Number(real),
        pess: Number(pess),
        project_id: Number(projectId),
        assignee_id: assigneeId !== '' ? Number(assigneeId) : null,
      });
    } finally {
      setLoading(false);
    }
  };

  if (!isOpen) return null;

  const isEditing = !!initialData;

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-2xl w-full max-w-lg shadow-2xl overflow-hidden">

        {/* Шапка */}
        <div className="flex justify-between items-center px-6 py-4 border-b bg-white">
          <h3 className="font-bold text-gray-800 text-sm">
            {isEditing ? `Редактировать задачу ID-${initialData.id}` : 'Новая задача'}
          </h3>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 transition-colors p-1">
            <X size={18} />
          </button>
        </div>

        <div className="p-6 space-y-4 max-h-[80vh] overflow-y-auto">
          <form onSubmit={handleSubmit} className="space-y-4">

            {/* Название */}
            <div className="space-y-1.5">
              <label className="text-[10px] font-bold text-gray-400 uppercase tracking-widest">
                Название задачи *
              </label>
              <input
                className="w-full p-3 bg-gray-50 border rounded-xl focus:ring-2 focus:ring-blue-500 outline-none text-sm transition-all"
                placeholder="Что нужно сделать?"
                value={title}
                onChange={e => setTitle(e.target.value)}
                required
                autoFocus
              />
            </div>

            {/* Оценки времени */}
            <div className="space-y-1.5">
              <label className="text-[10px] font-bold text-gray-400 uppercase tracking-widest">
                Оценки времени (часы) — PERT
              </label>
              <div className="grid grid-cols-3 gap-3">
                <div>
                  <label className="text-[9px] text-green-600 font-bold uppercase block mb-1 text-center">
                    Оптимист.
                  </label>
                  <input
                    type="number" min="1"
                    className="w-full p-2.5 bg-gray-50 border rounded-xl focus:ring-2 focus:ring-green-400 outline-none text-sm text-center font-bold transition-all"
                    value={opt}
                    onChange={e => setOpt(e.target.value)}
                  />
                </div>
                <div>
                  <label className="text-[9px] text-blue-600 font-bold uppercase block mb-1 text-center">
                    Реальная
                  </label>
                  <input
                    type="number" min="1"
                    className="w-full p-2.5 bg-gray-50 border rounded-xl focus:ring-2 focus:ring-blue-400 outline-none text-sm text-center font-bold transition-all"
                    value={real}
                    onChange={e => setReal(e.target.value)}
                  />
                </div>
                <div>
                  <label className="text-[9px] text-red-500 font-bold uppercase block mb-1 text-center">
                    Пессимист.
                  </label>
                  <input
                    type="number" min="1"
                    className="w-full p-2.5 bg-gray-50 border rounded-xl focus:ring-2 focus:ring-red-400 outline-none text-sm text-center font-bold transition-all"
                    value={pess}
                    onChange={e => setPess(e.target.value)}
                  />
                </div>
              </div>
              <p className="text-[9px] text-gray-400 text-center">
                Формула: (O + 4·M + P) / 6
              </p>
            </div>

            {/* Исполнитель */}
            <div className="space-y-1.5">
              <label className="text-[10px] font-bold text-gray-400 uppercase tracking-widest">
                Исполнитель
              </label>
              <select
                className="w-full p-3 bg-gray-50 border rounded-xl focus:ring-2 focus:ring-blue-500 outline-none text-sm transition-all"
                value={assigneeId}
                onChange={e => setAssigneeId(e.target.value)}
              >
                <option value="">— Не назначен —</option>
                {members.map(m => (
                  <option key={m.user_id} value={m.user_id}>
                    {m.username} ({m.role})
                  </option>
                ))}
              </select>
            </div>

            <button
              type="submit"
              disabled={loading || !title.trim()}
              className="w-full bg-blue-600 text-white font-bold py-3 rounded-xl hover:bg-blue-700 active:scale-[0.98] transition-all disabled:opacity-50 disabled:cursor-not-allowed text-sm"
            >
              {loading
                ? 'Сохранение...'
                : isEditing
                  ? '✓ СОХРАНИТЬ ИЗМЕНЕНИЯ'
                  : '+ СОЗДАТЬ ЗАДАЧУ'}
            </button>
          </form>

          {/* Логи аудита (только при редактировании существующей задачи) */}
          {isEditing && (
            <div className="pt-4 border-t">
              <h4 className="text-[10px] font-bold text-gray-400 uppercase tracking-widest mb-3">
                История изменений
              </h4>
              <div className="bg-gray-50 p-3 rounded-xl max-h-36 overflow-y-auto">
                {logs.length > 0 ? (
                  <ul className="space-y-2 text-xs">
                    {logs.map((log, i) => (
                      <li key={i} className="border-b border-gray-100 pb-2 last:border-0">
                        <div className="flex justify-between items-center">
                          <span className="font-semibold text-blue-700">
                            {log.action === 'STATUS_CHANGED' ? 'Смена статуса' : log.action}
                          </span>
                          <span className="text-gray-400 text-[10px]">
                            {new Date(log.timestamp).toLocaleString('ru-RU')}
                          </span>
                        </div>
                        {log.new_status && (
                          <p className="text-[10px] text-gray-500 mt-0.5">
                            → <span className="font-bold text-gray-700 uppercase">{log.new_status}</span>
                          </p>
                        )}
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-gray-400 text-xs text-center py-2 italic">
                    История пуста. Измените статус задачи, чтобы появились логи.
                  </p>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

TaskModal.propTypes = {
  isOpen: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  onSave: PropTypes.func.isRequired,
  initialData: PropTypes.object,
  projectId: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
};
