import { useEffect, useRef, useCallback } from 'react';
import PropTypes from 'prop-types';
import { Network } from 'vis-network/standalone';
import { api } from '../../api/client';

const POSITIONS_KEY_PREFIX = 'smartsync_graph_positions_';

export default function TaskGraph({ projectId }) {
  const containerRef = useRef(null);
  const networkRef = useRef(null);

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

  // Сохранение позиций узлов в localStorage
  const savePositions = useCallback(() => {
    if (!networkRef.current || !projectId) return;
    const positions = networkRef.current.getPositions();
    try {
      localStorage.setItem(POSITIONS_KEY_PREFIX + projectId, JSON.stringify(positions));
    } catch (e) {
      // localStorage может быть недоступен
    }
  }, [projectId]);

  // Восстановление позиций из localStorage
  const loadSavedPositions = useCallback(() => {
    if (!projectId) return null;
    try {
      const saved = localStorage.getItem(POSITIONS_KEY_PREFIX + projectId);
      return saved ? JSON.parse(saved) : null;
    } catch (e) {
      return null;
    }
  }, [projectId]);

  const loadGraphData = useCallback(async () => {
    if (!projectId || !containerRef.current) return;
    try {
      const data = await api.get(`/projects/${projectId}/graph`);
      const tasks = data?.tasks || data?.nodes || []; 
      const dependencies = data?.dependencies || data?.edges || [];

      const durations = {};
      tasks.forEach(t => {
        durations[t.id] = t.status === 'done' ? 0 : ((t.opt + 4 * t.real + t.pess) / 6);
      });

      const maxScore = tasks.length > 0 ? Math.max(...tasks.map(t => t.priority_score !== undefined ? t.priority_score : (t.PriorityScore || 0))) : 0;
      const criticalSet = new Set();

      tasks.forEach(t => {
        const score = t.priority_score !== undefined ? t.priority_score : (t.PriorityScore || 0);
        if (score >= maxScore - 0.1 && maxScore > 0) criticalSet.add(t.id);
      });

      let changed = true;
      while(changed) {
        changed = false;
        dependencies.forEach(d => {
          if (criticalSet.has(d.task_id) && !criticalSet.has(d.depends_on_id)) {
            const parentScore = tasks.find(t => t.id === d.depends_on_id)?.priority_score || tasks.find(t => t.id === d.depends_on_id)?.PriorityScore || 0;
            const childDur = durations[d.task_id] || 0;
            const childScore = tasks.find(t => t.id === d.task_id)?.priority_score || tasks.find(t => t.id === d.task_id)?.PriorityScore || 0;
            if (Math.abs(parentScore + childDur - childScore) < 0.1) {
              criticalSet.add(d.depends_on_id);
              changed = true;
            }
          }
        });
      }

      // Восстанавливаем сохранённые позиции узлов
      const savedPositions = loadSavedPositions();
      
      const nodes = tasks.map(t => {
        const score = t.priority_score !== undefined ? t.priority_score : t.PriorityScore || 0;
        const isDone = t.status === 'done';
        const isInProgress = t.status === 'in_progress';
        const isCritical = criticalSet.has(t.id) && !isDone; 

        let bg = '#ffffff', border = '#e5e7eb', text = '#1f2937'; 
        if (isDone) { bg = '#dcfce7'; border = '#16a34a'; text = '#166534'; }
        else if (isCritical) { bg = '#fee2e2'; border = '#ef4444'; text = '#991b1b'; }
        else if (isInProgress) { bg = '#dbeafe'; border = '#2563eb'; text = '#1e3a8a'; }

        const node = {
          id: t.id,
          label: `[${t.id}] ${t.title}\nИсп: ${t.assignee_name || 'Нет'}\n${score.toFixed(1)}ч`,
          title: `Автор: ${t.created_by_name || 'Неизвестно'}\nИсполнитель: ${t.assignee_name || 'Не назначен'}\nСтатус: ${t.status}\n\nОписание:\n${t.description || 'Нет описания'}`,
          color: { background: bg, border: border },
          font: { size: 13, multi: true, color: text },
          shape: 'box',
          margin: 12,
          borderWidth: 2,
          shadow: isCritical ? { enabled: true, color: 'rgba(239, 68, 68, 0.2)', size: 10 } : false
        };

        // Если есть сохранённые позиции — применяем их
        if (savedPositions && savedPositions[t.id]) {
          node.x = savedPositions[t.id].x;
          node.y = savedPositions[t.id].y;
          node.fixed = false; // можно будет двигать
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

      if (networkRef.current) networkRef.current.destroy();

      // Создаём сеть с отключённой физикой, без hierarchical
      networkRef.current = new Network(containerRef.current, { nodes, edges }, {
        physics: {
          enabled: !savedPositions, // включаем физику только если нет сохранённых позиций
          solver: 'forceAtlas2Based',
          forceAtlas2Based: {
            gravitationalConstant: -40,
            centralGravity: 0.005,
            springLength: 200,
            springConstant: 0.02,
            damping: 0.4
          },
          stabilization: {
            iterations: 100,
            updateInterval: 25,
            onlyDynamicEdges: false
          }
        },
        layout: {
          hierarchical: { enabled: false } // отключаем иерархию!
        },
        interaction: {
          multiselect: true,
          hover: true,
          dragNodes: true // разрешаем перетаскивание
        },
        nodes: {
          physics: true
        }
      });

      // После стабилизации физики отключаем её, чтобы узлы застыли
      if (!savedPositions) {
        networkRef.current.once('stabilizationIterationsDone', () => {
          networkRef.current.setOptions({ physics: { enabled: false } });
          // Сохраняем позиции после авто-расположения
          setTimeout(savePositions, 500);
        });
      }

      // Сохраняем позиции при перемещении узла пользователем
      networkRef.current.on('dragEnd', savePositions);

      // При клике на пустое место тоже сохраняем (пользователь мог подвинуть)
      networkRef.current.on('release', savePositions);

      networkRef.current.once("afterDrawing", () => networkRef.current.fit());

      networkRef.current.on("click", (params) => {
        if (params.nodes.length === 2) {
          const [from, to] = params.nodes;
          const currentEdges = networkRef.current.body.data.edges.get();
          if (wouldCreateCycle(from, to, currentEdges)) {
            alert("Ошибка: Это приведет к циклу!");
            networkRef.current.unselectAll();
            return;
          }
          api.post(`/tasks/${to}/dependencies`, { depends_on_id: from })
             .then(() => { networkRef.current.unselectAll(); setTimeout(loadGraphData, 300); })
             .catch(err => { alert(err.message); networkRef.current.unselectAll(); });
        }
      });

      networkRef.current.on("doubleClick", (params) => {
        if (params.edges.length > 0) {
          const edgeId = params.edges[0];
          if (typeof edgeId === 'string' && edgeId.includes('-')) {
            const [from, to] = edgeId.split('-');
            if (confirm(`Удалить связь?`)) {
              api.delete(`/tasks/${to}/dependencies/${from}`)
                 .then(() => { networkRef.current.unselectAll(); setTimeout(loadGraphData, 300); })
                 .catch(err => alert(err.message));
            }
          }
        }
      });
    } catch (err) { console.error(err); }
  }, [projectId, loadSavedPositions, savePositions]);

  useEffect(() => {
    loadGraphData();
    const interval = setInterval(() => {
      loadGraphData();
    }, 15000);

    return () => {
      clearInterval(interval);
      if (networkRef.current) networkRef.current.destroy();
    };
  }, [loadGraphData]);

  // Кнопка "Сбросить расположение"
  const resetLayout = () => {
    if (projectId) {
      localStorage.removeItem(POSITIONS_KEY_PREFIX + projectId);
    }
    if (networkRef.current) networkRef.current.destroy();
    loadGraphData();
  };

  return (
    <div className="h-full w-full bg-gray-50 p-6 overflow-hidden">
      <div className="w-full h-full flex flex-col bg-white rounded-2xl shadow-xl border overflow-hidden transition-all">
        <div className="h-[72px] min-h-[72px] px-8 border-b flex justify-between items-center bg-white z-10">
          <div className="flex items-center gap-4">
            <div>
              <h3 className="font-bold text-gray-700 uppercase text-xs tracking-widest">Сетевой график</h3>
              <p className="text-[10px] text-gray-400">Зеленый = Готово | Красный = Критический путь</p>
            </div>
            <button onClick={resetLayout} className="text-[10px] font-bold bg-gray-500 text-white px-3 py-1.5 rounded-xl hover:bg-gray-600 transition-all shadow">
              СБРОСИТЬ ПОЗИЦИИ
            </button>
          </div>
          <button onClick={loadGraphData} className="text-xs font-bold bg-blue-600 text-white px-6 py-2.5 rounded-xl hover:bg-blue-700 transition-all shadow-md">
            ОБНОВИТЬ ГРАФ
          </button>
        </div>
        <div ref={containerRef} className="flex-1 w-full bg-white" />
      </div>
    </div>
  );
}

TaskGraph.propTypes = { projectId: PropTypes.oneOfType([PropTypes.number, PropTypes.string]) };