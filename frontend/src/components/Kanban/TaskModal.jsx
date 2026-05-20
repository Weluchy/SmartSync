/* eslint-disable react/prop-types */
/* eslint-disable no-unused-vars */
import React, { useState, useEffect } from 'react';
import { api } from '../../api/client';
const TaskModal = ({ isOpen, onClose, task }) => {
  const [logs, setLogs] = useState([]);

  // Загружаем логи каждый раз, когда открывается окно задачи
  useEffect(() => {
    if (isOpen && task && task.id) {
      api.get(`/logs/${task.id}`)
        .then(res => setLogs(res.data || []))
        .catch(err => console.error("Ошибка загрузки логов:", err));
    }
  }, [isOpen, task]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 w-full max-w-md">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-xl font-bold">{task.title}</h3>
          <button onClick={onClose} className="text-gray-500 hover:text-gray-700">✕</button>
        </div>
        
        <div className="space-y-4">
          <div>
            <p className="text-sm text-gray-500">Описание</p>
            <p className="mt-1">{task.description || 'Нет описания'}</p>
          </div>
          
          <div className="flex justify-between">
            <div>
              <p className="text-sm text-gray-500">Статус</p>
              <span className="inline-block mt-1 px-2 py-1 text-xs font-semibold rounded-full bg-blue-100 text-blue-800">
                {task.status}
              </span>
            </div>
            <div>
              <p className="text-sm text-gray-500">Исполнитель</p>
              <p className="mt-1 text-sm font-medium">Не назначен</p>
            </div>
          </div>

          <div className="mt-6 pt-4 border-t">
            <h4 className="font-semibold mb-2">История (Аудит)</h4>
            <div className="bg-gray-50 p-3 rounded h-36 overflow-y-auto pr-2">
              {logs.length > 0 ? (
                <ul className="space-y-2 text-sm">
                  {logs.map((log, i) => (
                    <li key={i} className="border-b border-gray-200 pb-2">
                      <div className="flex justify-between text-blue-700 font-medium">
                        <span>{log.action === 'updated' ? 'Обновление статуса' : 'Создание'}</span>
                        <span className="text-xs text-gray-400 font-normal">
                          {new Date(log.timestamp).toLocaleString('ru-RU')}
                        </span>
                      </div>
                      {log.payload && log.payload.status && (
                        <p className="text-xs mt-1 text-gray-600">
                          Новый статус: <span className="font-semibold">{log.payload.status}</span>
                        </p>
                      )}
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="text-gray-500 text-sm italic text-center mt-4">История пуста</p>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default TaskModal;