import { useEffect, useRef, useCallback } from 'react';
import PropTypes from 'prop-types';
import { Network } from 'vis-network/standalone';
import { api } from '../../api/client';

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

      const nodes = tasks.map(t => {
        const score = t.priority_score !== undefined ? t.priority_score : t.PriorityScore || 0;
        const isDone = t.status === 'done';
        const isInProgress = t.status === 'in_progress';
        const isCritical = criticalSet.has(t.id) && !isDone; 

        let bg = '#ffffff', border = '#e5e7eb', text = '#1f2937'; 
        if (isDone) { bg = '#dcfce7'; border = '#16a34a'; text = '#166534'; }
        else if (isCritical) { bg = '#fee2e2'; border = '#ef4444'; text = '#991b1b'; }
        else if (isInProgress) { bg = '#dbeafe'; border = '#2563eb'; text = '#1e3a8a'; }

        return {
          id: t.id,
          label: `[${t.id}] ${t.title}\n${score.toFixed(1)}ч`,
          color: { background: bg, border: border },
          font: { size: 13, multi: true, color: text },
          shape: 'box',
          margin: 12,
          borderWidth: 2,
          shadow: isCritical ? { enabled: true, color: 'rgba(239, 68, 68, 0.2)', size: 10 } : false
        };
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
      networkRef.current = new Network(containerRef.current, { nodes, edges }, {
        physics: { enabled: false },
        layout: { hierarchical: { enabled: true, direction: 'LR', sortMethod: 'directed', levelSeparation: 200 } },
        interaction: { multiselect: true, hover: true }
      });

      // Авто-масштаб после отрисовки
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
  }, [projectId]);

  useEffect(() => {
  loadGraphData();

  // ПРИКАЗ: Граф тяжелее доски, поэтому опрашиваем чуть реже — раз в 15 секунд
  const interval = setInterval(() => {
    loadGraphData();
  }, 15000);

  return () => {
    clearInterval(interval);
    if (networkRef.current) networkRef.current.destroy();
  };
}, [loadGraphData]);

  return (
    <div className="h-full w-full bg-gray-50 p-6 overflow-hidden">
      {/* ПРИКАЗ: Окно во всю ширину */}
      <div className="w-full h-full flex flex-col bg-white rounded-2xl shadow-xl border overflow-hidden transition-all">
        <div className="h-[72px] min-h-[72px] px-8 border-b flex justify-between items-center bg-white z-10">
          <div>
            <h3 className="font-bold text-gray-700 uppercase text-xs tracking-widest">Сетевой график</h3>
            <p className="text-[10px] text-gray-400">Зеленый = Готово | Красный = Критический путь</p>
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