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
  const [userId, setUserId] = useState(null);
  const [auditLogs, setAuditLogs] = useState([]);
  const [userNames, setUserNames] = useState({});
  
  const loadProfile = async () => {
    try {
      const data = await api.get('/user/profile');
      if (data) {
        setProfile(prev => ({ ...prev, ...data }));
        // ФИКС (Пункт 6): Сохраняем ID из профиля, чтобы он не терялся
        if (data.id) {
          localStorage.setItem('userId', data.id);
          setUserId(data.id);
        }
      }
    } catch (err) { console.error(err); }
  };

  const loadAuditLogs = async () => {
    try {
        const logs = await api.get('/user/audit', {
            headers: { 'X-User-ID': localStorage.getItem('userId') }
        });
        const fetchedLogs = logs || [];
        setAuditLogs(fetchedLogs);

        // ФИКС (Пункт 2): Безопасный сбор имен без bulk, который блокировался
        const userIds = [...new Set(fetchedLogs.map(l => l.user_id).filter(Boolean))];
        const namesMap = {};
        for (const id of userIds) {
          try {
            const u = await api.get(`/users/${id}`);
            if (u && u.username) namesMap[id] = u.username;
          } catch (err) {

      console.error("Ошибка загрузки истории аудита:", err);

    }
        }
        setUserNames(namesMap);
    } catch (err) { console.error(err); }
  };

  useEffect(() => {
    setUserId(localStorage.getItem('userId'));
    loadProfile(); 
    loadAuditLogs();
  }, []);

  const handleSave = async () => {
    try {
      await api.put('/user/profile', profile);
      alert('Данные успешно обновлены!');
    } catch (err) { alert('Ошибка: ' + err.message); }
  };

  // ФИКС (Пункт 3): Убрали bg-white, используем переменные темы (var(--bg-card))
  return (
    <div className="h-full w-full p-6 overflow-y-auto" style={{ backgroundColor: 'var(--bg-page)', color: 'var(--text-primary)' }}>
      <div className="w-full max-w-4xl mx-auto space-y-6">
        
        <div className="flex flex-col rounded-2xl shadow-xl border overflow-hidden" style={{ backgroundColor: 'var(--bg-card)', borderColor: 'var(--border)' }}>
          <div className="h-[72px] px-8 border-b flex items-center justify-between" style={{ borderColor: 'var(--border)' }}>
            <h3 className="font-bold uppercase text-xs tracking-widest" style={{ color: 'var(--text-secondary)' }}>Настройки профиля</h3>
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
                <input 
                  className="text-xl font-bold bg-transparent border-b border-dashed outline-none mb-1 pb-1"
                  style={{ color: 'var(--text-primary)', borderColor: 'var(--border-hover)' }}
                  value={profile.username}
                  placeholder="Ваше имя"
                  onChange={e => setProfile({...profile, username: e.target.value})}
                />
                <p className="text-sm" style={{ color: 'var(--text-muted)' }}>ID пользователя: {userId || '—'}</p>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-6">
              <div className="space-y-2">
                <label className="text-[10px] font-bold uppercase" style={{ color: 'var(--text-muted)' }}>Стек технологий</label>
                <input 
                  className="w-full p-3 border rounded-xl outline-none transition-all"
                  style={{ backgroundColor: 'var(--bg-input)', color: 'var(--text-primary)', borderColor: 'var(--border)' }}
                  value={profile.stack}
                  onChange={e => setProfile({...profile, stack: e.target.value})}
                />
              </div>
              <div className="space-y-2">
                <label className="text-[10px] font-bold uppercase" style={{ color: 'var(--text-muted)' }}>Статус / Вуз</label>
                <input 
                  className="w-full p-3 border rounded-xl outline-none transition-all"
                  style={{ backgroundColor: 'var(--bg-input)', color: 'var(--text-primary)', borderColor: 'var(--border)' }}
                  value={profile.status}
                  onChange={e => setProfile({...profile, status: e.target.value})}
                />
              </div>
            </div>
          </div>
        </div>

        <div className="rounded-2xl shadow-xl border overflow-hidden" style={{ backgroundColor: 'var(--bg-card)', borderColor: 'var(--border)' }}>
          <div className="h-[72px] px-8 border-b flex items-center gap-2" style={{ borderColor: 'var(--border)' }}>
            <History size={16} style={{ color: 'var(--text-muted)' }} />
            <h3 className="font-bold uppercase text-xs tracking-widest" style={{ color: 'var(--text-secondary)' }}>История ваших действий</h3>
          </div>
          <div className="p-6">
            {auditLogs.length === 0 ? (
              <p className="text-xs text-center py-4" style={{ color: 'var(--text-muted)' }}>История действий пуста. Переместите задачи на доске.</p>
            ) : (
              <div className="overflow-x-auto rounded-xl border" style={{ borderColor: 'var(--border)' }}>
                <table className="w-full text-left text-xs border-collapse">
                  <thead>
                    <tr className="font-bold uppercase border-b text-[10px]" style={{ backgroundColor: 'var(--bg-input)', color: 'var(--text-muted)', borderColor: 'var(--border)' }}>
                      <th className="p-3">Время</th>
                      <th className="p-3">ID задачи</th>
                      <th className="p-3">Кто</th>
                      <th className="p-3">Действие</th>
                      <th className="p-3">Что изменилось</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y" style={{ color: 'var(--text-secondary)', borderColor: 'var(--border)' }}>
                    {auditLogs.map((log, index) => {
                      const time = new Date(log.timestamp);
                      const timeStr = time.toLocaleString('ru-RU', { hour: '2-digit', minute: '2-digit', day: 'numeric', month: 'numeric', year: 'numeric' });
                      const actionLabel = log.action === 'updated' ? 'Обновление' : log.action === 'created' ? 'Создание' : log.action === 'status_changed' ? 'Смена статуса' : log.action;
                      const userName = userNames[log.user_id] || `ID ${log.user_id}`;

                      return (
                      <tr key={index} className="transition-all" style={{ backgroundColor: 'var(--bg-card)' }}>
                        <td className="p-3 whitespace-nowrap text-[11px]" style={{ color: 'var(--text-muted)' }}>{timeStr}</td>
                        <td className="p-3 font-mono font-bold text-blue-600">#{log.task_id}</td>
                        <td className="p-3 font-medium" style={{ color: 'var(--text-primary)' }}>{userName}</td>
                        <td className="p-3 font-semibold">
                          <span className={`px-2 py-0.5 rounded-md text-[10px] border ${log.action === 'created' ? 'bg-green-500/10 text-green-500 border-green-500/20' : 'bg-blue-500/10 text-blue-500 border-blue-500/20'}`}>
                            {actionLabel}
                          </span>
                        </td>
                        <td className="p-3 max-w-xs">
                          {log.summary ? <span className="text-[11px] leading-tight block">{log.summary}</span> : <span className="italic text-[11px]" style={{ color: 'var(--text-muted)' }}>—</span>}
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