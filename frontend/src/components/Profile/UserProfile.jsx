import { useState, useEffect } from 'react';
import { api } from '../../api/client';
import { Save, User, History } from 'lucide-react';

export default function UserProfile() {
  const [profile, setProfile] = useState({
    username: '',
    full_name: 'Артем',
    stack: 'Go, PostgreSQL, Docker',
    status: 'Студент мехмата БГУ'
  });

  const [auditLogs, setAuditLogs] = useState([]);
  const [userNames, setUserNames] = useState({});

  const loadProfile = async () => {
    try {
      const data = await api.get('/user/profile');
      if (data) setProfile(prev => ({ ...prev, ...data }));
    } catch (err) { console.error(err); }
  };

  const loadAuditLogs = async () => {
    try {
      const logs = await api.get('/user/audit');
      setAuditLogs(logs || []);

      // Собираем все user_id из логов, чтобы получить имена
      const userIds = [...new Set((logs || []).map(l => l.user_id).filter(Boolean))];
      if (userIds.length > 0) {
        const namesMap = await api.post('/internal/users/bulk', { ids: userIds });
        if (namesMap) setUserNames(namesMap);
      }
    } catch (err) {
      console.error("Ошибка загрузки истории аудита:", err);
    }
  };

  useEffect(() => { 
    loadProfile(); 
    loadAuditLogs();
  }, []);

  const handleSave = async () => {
    try {
      await api.put('/user/profile', profile);
      alert('Данные успешно обновлены!');
    } catch (err) { alert('Ошибка: ' + err.message); }
  };

  return (
    <div className="h-full w-full bg-gray-50 p-6 overflow-y-auto">
      <div className="w-full max-w-4xl mx-auto space-y-6">
        
        {/* Карточка профиля */}
        <div className="flex flex-col bg-white rounded-2xl shadow-xl border overflow-hidden">
          <div className="h-[72px] px-8 border-b flex items-center justify-between bg-white">
            <h3 className="font-bold text-gray-700 uppercase text-xs tracking-widest">Настройки профиля</h3>
            <button onClick={handleSave} className="bg-blue-600 text-white px-4 py-2 rounded-lg text-xs font-bold flex items-center gap-2 hover:bg-blue-700 transition-all">
              <Save size={14} /> СОХРАНИТЬ
            </button>
          </div>
          
          <div className="p-8 space-y-6">
            <div className="flex items-center gap-6 mb-4">
              <div className="w-20 h-20 bg-blue-100 text-blue-600 rounded-full flex items-center justify-center shadow-inner">
                <User size={40} />
              </div>
              <div>
                <h2 className="text-xl font-bold text-gray-800">{profile.username}</h2>
                <p className="text-sm text-gray-400">ID пользователя: {localStorage.getItem('userId') || '—'}</p>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-6">
              <div className="space-y-2">
                <label className="text-[10px] font-bold text-gray-400 uppercase">Стек технологий</label>
                <input 
                  className="w-full p-3 bg-gray-50 border rounded-xl focus:ring-2 focus:ring-blue-500 outline-none transition-all"
                  value={profile.stack}
                  onChange={e => setProfile({...profile, stack: e.target.value})}
                />
              </div>
              <div className="space-y-2">
                <label className="text-[10px] font-bold text-gray-400 uppercase">Статус / Вуз</label>
                <input 
                  className="w-full p-3 bg-gray-50 border rounded-xl focus:ring-2 focus:ring-blue-500 outline-none transition-all"
                  value={profile.status}
                  onChange={e => setProfile({...profile, status: e.target.value})}
                />
              </div>
            </div>
          </div>
        </div>

        {/* Таблица аудита действий (Для исследования в дипломе) */}
        <div className="bg-white rounded-2xl shadow-xl border overflow-hidden">
          <div className="h-[72px] px-8 border-b flex items-center gap-2 bg-white">
            <History size={16} className="text-gray-500" />
            <h3 className="font-bold text-gray-700 uppercase text-xs tracking-widest">История ваших действий (NoSQL Audit Stream)</h3>
          </div>
          <div className="p-6">
            {auditLogs.length === 0 ? (
              <p className="text-gray-400 text-xs text-center py-4">История действий пуста. Переместите задачи на доске, чтобы логи зафиксировались.</p>
            ) : (
              <div className="overflow-x-auto rounded-xl border border-gray-100">
                <table className="w-full text-left text-xs border-collapse">
                  <thead>
                    <tr className="bg-gray-50 text-gray-400 font-bold uppercase border-b text-[10px]">
                      <th className="p-3">Время</th>
                      <th className="p-3">ID задачи</th>
                      <th className="p-3">Кто</th>
                      <th className="p-3">Действие</th>
                      <th className="p-3">Что изменилось</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y text-gray-600">
                    {auditLogs.map((log, index) => {
                      const time = new Date(log.timestamp);
                      const timeStr = time.toLocaleString('ru-RU', {
                        hour: '2-digit',
                        minute: '2-digit',
                        day: 'numeric',
                        month: 'numeric',
                        year: 'numeric'
                      });

                      const actionLabel = log.action === 'updated' ? 'Обновление' 
                        : log.action === 'created' ? 'Создание' 
                        : log.action === 'status_changed' ? 'Смена статуса' 
                        : log.action;

                      const userName = userNames[String(log.user_id)] || `ID ${log.user_id}`;

                      return (
                      <tr key={index} className="hover:bg-gray-50/80 transition-all">
                        <td className="p-3 whitespace-nowrap text-gray-400 text-[11px]">{timeStr}</td>
                        <td className="p-3 font-mono font-bold text-blue-600">#{log.task_id}</td>
                        <td className="p-3 text-gray-700 font-medium">{userName}</td>
                        <td className="p-3 font-semibold">
                          <span className={`px-2 py-0.5 rounded-md text-[10px] border ${
                            log.action === 'created' ? 'bg-green-50 text-green-700 border-green-200' 
                            : log.action === 'updated' ? 'bg-amber-50 text-amber-700 border-amber-200' 
                            : 'bg-blue-50 text-blue-700 border-blue-200'
                          }`}>
                            {actionLabel}
                          </span>
                        </td>
                        <td className="p-3 text-gray-600 max-w-xs">
                          {log.summary ? (
                            <span className="text-[11px] leading-tight block">{log.summary}</span>
                          ) : log.changes && log.changes.length > 0 ? (
                            <ul className="list-disc list-inside text-[11px] space-y-0.5">
                              {log.changes.map((ch, ci) => <li key={ci}>{ch}</li>)}
                            </ul>
                          ) : (
                            <span className="text-gray-400 italic text-[11px]">—</span>
                          )}
                        </td>
                      </tr>
                    )})}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>

      </div>
    </div>
  );
}