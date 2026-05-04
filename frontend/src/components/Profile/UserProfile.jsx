import { useState, useEffect } from 'react';
import { api } from '../../api/client';
import { Save, User } from 'lucide-react';

export default function UserProfile() {
  const [profile, setProfile] = useState({
    username: '',
    full_name: 'Артем', // Можно добавить в БД
    stack: 'Go, PostgreSQL, Docker',
    status: 'Студент мехмата БГУ'
  });

  const loadProfile = async () => {
    try {
      const data = await api.get('/user/profile'); // Убедись, что эндпоинт есть в auth-service
      if (data) setProfile(prev => ({ ...prev, ...data }));
    } catch (err) { console.error(err); }
  };

  useEffect(() => { loadProfile(); }, []);

  const handleSave = async () => {
    try {
      await api.put('/user/profile', profile);
      alert('Данные успешно обновлены!');
    } catch (err) { alert('Ошибка: ' + err.message); }
  };

  return (
    <div className="h-full w-full bg-gray-50 p-6 overflow-hidden">
      <div className="w-full h-full max-w-4xl mx-auto flex flex-col bg-white rounded-2xl shadow-xl border overflow-hidden">
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
    </div>
  );
}