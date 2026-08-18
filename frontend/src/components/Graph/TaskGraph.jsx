import { useEffect, useRef, useCallback, useState } from 'react';
import PropTypes from 'prop-types';
import { Network } from 'vis-network/standalone';
import { api } from '../../api/client';
import { Search, Filter, Download, RotateCcw, RefreshCw } from 'lucide-react';
import { toast } from 'react-hot-toast';

const POSITIONS_KEY_PREFIX = 'smartsync_graph_positions_';

export default function TaskGraph({ projectId }) {
  const containerRef = useRef(null);
  const networkRef = useRef(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [filterStatus, setFilterStatus] = useState('all');
  const [filterAssignee, setFilterAssignee] = useState('all');

  const exportPNG = () => {
    if (!networkRef.current) return;
    const canvas = containerRef.current.querySelector('canvas');
    if (!canvas) return;
    const link = document.createElement('a');
    link.download = `graph_project_${projectId}.png`;
    link.href = canvas.toDataURL('image/png');
    link.click();
    toast.success('Изображение графа сохранено');
  };

  const wouldCreateCycle = (fromId, toId, edges) => {
    const adj = {};
    edges.forEach(e => {
      if (!adj[e.from]) adj[e.from] = [];
      adj[e.from].push(e.to);
    });
    if (!adj[fromId]) adj[fromId] = [];
    adj[fromId].push(toId);
    const visited = new Set();
    const stack = new Set();
    const hasCycle = (node) => {
      visited.add(node);
      stack.add(node);
      const neighbors = adj[node] || [];
      for (let neighbor of neighbors) {
        if (!visited.has(neighbor)) {
          if (hasCycle(neighbor)) return true;
        } else if (stack.has(neighbor)) return true;
      }
      stack.delete(node);
      return false;
    };
    for (let node in adj) {
      if (!visited.has(Number(node))) {
        if (hasCycle(Number(node))) return true;
      }
    }
    return false;
  };

  const savePositions = useCallback(() => {
    if (!networkRef.current || !projectId) return;
    const positions = networkRef.current.getPositions();
    try {
      localStorage.setItem(POSITIONS_KEY_PREFIX + projectId, JSON.stringify(positions));
    } catch { /* ignore */ }
  }, [projectId]);

  const loadSavedPositions = useCallback(() => {
    if (!projectId) return null;
    try {
      const saved = localStorage.getItem(POSITIONS_KEY_PREFIX + projectId);
      return saved ? JSON.parse(saved) : null;
    } catch { return null; }
  }, [projectId]);

  const loadGraphData = useCallback(async () => {
    if (!projectId || !containerRef.current) return;
    try {
      // Добавляем уникальный параметр, чтобы браузер не кэшировал ответ
      const t = new Date().getTime();
      const data = await api.get(`/projects/${projectId}/graph?_t=${t}`);
      const msData = await api.get(`/projects/${projectId}/milestones?_t=${t}`).catch(() => []); 

      let tasks = data?.tasks || data?.nodes || []; 
      const dependencies = data?.dependencies || data?.edges || [];

      // Фильтры
      if (searchQuery) {
        const q = searchQuery.toLowerCase();
        tasks = tasks.filter(t =>
          t.title?.toLowerCase().includes(q) ||
          String(t.id).includes(q) ||
          t.description?.toLowerCase().includes(q) ||
          t.assignee_name?.toLowerCase().includes(q)
        );
      }
      if (filterStatus !== 'all') {
        tasks = tasks.filter(t => t.status === filterStatus);
      }
      if (filterAssignee === 'unassigned') {
        tasks = tasks.filter(t => !t.assignee_id);
      } else if (filterAssignee === 'assigned') {
        tasks = tasks.filter(t => !!t.assignee_id);
      }

      const durations = {};
      tasks.forEach(t => {
        durations[t.id] = t.status === 'done' ? 0 : ((t.opt + 4 * t.real + t.pess) / 6);
      });

      const maxScore = tasks.length > 0 ? Math.max(...tasks.map(t => t.priority_score || 0)) : 0;
      const criticalSet = new Set();

      tasks.forEach(t => {
        const score = t.priority_score || 0;
        if (score >= maxScore - 0.1 && maxScore > 0) criticalSet.add(t.id);
      });

      let changed = true;
      while(changed) {
        changed = false;
        dependencies.forEach(d => {
          if (criticalSet.has(d.task_id) && !criticalSet.has(d.depends_on_id)) {
            const parentScore = tasks.find(t => t.id === d.depends_on_id)?.priority_score || 0;
            const childDur = durations[d.task_id] || 0;
            const childScore = tasks.find(t => t.id === d.task_id)?.priority_score || 0;
            if (Math.abs(parentScore + childDur - childScore) < 0.1) {
              criticalSet.add(d.depends_on_id);
              changed = true;
            }
          }
        });
      }

      const savedPositions = loadSavedPositions();
      
      const nodes = tasks.map(t => {
        const score = t.priority_score || 0;
        const isDone = t.status === 'done';
        const isInProgress = t.status === 'in_progress';
        const isCritical = criticalSet.has(t.id) && !isDone; 

        // Цветовые схемы
        let bg = '#ffffff', border = '#e5e7eb', text = '#1f2937'; 
        if (isDone) { bg = '#f0fdf4'; border = '#22c55e'; text = '#166534'; }
        else if (isCritical) { bg = '#fef2f2'; border = '#ef4444'; text = '#991b1b'; }
        else if (isInProgress) { bg = '#eff6ff'; border = '#3b82f6'; text = '#1e3a8a'; }

        const duration = t.duration_hours || ((t.opt + 4 * t.real + t.pess) / 6);
        
        // Поиск вехи
        const taskMilestone = msData.find(m => m.id === t.milestone_id);
        const milestoneLabel = taskMilestone ? `   |   🎯 ${taskMilestone.title}` : '';

        const shortTitle = t.title.length > 22 ? t.title.substring(0, 22) + '...' : t.title;
        const assignee = t.assignee_name ? t.assignee_name.split(' ')[0] : 'Нет исп.';
        
        const labelStr = `*ID-${t.id}* ${shortTitle}\n👤 ${assignee}   |   ⏳ ${duration.toFixed(1)}ч   |   🔥 ${score.toFixed(1)}${milestoneLabel}`;

        const statusMap = {
          'todo': { text: 'Бэклог', bg: '#f3f4f6', color: '#4b5563' },
          'in_progress': { text: 'В работе', bg: '#dbeafe', color: '#1d4ed8' },
          'done': { text: 'Готово', bg: '#dcfce7', color: '#15803d' }
        };
        const s = statusMap[t.status] || statusMap['todo'];
        
        let deadlineHtml = '';
        if (t.deadline_at && t.deadline_at > 0) {
          const dDate = new Date(t.deadline_at).toLocaleString('ru-RU', {day:'2-digit', month:'2-digit', hour:'2-digit', minute:'2-digit'});
          deadlineHtml = `<div style="font-size: 11px; margin-bottom: 4px; color: #ef4444;">⏰ <b>Дедлайн:</b> ${dDate}</div>`;
        }

        const milestoneHtml = taskMilestone ? `<div style="font-size: 11px; margin-bottom: 4px; color: #7e22ce;">🎯 <b>Веха (Спринт):</b> ${taskMilestone.title}</div>` : '';
        const descStr = t.description ? (t.description.length > 100 ? t.description.substring(0, 100) + '...' : t.description) : 'Описание отсутствует';

        const tooltipContainer = document.createElement('div');
        tooltipContainer.innerHTML = `
          <div style="font-family: system-ui, sans-serif; padding: 6px 8px; max-width: 280px;">
              <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
                  <span style="color: #6b7280; font-size: 10px; font-weight: bold; text-transform: uppercase;">Задача #${t.id}</span>
                  <span style="background-color: ${s.bg}; color: ${s.color}; font-size: 10px; font-weight: bold; padding: 2px 6px; border-radius: 4px;">${s.text}</span>
              </div>
              <strong style="display: block; font-size: 13px; color: #111827; margin-bottom: 8px; line-height: 1.4;">${t.title}</strong>
              <div style="font-size: 11px; color: #4b5563; margin-bottom: 4px;">👤 <b>Исполнитель:</b> ${t.assignee_name || 'Не назначен'}</div>
              <div style="font-size: 11px; color: #4b5563; margin-bottom: 4px;">⏳ <b>Оценка (О/Р/П):</b> ${t.opt} / ${t.real} / ${t.pess} (Итог: ${duration.toFixed(1)}ч)</div>
              <div style="font-size: 11px; color: #ef4444; margin-bottom: 4px; font-weight: bold;">🔥 <b>PERT вес:</b> ${score.toFixed(2)}</div>
              ${deadlineHtml}
              ${milestoneHtml}
              <div style="margin-top: 8px; padding-top: 8px; border-top: 1px solid #e5e7eb; font-size: 11px; color: #6b7280; font-style: italic; line-height: 1.4;">
                  ${descStr}
              </div>
          </div>
        `;

        const node = {
          id: t.id,
          label: labelStr,
          title: tooltipContainer, 
          color: { 
            background: bg, border: border,
            highlight: { background: bg, border: '#3b82f6' },
            hover: { background: bg, border: '#3b82f6' }
          },
          font: { multi: 'markdown', size: 12, color: text, face: 'system-ui' },
          shape: 'box',
          margin: { top: 12, bottom: 12, left: 16, right: 16 },
          borderWidth: isCritical ? 3 : 1,
          shadow: isCritical ? { enabled: true, color: 'rgba(239, 68, 68, 0.25)', size: 15, x: 0, y: 4 } : { enabled: true, color: 'rgba(0,0,0,0.06)', size: 6, x: 0, y: 3 }
        };

        if (!networkRef.current && savedPositions && savedPositions[t.id]) {
          node.x = savedPositions[t.id].x;
          node.y = savedPositions[t.id].y;
        }
        return node;
      });

      const edges = dependencies.map(d => ({
        id: `${d.depends_on_id}-${d.task_id}`,
        from: d.depends_on_id, 
        to: d.task_id,
        arrows: 'to',
        color: { color: '#cbd5e1', highlight: '#3b82f6' },
        smooth: { type: 'cubicBezier', forceDirection: 'horizontal' }
      }));

      if (!networkRef.current) {
        networkRef.current = new Network(containerRef.current, { nodes, edges }, {
          physics: {
            enabled: !savedPositions,
            solver: 'forceAtlas2Based',
            forceAtlas2Based: { gravitationalConstant: -40, centralGravity: 0.005, springLength: 200, springConstant: 0.02, damping: 0.4 },
            stabilization: { iterations: 100, updateInterval: 25 }
          },
          layout: { hierarchical: { enabled: false } },
          interaction: { multiselect: true, hover: true, dragNodes: true },
          nodes: { physics: true }
        });

        if (!savedPositions) {
          networkRef.current.once('stabilizationIterationsDone', () => {
            networkRef.current.setOptions({ physics: { enabled: false } });
            savePositions();
          });
        }

        networkRef.current.on('dragEnd', savePositions);
        
        // Обработчик создания связи
        networkRef.current.on("click", (params) => {
          if (params.nodes.length === 2) {
            const [from, to] = params.nodes;
            const currentEdges = networkRef.current.body.data.edges.get();
            if (wouldCreateCycle(from, to, currentEdges)) {
              toast.error("Ошибка: Эта связь приведет к бесконечному циклу!");
              networkRef.current.unselectAll();
              return;
            }
            api.post(`/tasks/${to}/dependencies`, { depends_on_id: from })
               .then(() => { 
                 toast.success('Связь создана');
                 networkRef.current.unselectAll(); 
                 setTimeout(loadGraphData, 600); 
               })
               .catch(err => { toast.error(err.message); networkRef.current.unselectAll(); });
          }
        });

        // Обработчик удаления связи
        networkRef.current.on("doubleClick", (params) => {
          if (params.edges.length > 0) {
            const edgeId = params.edges[0];
            if (typeof edgeId === 'string' && edgeId.includes('-')) {
              const [from, to] = edgeId.split('-');
              
              toast((t) => (
                <div className="flex flex-col gap-2 p-1">
                  <span className="text-sm font-bold text-gray-800">Удалить зависимость?</span>
                  <div className="flex gap-2 mt-1">
                    <button onClick={() => {
                      toast.dismiss(t.id);
                      api.delete(`/tasks/${to}/dependencies/${from}`)
                         .then(() => { 
                           toast.success('Связь удалена');
                           networkRef.current.unselectAll(); 
                           setTimeout(loadGraphData, 600); 
                         }).catch(err => toast.error(err.message));
                    }} className="flex-1 bg-red-500 text-white py-1.5 rounded-lg text-xs font-bold hover:bg-red-600">Удалить</button>
                    <button onClick={() => toast.dismiss(t.id)} className="flex-1 bg-gray-100 text-gray-600 py-1.5 rounded-lg text-xs font-bold hover:bg-gray-200">Отмена</button>
                  </div>
                </div>
              ), { duration: 5000 });
            }
          }
        });
      } else {
        // Обновляем данные без сброса координат
        const nodesDataSet = networkRef.current.body.data.nodes;
        const edgesDataSet = networkRef.current.body.data.edges;
        
        const existingNodeIds = nodesDataSet.getIds();
        const existingEdgeIds = edgesDataSet.getIds();
        
        const newNodesIds = nodes.map(n => n.id);
        const newEdgesIds = edges.map(e => e.id);
        
        const nodesToRemove = existingNodeIds.filter(id => !newNodesIds.includes(id));
        const edgesToRemove = existingEdgeIds.filter(id => !newEdgesIds.includes(id));
        if (nodesToRemove.length > 0) nodesDataSet.remove(nodesToRemove);
        if (edgesToRemove.length > 0) edgesDataSet.remove(edgesToRemove);
        
        const hasNewNodes = nodes.some(n => !existingNodeIds.includes(n.id));
        
        nodesDataSet.update(nodes);
        edgesDataSet.update(edges);
        
        if (hasNewNodes) {
          networkRef.current.setOptions({ physics: { enabled: true } });
          setTimeout(() => {
            if (networkRef.current) {
              networkRef.current.setOptions({ physics: { enabled: false } });
              savePositions();
            }
          }, 1500); 
        }
      }
    } catch (err) { console.error("Graph load error:", err); }
  }, [projectId, loadSavedPositions, savePositions, searchQuery, filterStatus, filterAssignee]);

  // Граф обновляется автоматически через WebSocket после пересчёта
  useEffect(() => {
    loadGraphData();
    const ws = new WebSocket('ws://localhost:8000/ws');
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.project_id === Number(projectId)) {
          setTimeout(loadGraphData, 600);
        }
      } catch(err) { console.error("WS parse error:", err); }
    };
    return () => { 
      ws.close(); 
      if (networkRef.current) {
        networkRef.current.destroy(); 
        networkRef.current = null;
      }
    };
  }, [projectId, loadGraphData]);

  const resetLayout = () => {
    toast((t) => (
      <div className="flex flex-col gap-2 p-1">
        <span className="text-sm font-bold text-gray-800">Сбросить позиции графа?</span>
        <span className="text-[11px] text-gray-500 mb-1">Узлы выстроятся заново. Отменит ручные настройки.</span>
        <div className="flex gap-2">
          <button onClick={() => {
            toast.dismiss(t.id);
            if (projectId) localStorage.removeItem(POSITIONS_KEY_PREFIX + projectId);
            if (networkRef.current) {
              networkRef.current.destroy();
              networkRef.current = null;
            }
            loadGraphData();
            toast.success('Граф перестроен', { icon: '✨' });
          }} className="flex-1 bg-blue-600 text-white py-1.5 rounded-lg text-xs font-bold hover:bg-blue-700">Сбросить</button>
          <button onClick={() => toast.dismiss(t.id)} className="flex-1 bg-gray-100 text-gray-600 py-1.5 rounded-lg text-xs font-bold hover:bg-gray-200">Отмена</button>
        </div>
      </div>
    ), { duration: 5000 });
  };

  return (
    <div className="h-full w-full bg-gray-50 p-6 overflow-hidden">
      <div className="w-full h-full flex flex-col bg-white rounded-2xl shadow-xl border overflow-hidden transition-all">
        <div className="h-[72px] min-h-[72px] px-8 border-b flex items-center gap-3 bg-white z-10">
          <div className="flex items-center gap-2 flex-1">
            <Search size={14} className="text-gray-400" />
            <input placeholder="Поиск задач..." value={searchQuery} onChange={e => setSearchQuery(e.target.value)}
              className="flex-1 bg-transparent text-sm outline-none text-gray-700" />
          </div>
          <Filter size={14} className="text-gray-400" />
          <select value={filterStatus} onChange={e => setFilterStatus(e.target.value)}
            className="text-xs border rounded-lg p-1.5 outline-none bg-white text-gray-600">
            <option value="all">Все статусы</option>
            <option value="todo">Бэклог</option>
            <option value="in_progress">В работе</option>
            <option value="done">Готово</option>
          </select>
          <select value={filterAssignee} onChange={e => setFilterAssignee(e.target.value)}
            className="text-xs border rounded-lg p-1.5 outline-none bg-white text-gray-600">
            <option value="all">Все исполнители</option>
            <option value="assigned">Назначенные</option>
            <option value="unassigned">Не назначены</option>
          </select>
          <div className="flex items-center gap-2 ml-2">
            <button onClick={exportPNG}
              className="text-xs font-bold bg-green-600 text-white px-4 py-2 rounded-xl hover:bg-green-700 transition-all shadow flex items-center gap-1.5">
              <Download size={12} /> СКАЧАТЬ PNG
            </button>
            <button onClick={loadGraphData}
              className="text-xs font-bold bg-blue-600 text-white px-5 py-2 rounded-xl hover:bg-blue-700 transition-all shadow flex items-center gap-1.5">
              <RefreshCw size={12} /> ОБНОВИТЬ ДАННЫЕ
            </button>
            <button onClick={resetLayout}
              className="text-xs font-bold bg-gray-500 text-white px-3 py-2 rounded-xl hover:bg-gray-600 transition-all shadow flex items-center gap-1">
              <RotateCcw size={12} /> СБРОСИТЬ ПОЗИЦИИ
            </button>
          </div>
        </div>
        <div ref={containerRef} className="flex-1 w-full bg-white" />
        <div className="h-7 px-8 border-t flex items-center gap-4 bg-gray-50 text-[10px] text-gray-400">
          <span>🟢 Готово</span>
          <span>🔵 В работе</span>
          <span>🔴 Критический путь</span>
          <span className="ml-auto">Клик по 2 задачам = создать связь | Двойной клик по связи = удалить | Перетаскивание = переместить</span>
        </div>
      </div>
    </div>
  );
}

TaskGraph.propTypes = { projectId: PropTypes.oneOfType([PropTypes.number, PropTypes.string]) };