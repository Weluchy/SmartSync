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

  const loadProfile = async () => {
    try {
      const data = await api.get('/user/profile');
      if (data) setProfile(prev => ({ ...prev, ...data }));
    } catch (err) { console.error(err); }
  };

  const loadAuditLogs = async () => {
    try {
      const response = await api.get('/user/audit');
      // В response.data уже лежит готовый массив от axios
      console.log(response.data)
      setAuditLogs(response.data || []);
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
                      <th className="p-3">Время действия</th>
                      <th className="p-3">ID задачи</th>
                      <th className="p-3">Операция</th>
                      <th className="p-3">Новое состояние</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y text-gray-600">
                    {auditLogs.map((log, index) => (
                      <tr key={index} className="hover:bg-gray-50/80 transition-all">
                        <td className="p-3 whitespace-nowrap text-gray-400">
                          {new Date(log.created_at).toLocaleString()}
                        </td>
                        <td className="p-3 font-mono font-bold text-blue-600">#{log.task_id}</td>
                        <td className="p-3 font-semibold">
                          <span className="px-2 py-0.5 bg-amber-50 text-amber-700 border border-amber-200 rounded-md text-[10px]">
                            {log.action}
                          </span>
                        </td>
                        <td className="p-3">
                          <span className="px-2 py-0.5 bg-emerald-50 text-emerald-700 border border-emerald-200 rounded-md text-[10px] font-mono uppercase">
                            {log.new_status}
                          </span>
                        </td>
                      </tr>
                    ))}
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