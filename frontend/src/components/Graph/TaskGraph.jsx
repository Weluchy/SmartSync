import { useEffect, useRef, useCallback } from 'react';
import PropTypes from 'prop-types';
import { Network } from 'vis-network/standalone';
import { api } from '../../api/client';

export default function TaskGraph({ projectId }) {
  const containerRef = useRef(null);
  const networkRef = useRef(null);

  const loadGraphData = useCallback(async () => {
    if (!projectId || !containerRef.current) return;
    
    try {
      const data = await api.get(`/projects/${projectId}/graph`);
      
      const nodes = data.tasks.map(t => ({
        id: t.id,
        label: t.title,
        color: {
          background: t.status === 'done' ? '#dcfce7' : '#ffffff',
          border: t.status === 'in_progress' ? '#3b82f6' : '#e5e7eb',
        },
        font: { size: 14, color: '#1f2937' },
        shape: 'box',
        margin: 10,
        borderWidth: 2
      }));

      const edges = data.dependencies.map(d => ({
        from: d.parent_id,
        to: d.child_id,
        arrows: 'to',
        color: { color: '#94a3b8' },
        smooth: { type: 'cubicBezier' }
      }));

      const options = {
        physics: {
          enabled: true,
          barnesHut: { gravConstant: -2000, centralGravity: 0.3, springLength: 150 },
          stabilization: { iterations: 150 }
        },
        interaction: { hover: true, navigationButtons: true }
      };

      if (networkRef.current) networkRef.current.destroy();
      networkRef.current = new Network(containerRef.current, { nodes, edges }, options);
    } catch (err) {
      console.error('Ошибка загрузки графа:', err);
    }
  }, [projectId]);

  useEffect(() => {
    loadGraphData();
  }, [loadGraphData]);

  return (
    <div className="h-full w-full bg-gray-50 flex flex-col">
      <div className="p-4 bg-white border-b flex justify-between items-center shadow-sm z-10">
        <h3 className="font-bold text-gray-700">Сетевой график (PERT)</h3>
        <button 
          onClick={loadGraphData}
          className="text-xs font-bold bg-blue-50 text-blue-600 hover:bg-blue-100 px-4 py-2 rounded-lg transition-colors"
        >
          Обновить связи
        </button>
      </div>
      <div ref={containerRef} className="flex-1 cursor-grab active:cursor-grabbing" />
    </div>
  );
}

TaskGraph.propTypes = {
  projectId: PropTypes.oneOfType([PropTypes.number, PropTypes.string])
};