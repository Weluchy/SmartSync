import { useEffect, useRef, useCallback } from 'react'; // Добавили useCallback
import PropTypes from 'prop-types'; // Импорт PropTypes
import { Network } from 'vis-network/standalone';
import { api } from '../../api/client';

export default function TaskGraph({ projectId }) {
  const containerRef = useRef(null);
  const networkRef = useRef(null);

  const loadGraphData = useCallback(async () => {
    try {
      const data = await api.get(`/projects/${projectId}/graph`);
      
      const nodes = data.tasks.map(t => ({
        id: t.id,
        label: t.title,
        color: {
          background: t.status === 'done' ? '#dcfce7' : '#ffffff',
          border: t.status === 'in_progress' ? '#3b82f6' : '#e5e7eb',
        },
        font: { size: 14, face: 'Inter' },
        shape: 'box',
        margin: 10
      }));

      const edges = data.dependencies.map(d => ({
        from: d.parent_id,
        to: d.child_id,
        arrows: 'to',
        color: { color: '#94a3b8' }
      }));

      const graphData = { nodes, edges };
      
      const options = {
        physics: {
          enabled: true,
          barnesHut: { gravConstant: -2000, centralGravity: 0.3, springLength: 150 }
        },
        interaction: { hover: true, navigationButtons: true }
      };

      if (networkRef.current) {
        networkRef.current.destroy();
      }
      
      networkRef.current = new Network(containerRef.current, graphData, options);
    } catch (err) {
      console.error('Ошибка загрузки графа:', err);
    }
  }, [projectId]); // Зависит от projectId

  useEffect(() => {
    if (projectId && containerRef.current) {
      loadGraphData();
    }
  }, [projectId, loadGraphData]); // Добавили зависимости

  return (
    <div className="h-full w-full bg-gray-50 flex flex-col">
      <div className="p-4 bg-white border-b flex justify-between items-center">
        <h3 className="font-bold text-gray-700">Граф зависимостей (PERT)</h3>
        <button 
          onClick={loadGraphData}
          className="text-xs bg-gray-100 hover:bg-gray-200 px-3 py-1 rounded transition-colors"
        >
          Обновить граф
        </button>
      </div>
      <div ref={containerRef} className="flex-1" />
    </div>
  );
}

// Добавили валидацию пропсов
TaskGraph.propTypes = {
  projectId: PropTypes.oneOfType([PropTypes.string, PropTypes.number])
};