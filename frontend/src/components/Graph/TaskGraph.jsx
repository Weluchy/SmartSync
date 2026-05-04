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
      const tasks = data?.tasks || [];
      const dependencies = data?.dependencies || [];

      const nodes = tasks.map(t => ({
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

      const edges = dependencies.map(d => ({
        from: d.depends_on_id, 
        to: d.task_id,
        arrows: 'to',
        color: { color: '#94a3b8' },
        smooth: { type: 'cubicBezier' }
      }));

      const options = {
        physics: {
          enabled: false // Выключаем физику, чтобы граф не улетал
        },
        layout: {
          hierarchical: {
            enabled: true,
            direction: 'LR', // Слева направо
            sortMethod: 'directed',
            nodeSpacing: 150,
            levelSeparation: 250,
          }
        },
        interaction: { hover: true, navigationButtons: true, dragNodes: true }
      };

      if (networkRef.current) {
        networkRef.current.destroy();
      }
      
      networkRef.current = new Network(containerRef.current, { nodes, edges }, options);
    } catch (err) {
      console.error('Ошибка загрузки графа:', err);
    }
  }, [projectId]);

  useEffect(() => {
    loadGraphData();
    
    return () => {
      if (networkRef.current) {
        networkRef.current.destroy();
      }
    };
  }, [loadGraphData]);

  return (
    <div className="flex flex-col h-full w-full bg-gray-50 overflow-hidden">
      <div className="p-4 bg-white border-b flex justify-between items-center shadow-sm z-10">
        <h3 className="font-bold text-gray-700">Сетевой график задач</h3>
        <button 
          onClick={loadGraphData}
          className="text-xs font-bold bg-blue-50 text-blue-600 hover:bg-blue-100 px-4 py-2 rounded-lg transition-colors"
        >
          Обновить связи
        </button>
      </div>
      <div 
        ref={containerRef} 
        className="flex-1 w-full" 
        style={{ minHeight: '600px' }} 
      />
    </div>
  );
}

TaskGraph.propTypes = {
  projectId: PropTypes.oneOfType([PropTypes.number, PropTypes.string])
};