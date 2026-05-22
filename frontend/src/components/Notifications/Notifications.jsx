import { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { Bell, BellRing, X } from 'lucide-react';

export default function Notifications({ tasks }) {
  const [notifications, setNotifications] = useState([]);
  const [isOpen, setIsOpen] = useState(false);
  const userId = Number(localStorage.getItem('userId'));

  useEffect(() => {
    if (!tasks || !userId) return;

    const now = Date.now();
    const newNotifs = [];

    tasks.forEach(task => {
      // 1. Назначили задачу
      if (Number(task.assignee_id) === userId) {
        const stored = localStorage.getItem(`notif_assigned_${task.id}`);
        if (!stored) {
          newNotifs.push({
            id: `assigned_${task.id}`,
            type: 'assigned',
            text: `Вам назначена задача «${task.title}»`,
            taskId: task.id,
          });
        }
      }

      // 2. Дедлайн скоро (меньше 4 часов до deadline_at)
      if (task.deadline_at > 0 && Number(task.assignee_id) === userId) {
        const timeLeft = task.deadline_at - now;
        if (timeLeft > 0 && timeLeft < 4 * 3600000 && task.status !== 'done') {
          const stored = localStorage.getItem(`notif_deadline_${task.id}`);
          const hoursLeft = Math.ceil(timeLeft / 3600000);
          if (!stored) {
            newNotifs.push({
              id: `deadline_${task.id}`,
              type: 'deadline',
              text: `Задача «${task.title}» — осталось ${hoursLeft} ч.`,
              taskId: task.id,
            });
          }
        }
      }

      // 3. Статус изменился (кто-то передвинул)
      const storedStatus = localStorage.getItem(`notif_status_${task.id}`);
      if (storedStatus && storedStatus !== task.status && Number(task.assignee_id) === userId) {
        newNotifs.push({
          id: `status_${task.id}_${Date.now()}`,
          type: 'status',
          text: `Задача «${task.title}» → ${task.status === 'in_progress' ? 'В работе' : task.status === 'done' ? 'Готово' : task.status}`,
          taskId: task.id,
        });
      }
      localStorage.setItem(`notif_status_${task.id}`, task.status);
    });

    // Сохраняем новые уведомления
    if (newNotifs.length > 0) {
      setNotifications(prev => {
        const merged = [...newNotifs, ...prev].slice(0, 50);
        localStorage.setItem('notifications', JSON.stringify(merged));
        return merged;
      });
    }
  }, [tasks, userId]);

  // При монтировании загружаем сохранённые
  useEffect(() => {
    const saved = JSON.parse(localStorage.getItem('notifications') || '[]');
    setNotifications(saved);
  }, []);

  const dismissNotification = (id) => {
    const updated = notifications.filter(n => n.id !== id);
    setNotifications(updated);
    localStorage.setItem('notifications', JSON.stringify(updated));
    localStorage.setItem(`notif_${id}`, 'seen');
  };

  const unreadCount = notifications.length;

  return (
    <div className="relative">
      <button onClick={() => setIsOpen(!isOpen)} className="relative p-2 rounded-lg transition-colors" style={{ color: 'var(--text-muted)' }}>
        {unreadCount > 0 ? <BellRing size={18} className="text-yellow-500 animate-pulse" /> : <Bell size={18} />}
        {unreadCount > 0 && (
          <span className="absolute -top-0.5 -right-0.5 bg-red-500 text-white text-[9px] font-bold w-4 h-4 rounded-full flex items-center justify-center">
            {unreadCount > 9 ? '9+' : unreadCount}
          </span>
        )}
      </button>

      {isOpen && (
        <div className="absolute right-0 top-10 w-80 bg-white rounded-xl shadow-2xl border z-50 max-h-96 overflow-y-auto">
          <div className="p-3 border-b flex justify-between items-center">
            <span className="text-xs font-bold text-gray-700 uppercase">Уведомления</span>
            <button onClick={() => { setNotifications([]); localStorage.setItem('notifications', '[]'); }} className="text-[10px] text-blue-600 hover:underline">Все прочитаны</button>
          </div>
          {notifications.length === 0 ? (
            <p className="text-xs text-gray-400 italic text-center py-6">Нет новых уведомлений</p>
          ) : (
            notifications.map(n => (
              <div key={n.id} className="p-3 border-b last:border-0 hover:bg-gray-50 flex justify-between items-start gap-2">
                <div className="flex-1 min-w-0">
                  <p className="text-xs text-gray-700">{n.text}</p>
                  <p className="text-[9px] text-gray-400 mt-0.5">
                    {n.type === 'assigned' ? '📌 Назначение' : n.type === 'deadline' ? '⏰ Дедлайн' : '🔄 Изменение'}
                  </p>
                </div>
                <button onClick={() => dismissNotification(n.id)} className="shrink-0 text-gray-400 hover:text-gray-600"><X size={14} /></button>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
}

Notifications.propTypes = {
  tasks: PropTypes.array,
};
