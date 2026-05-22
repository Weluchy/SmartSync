import { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { api } from '../../api/client';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, Legend } from 'recharts';

const COLORS = { todo: '#f59e0b', in_progress: '#3b82f6', done: '#22c55e' };

export default function Dashboard({ projectId }) {
  const [stats, setStats] = useState(null);
  const [milestones, setMilestones] = useState([]);

  useEffect(() => {
    if (!projectId) return;
    api.get(`/projects/${projectId}/stats`).then(setStats).catch(console.error);
    api.get(`/projects/${projectId}/milestones`).then(setMilestones).catch(console.error);
  }, [projectId]);

  if (!stats) return <div className="p-8 text-center" style={{ color: 'var(--text-muted)' }}>Загрузка статистики...</div>;

  const pieData = [
    { name: 'Бэклог', value: stats.todo, color: COLORS.todo },
    { name: 'В работе', value: stats.in_progress, color: COLORS.in_progress },
    { name: 'Готово', value: stats.done, color: COLORS.done },
  ].filter(d => d.value > 0);

  const completionRate = stats.total > 0 ? Math.round((stats.done / stats.total) * 100) : 0;

  return (
    <div className="h-full w-full p-6 overflow-y-auto" style={{ backgroundColor: 'var(--bg-page)' }}>
      <div className="max-w-5xl mx-auto space-y-6">

        {/* Карточки с метриками */}
        <div className="grid grid-cols-4 gap-4">
          {[
            { label: 'Всего задач', value: stats.total, color: '#6b7280' },
            { label: 'В работе', value: stats.in_progress, color: COLORS.in_progress },
            { label: 'Готово', value: stats.done, color: COLORS.done },
            { label: 'Выполнено', value: `${completionRate}%`, color: '#22c55e' },
          ].map((m, i) => (
            <div key={i} className="rounded-xl p-5 border" style={{ backgroundColor: 'var(--bg-card)', borderColor: 'var(--border)' }}>
              <p className="text-[11px] font-bold uppercase" style={{ color: 'var(--text-muted)' }}>{m.label}</p>
              <p className="text-3xl font-black mt-1" style={{ color: m.color }}>{m.value}</p>
            </div>
          ))}
        </div>

        {/* График: задачи по статусам */}
        <div className="grid grid-cols-2 gap-6">
          <div className="rounded-xl p-5 border" style={{ backgroundColor: 'var(--bg-card)', borderColor: 'var(--border)' }}>
            <h3 className="text-sm font-bold mb-4" style={{ color: 'var(--text-primary)' }}>Распределение задач</h3>
            <ResponsiveContainer width="100%" height={250}>
              <PieChart>
                <Pie data={pieData} dataKey="value" nameKey="name" cx="50%" cy="50%" outerRadius={80} label>
                  {pieData.map((e, i) => <Cell key={i} fill={e.color} />)}
                </Pie>
                <Tooltip />
                <Legend />
              </PieChart>
            </ResponsiveContainer>
          </div>

          <div className="rounded-xl p-5 border" style={{ backgroundColor: 'var(--bg-card)', borderColor: 'var(--border)' }}>
            <h3 className="text-sm font-bold mb-4" style={{ color: 'var(--text-primary)' }}>Часов по статусам</h3>
            <ResponsiveContainer width="100%" height={250}>
              <BarChart data={[
                { name: 'Бэклог', hours: stats.total_hours * (stats.todo / stats.total || 0) },
                { name: 'В работе', hours: stats.total_hours * (stats.in_progress / stats.total || 0) },
                { name: 'Готово', hours: stats.total_hours * (stats.done / stats.total || 0) },
              ]}>
                <XAxis dataKey="name" tick={{ fontSize: 10 }} />
                <YAxis tick={{ fontSize: 10 }} />
                <Tooltip />
                <Bar dataKey="hours" fill="#3b82f6" radius={[4,4,0,0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Средний приоритет */}
        <div className="rounded-xl p-5 border" style={{ backgroundColor: 'var(--bg-card)', borderColor: 'var(--border)' }}>
          <h3 className="text-sm font-bold mb-3" style={{ color: 'var(--text-primary)' }}>Средний приоритет задач</h3>
          <div className="w-full bg-gray-200 rounded-full h-4">
            <div className="h-4 rounded-full bg-gradient-to-r from-green-400 via-yellow-400 to-red-500" 
              style={{ width: `${Math.min((stats.avg_priority || 0) / 10 * 100, 100)}%` }} />
          </div>
          <p className="text-xs mt-2 font-bold" style={{ color: 'var(--text-secondary)' }}>
            {stats.avg_priority?.toFixed(1) || '0'} / 10
          </p>
        </div>

        {/* Вехи (Milestones) */}
        {milestones.length > 0 && (
          <div className="rounded-xl p-5 border" style={{ backgroundColor: 'var(--bg-card)', borderColor: 'var(--border)' }}>
            <h3 className="text-sm font-bold mb-3" style={{ color: 'var(--text-primary)' }}>Вехи проекта</h3>
            <div className="space-y-3">
              {milestones.map(m => (
                <div key={m.id} className="flex items-center gap-3 p-3 rounded-lg" style={{ backgroundColor: 'var(--bg-card-hover)' }}>
                  <div className="w-2 h-2 rounded-full bg-blue-500" />
                  <div className="flex-1">
                    <p className="text-sm font-bold" style={{ color: 'var(--text-primary)' }}>{m.title}</p>
                    <p className="text-[10px]" style={{ color: 'var(--text-muted)' }}>
                      Дедлайн: {new Date(m.deadline).toLocaleString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' })}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

Dashboard.propTypes = { projectId: PropTypes.oneOfType([PropTypes.number, PropTypes.string]) };