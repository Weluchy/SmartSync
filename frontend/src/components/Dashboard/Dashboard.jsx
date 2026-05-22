import { useState, useEffect, useCallback } from 'react';
import PropTypes from 'prop-types';
import { api } from '../../api/client';
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer,
  PieChart, Pie, Cell
} from 'recharts';
import { CheckCircle, Clock, AlertTriangle, Layers, Target, Activity } from 'lucide-react';

const COLORS = ['#3b82f6', '#f59e0b', '#10b981', '#ef4444'];

export default function Dashboard({ projectId }) {
  const [stats, setStats] = useState([]); // Изначально пустой массив!
  const [milestones, setMilestones] = useState([]);
  const [loading, setLoading] = useState(true);

  const loadData = useCallback(async () => {
    if (!projectId) return;
    setLoading(true);
    try {
      // ИСПРАВЛЕНИЕ: Ждем массив задач. Если вернется null, ставим []
      const tasksData = await api.get(`/projects/${projectId}/tasks`);
      setStats(Array.isArray(tasksData) ? tasksData : []);

      // Если в API нет milestones, ловим ошибку и ставим []
      try {
        const milestonesData = await api.get(`/projects/${projectId}/milestones`);
        setMilestones(Array.isArray(milestonesData) ? milestonesData : []);
      } catch {
        setMilestones([]);
      }
    } catch (err) {
      console.error("Ошибка загрузки дашборда:", err);
      setStats([]); // Бронежилет от краша при ошибке 500
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full w-full bg-gray-50">
        <div className="flex flex-col items-center gap-3">
          <Activity className="animate-spin text-blue-600" size={32} />
          <p className="text-gray-500 font-medium">Сбор аналитики...</p>
        </div>
      </div>
    );
  }

  // БРОНЕЖИЛЕТ: Теперь stats всегда массив, ошибки null.length не будет
  const totalTasks = stats.length;
  const inProgressTasks = stats.filter(t => t.status === 'in_progress').length;
  const doneTasks = stats.filter(t => t.status === 'done').length;
  const backlogTasks = stats.filter(t => t.status === 'todo').length;

  const statusData = [
    { name: 'Бэклог', value: backlogTasks },
    { name: 'В работе', value: inProgressTasks },
    { name: 'Готово', value: doneTasks }
  ];

  const priorityData = [
    { name: 'Высокий (>80)', value: stats.filter(t => t.priority_score >= 80).length },
    { name: 'Средний (40-80)', value: stats.filter(t => t.priority_score >= 40 && t.priority_score < 80).length },
    { name: 'Низкий (<40)', value: stats.filter(t => t.priority_score < 40).length }
  ];

  return (
    <div className="h-full w-full bg-gray-50 p-6 overflow-y-auto">
      <div className="max-w-6xl mx-auto space-y-6">
        
        {/* Шапка */}
        <div className="flex justify-between items-end">
          <div>
            <h2 className="text-2xl font-bold text-gray-800 tracking-tight">Аналитика проекта</h2>
            <p className="text-sm text-gray-500 mt-1">Ключевые метрики и распределение нагрузки</p>
          </div>
          <button onClick={loadData} className="text-sm font-bold text-blue-600 bg-blue-50 px-4 py-2 rounded-xl hover:bg-blue-100 transition-colors">
            Обновить данные
          </button>
        </div>

        {/* KPI Карточки */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="bg-white p-5 rounded-2xl border shadow-sm flex items-center gap-4">
            <div className="w-12 h-12 rounded-full bg-blue-50 text-blue-600 flex items-center justify-center"><Layers size={24}/></div>
            <div>
              <p className="text-xs font-bold text-gray-400 uppercase">Всего задач</p>
              <h3 className="text-2xl font-black text-gray-800">{totalTasks}</h3>
            </div>
          </div>
          <div className="bg-white p-5 rounded-2xl border shadow-sm flex items-center gap-4">
            <div className="w-12 h-12 rounded-full bg-amber-50 text-amber-600 flex items-center justify-center"><Clock size={24}/></div>
            <div>
              <p className="text-xs font-bold text-gray-400 uppercase">В работе</p>
              <h3 className="text-2xl font-black text-gray-800">{inProgressTasks}</h3>
            </div>
          </div>
          <div className="bg-white p-5 rounded-2xl border shadow-sm flex items-center gap-4">
            <div className="w-12 h-12 rounded-full bg-green-50 text-green-600 flex items-center justify-center"><CheckCircle size={24}/></div>
            <div>
              <p className="text-xs font-bold text-gray-400 uppercase">Завершено</p>
              <h3 className="text-2xl font-black text-gray-800">{doneTasks}</h3>
            </div>
          </div>
          <div className="bg-white p-5 rounded-2xl border shadow-sm flex items-center gap-4">
            <div className="w-12 h-12 rounded-full bg-red-50 text-red-600 flex items-center justify-center"><AlertTriangle size={24}/></div>
            <div>
              <p className="text-xs font-bold text-gray-400 uppercase">Высокий приоритет</p>
              <h3 className="text-2xl font-black text-gray-800">{priorityData[0].value}</h3>
            </div>
          </div>
        </div>

        {/* Графики */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          
          <div className="bg-white p-6 rounded-2xl border shadow-sm">
            <h3 className="font-bold text-gray-800 mb-6">Задачи по статусам</h3>
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={statusData}>
                  <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f3f4f6" />
                  <XAxis dataKey="name" axisLine={false} tickLine={false} tick={{fontSize: 12, fill: '#9ca3af'}} />
                  <YAxis axisLine={false} tickLine={false} tick={{fontSize: 12, fill: '#9ca3af'}} />
                  <Tooltip cursor={{fill: '#f3f4f6'}} contentStyle={{borderRadius: '12px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)'}} />
                  <Bar dataKey="value" radius={[6, 6, 0, 0]}>
                    {statusData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>
          </div>

          <div className="bg-white p-6 rounded-2xl border shadow-sm">
            <h3 className="font-bold text-gray-800 mb-6">Распределение приоритетов (PERT)</h3>
            <div className="h-64 flex items-center justify-center">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie data={priorityData} cx="50%" cy="50%" innerRadius={60} outerRadius={80} paddingAngle={5} dataKey="value">
                    {priorityData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={COLORS[3 - index]} />
                    ))}
                  </Pie>
                  <Tooltip contentStyle={{borderRadius: '12px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)'}} />
                  <Legend iconType="circle" wrapperStyle={{fontSize: '12px'}} />
                </PieChart>
              </ResponsiveContainer>
            </div>
          </div>
        </div>

        {/* Milestones / Спринты */}
        <div className="bg-white rounded-2xl border shadow-sm overflow-hidden">
          <div className="p-6 border-b flex items-center gap-3">
            <Target className="text-blue-600" size={20} />
            <h3 className="font-bold text-gray-800">Вехи проекта (Milestones)</h3>
          </div>
          <div className="p-6">
            {milestones.length > 0 ? (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {milestones.map(m => (
                  <div key={m.id} className="border border-gray-100 rounded-xl p-4 hover:border-blue-200 hover:shadow-md transition-all">
                    <div className="flex justify-between items-start mb-2">
                      <h4 className="font-bold text-gray-800">{m.title}</h4>
                      <span className="text-[10px] font-bold px-2 py-1 bg-gray-100 text-gray-500 rounded-md">ID: {m.id}</span>
                    </div>
                    <p className="text-xs text-gray-500 mb-4 line-clamp-2">{m.description || 'Нет описания'}</p>
                    <div className="flex justify-between items-center text-xs font-medium">
                      <span className="text-blue-600">{new Date(m.due_date).toLocaleDateString('ru-RU')}</span>
                      <span className="text-gray-400">Статус: {m.status}</span>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8">
                <div className="w-16 h-16 bg-gray-50 rounded-full flex items-center justify-center mx-auto mb-3">
                  <Target className="text-gray-300" size={24} />
                </div>
                <p className="text-gray-500 font-medium">Спринты и Вехи пока не созданы</p>
                <p className="text-xs text-gray-400 mt-1">Добавьте milestone через API, чтобы отслеживать этапы</p>
              </div>
            )}
          </div>
        </div>

      </div>
    </div>
  );
}

Dashboard.propTypes = {
  projectId: PropTypes.oneOfType([PropTypes.number, PropTypes.string])
};